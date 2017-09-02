package main

import (
	"bytes"
	"database/sql"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const dbFilename = "shiritoriwords.db"
const jmdictFileName = "jmdict.xml"
const kanjidictFileName = "kanjidic2.xml"
const maxCharScore = 15 // Max grade of 10 + Max JLPT of 5
const maxLimit = 4      // Arbitrary limit
const scoreFactor = maxCharScore / maxLimit
const maxJLPT = 6

type kmap map[string]int

var db *sql.DB
var insertcount = 0
var failcount = 0
var kanjiExp = regexp.MustCompile(`\p{Han}+`)
var endsInNExp = regexp.MustCompile(`(ん|ン)$`)

func main() {
	log.Printf("Loading kanji dictionary...")
	dict, err := getKanjiDict()
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	log.Printf("Loading kanji score map...")
	kscores, err := getKanjiScoreMap()
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	log.Printf("Creating kanji database...")
	err = createKanjiDb(dict, kscores)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
}

func createKanjiDb(dict *jmdict, kscores kmap) error {
	var err error
	err = os.Remove(dbFilename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	db, err = sql.Open("sqlite3", dbFilename)
	if err != nil {
		return err
	}
	err = createTable("words (kanji TEXT PRIMARY KEY, kana TEXT, score INT)")
	if err != nil {
		return err
	}
	// Prepare the statement and use a transaction for massive speed increase.
	stmt, err := db.Prepare("INSERT INTO words (kanji, kana, score) VALUES (:KJ, :KN, :SC)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	execDb("PRAGMA synchronous = OFF")
	execDb("PRAGMA journal_mode = MEMORY")
	execDb("BEGIN")
	defer execDb("COMMIT")
	for _, e := range dict.Entry {
		if isNoun(e) {
			saveWord(stmt, e, kscores)
		}
	}
	log.Printf("Inserted %d record(s)", insertcount)
	log.Printf("Merged %d record(s)", failcount)
	return nil
}

func isNoun(e entry) bool {
	for _, s := range e.Sense {
		for _, p := range s.Pos {
			if p == "n" {
				return true
			}
		}
	}
	return false
}

func saveWord(stmt *sql.Stmt, e entry, kscores kmap) {
	var score int
	kana, endsInN := getKana(e)
	// Get all of the entry kanji variants
	for _, kanji := range getKanjis(e) {
		if endsInN {
			// Automatic zero for ending in 'ん'
			score = 0
		} else {
			score = getKanjiWordScore(kanji, kscores)
		}
		_, err := stmt.Exec(sql.Named("KJ", &kanji), sql.Named("KN", &kana), sql.Named("SC", &score))
		if err != nil {
			mergeRecords(kanji, kana, score)
			failcount++
			// return err
		} else {
			insertcount++
		}
	}
}

func mergeRecords(kanji string, kana string, score int) {
	var existingKana string
	var existingScore int
	found := queryDb(fmt.Sprintf("SELECT kana, score FROM words WHERE kanji = '%s'", kanji),
		&existingKana, &existingScore)
	if !found {
		log.Printf("DBERROR: Could not find record for %s", kanji)
		return
	}
	// Merge kana and the word score.
	newKana := mergeStrings(existingKana, kana)
	newScore := mergeScore(existingScore, score)
	err := execDb("UPDATE words SET kana = ?, score = ? WHERE kanji = ?", &newKana, &newScore, &kanji)
	if err != nil {
		log.Printf("DBERROR updating %s", kanji)
	}
}

func mergeStrings(first string, second string) string {
	mergedmap := make(map[string]bool)
	addStrings(mergedmap, first)
	addStrings(mergedmap, second)
	mergedString := ""
	for str := range mergedmap {
		if len(mergedString) > 0 {
			mergedString += ","
		}
		mergedString += str
	}
	return mergedString
}

func addStrings(dest map[string]bool, strs string) {
	for _, str := range strings.Split(strs, ",") {
		dest[str] = true
	}
}

func mergeScore(first int, second int) int {
	if first > second {
		return first
	}
	return second
}

func getKanjiWordScore(kanji string, kscores kmap) int {
	// Words entirely of hiragana or katakana are worth 1 point.
	score := 1
	if kanjiExp.MatchString(kanji) {
		for _, k := range kanji {
			kscore := kscores[string(k)]
			if kscore > score {
				// The word score is equal to the highest kanji score in the word.
				score = kscore
			}
		}
	}
	return score
}

/*
The kanji element, or in its absence, the reading element, is
the defining component of each entry.
The overwhelming majority of entries will have a single kanji
element associated with a word in Japanese. Where there are
multiple kanji elements within an entry, they will be orthographical
variants of the same word, either using variations in okurigana, or
alternative and equivalent kanji. Common "mis-spellings" may be
included, provided they are associated with appropriate information
fields. Synonyms are not included; they may be indicated in the
cross-reference field associated with the sense element.
*/
func getKanjis(e entry) []string {
	kanjis := make([]string, 0)
	for _, k := range e.Kele {
		if len(k.Keb) > 0 {
			kanjis = append(kanjis, k.Keb)
		}
	}
	return kanjis
}

func getKana(e entry) (string, bool) {
	kanas := ""
	endsInN := true
	for _, k := range e.Rele {
		if !endsInNExp.MatchString(k.Reb) {
			// If at least one variation does not end in 'ん', then it's valid.
			endsInN = false
		}
		if len(kanas) > 0 {
			kanas += ","
		}
		kanas += k.Reb
	}
	return kanas, endsInN
}

func getKanjiDict() (*jmdict, error) {
	data, err := loadXMLFile(jmdictFileName)
	if err != nil {
		return nil, err
	}
	d := xml.NewDecoder(bytes.NewReader(data))
	// Map the entities to standard XML, or else a parsing error occurs.
	d.Entity = jmentitymap
	dict := jmdict{}
	err = d.Decode(&dict)
	if err != nil {
		return nil, err
	}
	return &dict, nil
}

func getKanjiScoreMap() (kmap, error) {
	data, err := loadXMLFile(kanjidictFileName)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}
	kanji := kanjidic{}
	err = xml.Unmarshal(data, &kanji)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}
	kscores := make(kmap)
	for _, ch := range kanji.Character {
		kscores[ch.Literal] = getCharacterScore(ch)
	}
	return kscores, nil
}

func getCharacterScore(ch character) int {
	score := ch.Misc.Grade
	if ch.Misc.JLPT > 0 {
		// JLPT is in reverse order. Higher level is lower number.
		score += maxJLPT - ch.Misc.JLPT
	}
	// Kanji Score: RawScore / ScoreFactor
	return score / scoreFactor
}

func loadXMLFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func createTable(tableDef string) error {
	return execDb(fmt.Sprintf("CREATE TABLE %s", tableDef))
}

func createIndex(indexDef string) error {
	return execDb(fmt.Sprintf("CREATE INDEX %s", indexDef))
}

func execDb(stmt string, args ...interface{}) error {
	statement, err := db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(args...)
	return err
}

func queryDb(stmt string, args ...interface{}) bool {
	rows, err := db.Query(stmt)
	defer closeRows(rows)
	if err != nil {
		log.Printf("DBERROR: Querying %s: %v", stmt, err)
		return false
	}
	if rows.Next() {
		if args != nil {
			rows.Scan(args...)
		}
		return true
	}
	return false
}

func closeRows(rows *sql.Rows) {
	if nil != rows {
		rows.Close()
	}
}

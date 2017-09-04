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

const dbFilename = "torigemu.db"

const jmdictFileName = "jmdict.xml"
const kanjidictFileName = "kanjidic2.xml"
const maxPts = 15  // Max grade of 10 + Max JLPT of 5
const maxLimit = 4 // Arbitrary limit
const ptsFactor = maxPts / maxLimit
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
	log.Printf("Loading kanji points map...")
	kptsmap, err := getKanjiPtsMap()
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	log.Printf("Creating kanji database...")
	err = createKanjiDb(dict, kptsmap)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
}

func createKanjiDb(dict *jmdict, kptsmap kmap) error {
	var err error
	db, err = sql.Open("sqlite3", dbFilename)
	if err != nil {
		return err
	}
	dropTable("words")
	err = createTable("words (kanji TEXT PRIMARY KEY, kana TEXT, points INT)")
	if err != nil {
		return err
	}
	// Prepare the statement and use a transaction for massive speed increase.
	stmt, err := db.Prepare("INSERT INTO words (kanji, kana, points) VALUES (:KJ, :KN, :SC)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	// Optimize the database insertion
	execDb("BEGIN")
	for _, e := range dict.Entry {
		if isNoun(e) {
			saveWord(stmt, e, kptsmap)
		}
	}
	execDb("COMMIT")
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

func saveWord(stmt *sql.Stmt, e entry, kptsmap kmap) {
	var pts int
	kana, endsInN := getKana(e)
	// Get all of the entry kanji variants
	for _, kanji := range getKanjis(e) {
		if endsInN {
			// Automatic zero for ending in 'ん'
			pts = 0
		} else {
			pts = getKanjiWordPts(kanji, kptsmap)
		}
		_, err := stmt.Exec(sql.Named("KJ", &kanji), sql.Named("KN", &kana), sql.Named("SC", &pts))
		if err != nil {
			mergeRecords(kanji, kana, pts)
			failcount++
			// return err
		} else {
			insertcount++
		}
	}
}

func mergeRecords(kanji string, kana string, pts int) {
	var existingKana string
	var existingPts int
	found := queryDb(fmt.Sprintf("SELECT kana, points FROM words WHERE kanji = '%s'", kanji),
		&existingKana, &existingPts)
	if !found {
		log.Printf("DBERROR: Could not find record for %s", kanji)
		return
	}
	// Merge kana and the word pts.
	newKana := mergeStrings(existingKana, kana)
	newPts := mergePts(existingPts, pts)
	err := execDb("UPDATE words SET kana = ?, points = ? WHERE kanji = ?", &newKana, &newPts, &kanji)
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

func mergePts(first int, second int) int {
	if first > second {
		return first
	}
	return second
}

func getKanjiWordPts(kanji string, kptsmap kmap) int {
	// Words entirely of hiragana or katakana are worth 1 point.
	pts := 1
	if kanjiExp.MatchString(kanji) {
		for _, k := range kanji {
			kpts := kptsmap[string(k)]
			if kpts > pts {
				// The word pts is equal to the highest kanji pts in the word.
				pts = kpts
			}
		}
	}
	return pts
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

func getKanjiPtsMap() (kmap, error) {
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
	kptsmap := make(kmap)
	for _, ch := range kanji.Character {
		kptsmap[ch.Literal] = getCharacterPts(ch)
	}
	return kptsmap, nil
}

func getCharacterPts(ch character) int {
	pts := ch.Misc.Grade
	if ch.Misc.JLPT > 0 {
		// JLPT is in reverse order. Higher level is lower number.
		pts += maxJLPT - ch.Misc.JLPT
	}
	// Kanji Pts: RawPts / PtsFactor
	return pts / ptsFactor
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

func dropTable(tableDef string) {
	execDb(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableDef))
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

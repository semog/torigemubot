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
const maxPts = 15.0   // Max grade of 10 + Max JLPT of 5
const maxLimit = 4.0  // Arbitrary limit
const ptsFactor = 3.0 // maxPts / maxLimit
const maxJLPT = 6

type kmap map[string]int

var db *sql.DB
var insertStmt *sql.Stmt
var insertcount = 0
var failcount = 0
var kanjiExp = regexp.MustCompile(`\p{Han}+`)
var endsInNExp = regexp.MustCompile(`(ん|ン)$`)
var katakanaExp = regexp.MustCompile(`(\p{Katakana}|ー)`)

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
	dropTable("kanjipoints")
	err = createTable("words (seq TEXT, kanji TEXT PRIMARY KEY, kana TEXT, points INT)")
	if err != nil {
		return err
	}
	err = createTable("kanjipoints (kanji TEXT PRIMARY KEY, points INT)")
	if err != nil {
		return err
	}
	insertKanjiPoints(kptsmap)
	insertWords(dict, kptsmap)
	return nil
}

func insertKanjiPoints(kptsmap kmap) error {
	var err error
	// Prepare the statement and use a transaction for massive speed increase.
	insertStmt, err = db.Prepare("INSERT INTO kanjipoints (kanji, points) VALUES (:KJ, :SC)")
	if err != nil {
		return err
	}
	defer insertStmt.Close()
	// Optimize the database insertion
	execDb("BEGIN")
	for k, p := range kptsmap {
		_, err = insertStmt.Exec(sql.Named("KJ", &k), sql.Named("SC", &p))
		if err != nil {
			execDb("ROLLBACK")
			return err
		}
	}
	execDb("COMMIT")
	return nil
}

func insertWords(dict *jmdict, kptsmap kmap) error {
	var err error
	// Prepare the statement and use a transaction for massive speed increase.
	insertStmt, err = db.Prepare("INSERT INTO words (seq, kanji, kana, points) VALUES (:SQ, :KJ, :KN, :SC)")
	if err != nil {
		return err
	}
	defer insertStmt.Close()
	// Optimize the database insertion
	execDb("BEGIN")
	for _, e := range dict.Entry {
		if isNoun(e) {
			err = saveWord(e, kptsmap)
			if err != nil {
				log.Printf("Error saving seq: %d", e.Seq)
				execDb("ROLLBACK")
				return err
			}
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
			if p == "n" ||
				p == "n-adv" ||
				p == "n-suf" ||
				p == "n-pref" ||
				p == "n-t" ||
				p == "pn" ||
				p == "n-pr" {
				return true
			}
		}
	}
	return false
}

func saveWord(e entry, kptsmap kmap) error {
	var err error
	kana, endsInN := getKana(e)
	kanjis := getKanjis(e)
	seq := fmt.Sprintf("%d", e.Seq)
	if len(kanjis) > 0 {
		// Get all of the entry kanji variants
		hiragana := convertToHiragana(kana)
		for _, kanji := range kanjis {
			err = saveKanji(seq, kanji, hiragana, endsInN, kptsmap)
			if err != nil {
				return err
			}
		}
	} else {
		// Only kana for this entry.
		for _, kn := range strings.Split(kana, ",") {
			hiragana := convertToHiragana(kn)
			err = saveKanji(seq, kn, hiragana, endsInN, kptsmap)
			if err != nil {
				return err
			}
		}
	}
	// Handle nokanji variations for this entry.
	nokanjis := getNoKanjis(e)
	if len(nokanjis) > 0 {
		for _, nkn := range nokanjis {
			err = saveKanji(seq, nkn, convertToHiragana(nkn), endsInNExp.MatchString(nkn), kptsmap)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func convertToHiragana(kana string) string {
	prevKana := ""
	hiragana := ""
	for _, kn := range []rune(kana) {
		thisKana := string(kn)
		if katakanaExp.MatchString(thisKana) {
			if thisKana == "ー" {
				// Check to see what vowel the preceding character ends on.
				// Replace the 'ー' with that hiragana vowel.
				hiragana += kanaVowelEndingMap[prevKana]
			} else {
				hiragana += kanaMap[thisKana]
			}
		} else {
			// No conversion necessary.
			hiragana += thisKana
		}
		prevKana = thisKana
	}
	return hiragana
}

func saveKanji(seq string, kanji string, kana string, endsInN bool, kptsmap kmap) error {
	var pts int
	if endsInN {
		// Automatic zero for ending in 'ん'
		pts = 0
	} else {
		pts = getKanjiWordPts(kanji, kptsmap)
	}
	_, err := insertStmt.Exec(sql.Named("SQ", seq), sql.Named("KJ", &kanji), sql.Named("KN", &kana), sql.Named("SC", &pts))
	if err != nil {
		err = mergeRecords(seq, kanji, kana, pts)
		failcount++
		// return err
	} else {
		insertcount++
	}
	return err
}

func mergeRecords(seq string, kanji string, kana string, pts int) error {
	var existingKana, existingSeq string
	var existingPts int
	found := queryDb(fmt.Sprintf("SELECT seq, kana, points FROM words WHERE kanji = '%s'", kanji),
		&existingSeq, &existingKana, &existingPts)
	if !found {
		return fmt.Errorf("DBERROR: Could not find record for %s", kanji)
	}
	// Merge kana and the word pts.
	newSeq := mergeStrings(existingSeq, seq)
	newKana := mergeStrings(existingKana, kana)
	newPts := mergePts(existingPts, pts)
	err := execDb("UPDATE words SET seq = ?, kana = ?, points = ? WHERE kanji = ?", &newSeq, &newKana, &newPts, &kanji)
	if err != nil {
		log.Printf("DBERROR updating %s", kanji)
		return err
	}
	return nil
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

/*
This element, which will usually have a null value, indicates
that the reb, while associated with the keb, cannot be regarded
as a true reading of the kanji. It is typically used for words
such as foreign place names, gairaigo which can be in kanji or
katakana, etc.
*/
func getNoKanjis(e entry) []string {
	kanjis := make([]string, 0)
	for _, k := range e.Rele {
		if k.NoKanji != nil {
			if len(k.Reb) > 0 {
				kanjis = append(kanjis, k.Reb)
			}
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
	pts := float32(ch.Misc.Grade)
	if pts < 1.0 {
		pts = 1.0
	}
	jlpt := ch.Misc.JLPT
	if jlpt < 1 {
		jlpt = 5
	}
	// JLPT is in reverse order. Higher level is lower number.
	pts += maxJLPT - float32(jlpt)
	// Kanji Pts: RawPts / PtsFactor
	return int((pts / ptsFactor) + 0.5)
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

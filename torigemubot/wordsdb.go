package main

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const wordsTablename = "words"
const kanaExp = `(\p{Katakana}|\p{Hiragana}|ー)([ゃャょョゅュ])?`
const seiExp = "なりくせ|せい|さが|しょう"

var wordsdb *sql.DB
var endKanaExp = regexp.MustCompile(fmt.Sprintf("%s$", kanaExp))
var endsInKanjiSeiExp = regexp.MustCompile("性$")
var endsInKanaSeiExp = regexp.MustCompile(fmt.Sprintf("(%s)$", seiExp))
var beginKanaExp = regexp.MustCompile(fmt.Sprintf("^%s", kanaExp))

func getWordPts(theWord string, lastEntry *wordEntry) int {
	// If points are zero or not found, then return zero. Probably ends in 'n', or not a noun.
	kana, pts := lookupKana(theWord)
	if lastEntry != nil {
		// Get kana of current word.
		lastEntryKana, _ := lookupKana(lastEntry.word)
		// If first kana of new word does not match ending kana of last word, then return zero.
		if !matchKana(lastEntryKana, kana) {
			return 0
		}
	}
	return pts
}

func lookupKana(theWord string) (string, int) {
	var kana string
	var pts int
	found := gamedb.Query(fmt.Sprintf("SELECT kana, points FROM %s WHERE kanji = '%s'", wordsTablename, theWord), &kana, &pts)
	if !found || pts == 0 {
		endsInSei := endsInKanjiSeiExp.MatchString(theWord)
		if !endsInSei {
			return kana, 0
		}
		// Also search for non-dictionary forms of nouns that end in 性.
		// Remove the ending 性 so we can search on the dictionary form.
		lookupWord := endsInKanjiSeiExp.ReplaceAllString(theWord, "")
		kana, pts = lookupKana(lookupWord)
		if len(kana) == 0 {
			return kana, 0
		}
		// Need to append the sei kana variations
		kana = appendKanaSei(kana)
	}
	return kana, pts
}

func appendKanaSei(kana string) string {
	newkana := ""
	for _, kn := range strings.Split(kana, ",") {
		if len(newkana) > 0 {
			newkana += ","
		}
		newkana += fmt.Sprintf("%sせい", kn)
	}
	return newkana
}

func matchKana(lastWordKana string, newWordKana string) bool {
	lastKana := strings.Split(lastWordKana, ",")
	newKana := strings.Split(newWordKana, ",")
	for _, lk := range lastKana {
		endingMatch := endKanaExp.FindStringSubmatch(lk)
		for _, nk := range newKana {
			if endAndBeginMatch(endingMatch, beginKanaExp.FindStringSubmatch(nk)) {
				return true
			}
		}
	}
	return false
}

func endAndBeginMatch(endingMatch []string, beginningMatch []string) bool {
	if len(endingMatch) != len(beginningMatch) {
		return false
	}
	for index := range endingMatch {
		if endingMatch[index] != beginningMatch[index] {
			return false
		}
	}
	return true
}

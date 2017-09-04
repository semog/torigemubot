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

var wordsdb *sql.DB
var endKanaExp = regexp.MustCompile(fmt.Sprintf("%s$", kanaExp))
var beginKanaExp = regexp.MustCompile(fmt.Sprintf("^%s", kanaExp))

func getWordPts(theWord string, lastEntry *wordEntry) int {
	var kana string
	var pts int
	// If points are zero or not found, then return zero. Probably ends in 'n', or not a noun.
	if !gamedb.Query(fmt.Sprintf("SELECT kana, points FROM %s WHERE kanji = '%s'", wordsTablename, theWord), &kana, &pts) || pts == 0 {
		return 0
	}
	if lastEntry != nil {
		var lastEntryKana string
		// Get kana of current word.
		// If first kana of new word does not match ending kana of current word, then return zero.
		if !gamedb.Query(fmt.Sprintf("SELECT kana FROM %s WHERE kanji = '%s'", wordsTablename, lastEntry.word), &lastEntryKana) ||
			!matchKana(lastEntryKana, kana) {
			pts = 0
		}
	}
	// Return word points.
	return pts
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

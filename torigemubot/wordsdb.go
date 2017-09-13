package main

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const wordsTablename = "words"
const customwordsTablename = "customwords"
const kanjipointsTablename = "kanjipoints"
const addCustomWordSavePoint = "AddCustomWord"
const removeCustomWordSavePoint = "RemoveCustomWord"
const kanaExp = `(\p{Katakana}|\p{Hiragana}|ー)([ゃャょョゅュ])?`

var wordsdb *sql.DB
var endKanaExp = regexp.MustCompile(fmt.Sprintf("%s$", kanaExp))
var beginKanaExp = regexp.MustCompile(fmt.Sprintf("^%s", kanaExp))
var endsInNExp = regexp.MustCompile(`(ん|ン)$`)

func getWordPts(chatID int64, theWord string, lastEntry *wordEntry) int {
	// If points are zero or not found, then return zero. Probably ends in 'n', or not a noun.
	kana, pts := lookupKana(chatID, theWord)
	if lastEntry != nil {
		// Get kana of current word.
		lastEntryKana, _ := lookupKana(chatID, lastEntry.word)
		// If first kana of new word does not match ending kana of last word, then return zero.
		if !matchKana(lastEntryKana, kana) {
			return 0
		}
	}
	return pts
}

func lookupKana(chatID int64, theWord string) (string, int) {
	var kana string
	var pts int
	found := gamedb.Query(fmt.Sprintf("SELECT kana, points FROM %s WHERE kanji = '%s'", wordsTablename, theWord), &kana, &pts)
	if !found || pts == 0 {
		if !gamedb.Query(fmt.Sprintf("SELECT kana, points FROM %s WHERE chatid = %d AND kanji = '%s'", customwordsTablename, chatID, theWord), &kana, &pts) {
			return kana, 0
		}
	}
	return kana, pts
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

func endsInN(kana string) bool {
	// Check for ending in ん.
	for _, k := range strings.Split(kana, ",") {
		if endsInNExp.MatchString(k) {
			return true
		}
	}
	return false
}

func calcWordPoints(kanji string) int {
	// Words entirely of hiragana or katakana are worth 1 point.
	pts := 1
	if kanjiExp.MatchString(kanji) {
		for _, k := range kanji {
			kpts := getKanjiPoints(string(k))
			if kpts > pts {
				// The word pts is equal to the highest kanji pts in the word.
				pts = kpts
			}
		}
	}
	return pts
}

func getKanjiPoints(kanjiCharacter string) int {
	var pts int
	if !gamedb.Query(fmt.Sprintf("SELECT points FROM %s WHERE kanji = '%s'", kanjipointsTablename, kanjiCharacter), &pts) {
		return 0
	}
	return pts
}

func addCustomWord(chatID int64, userid int, kanji string, kana string) bool {
	wordpts := 0
	if !endsInN(kana) {
		wordpts = calcWordPoints(kanji)
	}
	gamedb.CreateSavePoint(addCustomWordSavePoint)
	// Replace any existing custom word with the updated version of it.
	removeCustomWord(chatID, kanji)
	return gamedb.CommitSavePointOnSuccess(addCustomWordSavePoint,
		gamedb.Exec(fmt.Sprintf("INSERT INTO %s (chatid, userid, kanji, kana, points) VALUES (?, ?, ?, ?, ?)", customwordsTablename), chatID, userid, kanji, kana, wordpts) &&
			updatePlayerScore(chatID, userid, addWordPts))
}

func removeCustomWord(chatID int64, kanji string) bool {
	// Get userid that submitted the custom word.
	var userid int
	if !gamedb.Query(fmt.Sprintf("SELECT userid from %s where chatid = %d AND kanji = '%s'", customwordsTablename, chatID, kanji), &userid) {
		return false
	}
	gamedb.CreateSavePoint(removeCustomWordSavePoint)
	return gamedb.CommitSavePointOnSuccess(removeCustomWordSavePoint,
		gamedb.Exec(fmt.Sprintf("DELETE FROM %s WHERE chatid = %d AND kanji = '%s'", customwordsTablename, chatID, kanji)) &&
			updatePlayerScore(chatID, userid, -addWordPts))
}

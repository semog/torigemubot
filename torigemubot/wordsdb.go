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

func getWordPts(chatID int64, theWord string, lastEntry *wordEntry) (int, string) {
	// If points are zero or not found, then return zero. Probably ends in 'n', or not a noun.
	kana, pts := lookupKana(chatID, theWord)
	if lastEntry != nil {
		// Get kana of last word.
		lastEntryKana, _ := lookupKana(chatID, lastEntry.word)
		// If first kana of new word does not match ending kana of last word, then return zero.
		if !matchKana(lastEntryKana, kana) {
			return 0, fmt.Sprintf("初めの仮名は終わりのかなと一致しません: %s「%s」-> %s「%s」", lastEntry.word, lastEntryKana, theWord, kana)
		}
	} else if pts == 0 {
		if endsInN(kana) {
			return pts, fmt.Sprintf("言葉は'ん'が終わることが禁止されています: %s", theWord)
		}
		return pts, fmt.Sprintf("無効言葉: %s", theWord)
	}
	return pts, ""
}

func lookupKana(chatID int64, theWord string) (string, int) {
	found, kana, pts := lookupStandardKana(chatID, theWord)
	if !found || pts == 0 {
		var customkana string
		found, customkana, pts = lookupCustomKana(chatID, theWord)
		if !found {
			return kana, 0
		}
		kana = customkana
	}
	return kana, pts
}

func lookupStandardKana(chatID int64, theWord string) (bool, string, int) {
	var kana string
	var pts int
	found := gamedb.Query(fmt.Sprintf("SELECT kana, points FROM %s WHERE kanji = '%s'", wordsTablename, theWord), &kana, &pts)
	return found, kana, pts
}

func lookupCustomKana(chatID int64, theWord string) (bool, string, int) {
	var kana string
	var pts int
	found := gamedb.Query(fmt.Sprintf("SELECT kana, points FROM %s WHERE chatid = %d AND kanji = '%s'", customwordsTablename, chatID, theWord), &kana, &pts)
	return found, kana, pts
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
	// If the word ends in a combined phonic (i.e., しゃ), then the
	// next word must begin with that same combination.
	// However, if the word ends in just し, then the next word can
	// optionally start with combined phonic (i.e., しゃ) or just し.
	for index := 1; index < len(endingMatch); index++ {
		if len(endingMatch[index]) > 0 && endingMatch[index] != beginningMatch[index] {
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
	wordpts := calcWordPoints(kanji)
	gamedb.CreateSavePoint(addCustomWordSavePoint)
	// Replace any existing custom word with the updated version of it.
	return gamedb.CommitSavePointOnSuccess(addCustomWordSavePoint,
		removeCustomWord(chatID, kanji) &&
			gamedb.Exec(fmt.Sprintf("INSERT INTO %s (chatid, userid, kanji, kana, points) VALUES (?, ?, ?, ?, ?)", customwordsTablename), chatID, userid, kanji, kana, wordpts) &&
			updatePlayerScore(chatID, userid, addWordPts))
}

func removeCustomWord(chatID int64, kanji string) bool {
	// Get userid that submitted the custom word.
	var userid int
	if !gamedb.Query(fmt.Sprintf("SELECT userid from %s where chatid = %d AND kanji = '%s'", customwordsTablename, chatID, kanji), &userid) {
		// Custom word does not exist, so it has been removed.
		return true
	}
	gamedb.CreateSavePoint(removeCustomWordSavePoint)
	return gamedb.CommitSavePointOnSuccess(removeCustomWordSavePoint,
		gamedb.Exec(fmt.Sprintf("DELETE FROM %s WHERE chatid = %d AND kanji = '%s'", customwordsTablename, chatID, kanji)) &&
			updatePlayerScore(chatID, userid, -addWordPts))
}

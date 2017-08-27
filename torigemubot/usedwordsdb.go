package main

import (
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const usedwordsTableName = "usedwords"

func saveWord(entry *wordEntry) {
	// TODO: Use the timestamp seconds for wordorder.
	usedWords[entry.chatid] = append(usedWords[entry.chatid], entry)
	updatePlayerWords(entry.chatid, entry.userid, 1)
}

func alreadyUsedWord(chatID int64, theWord string) bool {
	wordCheck := strings.ToLower(theWord)
	for _, usedWord := range usedWords[chatID] {
		if wordCheck == strings.ToLower(usedWord.word) {
			return true
		}
	}
	return false
}

func getFirstEntry(chatID int64) *wordEntry {
	return usedWords[chatID][0]
}

func getLastEntry(chatID int64) *wordEntry {
	numWords := len(usedWords[chatID])
	if numWords == 0 {
		return nil
	}
	return usedWords[chatID][numWords-1]
}

func removeLastEntry(chatID int64) {
	currentWords := usedWords[chatID]
	usedWords[chatID] = currentWords[:len(currentWords)-1]
}

func getUsedWords(chatID int64) wordList {
	return usedWords[chatID]
}

func getNumWords(chatID int64) int {
	return len(usedWords[chatID])
}

func clearSavedWords(chatID int64) {
	// Clear out the word history.
	delete(usedWords, chatID)
}

package main

import (
	"database/sql"
	"fmt"
	"time"
)

const usedwordsTableName = "usedwords"

func addEntry(entry *wordEntry) {
	// Use the timestamp seconds for wordindex.
	gamedb.Exec(fmt.Sprintf("INSERT INTO %s (chatid, userid, wordindex, word, points) VALUES (?, ?, ?, ?, ?)", usedwordsTableName),
		entry.chatid, entry.userid, time.Now().Unix(), entry.word, entry.points)
	updatePlayerWords(entry.chatid, entry.userid, 1)
}

func alreadyUsedWord(chatID int64, theWord string) bool {
	return gamedb.SingleQuery(fmt.Sprintf("SELECT userid FROM %s WHERE chatid = %d and word = '%s'", usedwordsTableName, chatID, theWord))
}

func getFirstEntry(chatID int64) *wordEntry {
	word := &wordEntry{
		chatid: chatID,
	}
	if !gamedb.SingleQuery(fmt.Sprintf("SELECT userid, word, points FROM %s WHERE chatid = %d ORDER BY wordindex ASC LIMIT 1", usedwordsTableName, chatID),
		&word.userid, &word.word, &word.points) {
		return nil
	}
	return word
}

func getLastEntry(chatID int64) *wordEntry {
	word := &wordEntry{
		chatid: chatID,
	}
	if !gamedb.SingleQuery(fmt.Sprintf("SELECT userid, word, points FROM %s WHERE chatid = %d ORDER BY wordindex DESC LIMIT 1", usedwordsTableName, chatID),
		&word.userid, &word.word, &word.points) {
		return nil
	}
	return word
}

func updateFirstEntryPoints(chatID int64, wordsUpdate int) bool {
	// The following simplified version is not working with the current library, but works with the command-line client.
	// return execDb(fmt.Sprintf("UPDATE %s SET points = %d WHERE chatid = %d ORDER BY wordindex ASC LIMIT 1", usedwordsTableName, wordsUpdate, chatID))
	return gamedb.Exec(fmt.Sprintf("UPDATE %s SET points = %d WHERE chatid = %d AND wordindex = (SELECT wordindex FROM %s WHERE chatid = %d ORDER BY wordindex ASC LIMIT 1)", usedwordsTableName, wordsUpdate, chatID, usedwordsTableName, chatID))
}

func getWordHistory(chatID int64) wordList {
	words := make(wordList, 0)
	gamedb.MultiQuery(fmt.Sprintf("SELECT userid, word, points FROM %s WHERE chatid = %d ORDER BY wordindex", usedwordsTableName, chatID),
		func(rows *sql.Rows) bool {
			word := &wordEntry{
				chatid: chatID,
			}
			rows.Scan(&word.userid, &word.word, &word.points)
			words = append(words, word)
			return true
		})
	return words
}

func getNumEntries(chatID int64) int {
	numwords := 0
	gamedb.SingleQuery(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE chatid = %d", usedwordsTableName, chatID), &numwords)
	return numwords
}

func clearWordHistory(chatID int64) {
	// Clear out the word history.
	gamedb.Exec(fmt.Sprintf("DELETE FROM %s WHERE chatid = %d", usedwordsTableName, chatID))
}

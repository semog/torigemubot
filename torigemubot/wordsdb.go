package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const wordsdbFilename = "shiritoriwords.db"

var wordsdb *sql.DB

func getWordPts(theWord string) int {
	// TODO
	// Get wordentry of new word.
	// If word score is zero or not found, then return zero. Probably ends in 'no'.
	// Get wordentry of current word.
	// If first kana of new word does not match ending kana of current word, then return zero.
	// Return word points.
	return len(kanjiExp.FindString(theWord))
}

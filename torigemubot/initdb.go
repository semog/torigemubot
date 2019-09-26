package main

import (
	"github.com/semog/go-sqldb"
)

const gamedbFilename = "torigemu.db"

var gamedb *sqldb.SQLDb

func initgameDb() error {
	var err error
	gamedb, err = sqldb.OpenAndPatchDb(gamedbFilename, gamedbPatchFuncs)
	return err
}

// The array of patch functions that will automatically upgrade the database.
var gamedbPatchFuncs = []sqldb.PatchFuncType{
	// Add new patch functions to this array to automatically upgrade the database.
	{PatchID: 1, PatchFunc: func(sdb *sqldb.SQLDb) error {
		if err := sdb.CreateTable("players (chatid INTEGER, userid INTEGER, firstname TEXT, lastname TEXT, username TEXT, nickname TEXT COLLATE NOCASE, score INTEGER, numwords INTEGER)"); err != nil {
			return err
		}
		if err := sdb.CreateIndex("playerchat_idx ON players (chatid)"); err != nil {
			return err
		}
		if err := sdb.CreateIndex("player_idx ON players (chatid, userid)"); err != nil {
			return err
		}
		if err := sdb.CreateTable("usedwords (chatid INTEGER, userid INTEGER, wordindex INTEGER, word TEXT COLLATE NOCASE, points INTEGER)"); err != nil {
			return err
		}
		if err := sdb.CreateIndex("usedwordschat_idx ON usedwords (chatid)"); err != nil {
			return err
		}
		return sdb.CreateIndex("wordcheck_idx ON usedwords (chatid, word)")
	}},
	{PatchID: 2, PatchFunc: func(sdb *sqldb.SQLDb) error {
		if err := sdb.CreateTable("customwords (chatid INTEGER, userid INTEGER, kanji TEXT, kana TEXT, points INT)"); err != nil {
			return err
		}
		return sdb.CreateIndex("customwords_idx ON customwords (chatid, kanji)")
	}},
}

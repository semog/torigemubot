package main

const gamedbFilename = "torigemu.db"

var gamedb sqldb

func initgameDb() bool {
	gamedb := openAndPatchDb(gamedbFilename, gamedbPatchFuncs)
	return nil != gamedb
}

// The array of patch functions that will automatically upgrade the database.
var gamedbPatchFuncs = []patchFuncType{
	// Add new patch functions to this array to automatically upgrade the database.
	{1, func(sdb *sqldb) bool {
		return sdb.createTable("players (chatid INTEGER, userid INTEGER, firstname TEXT, lastname TEXT, username TEXT, nickname TEXT COLLATE NOCASE, score INTEGER, numwords INTEGER)") &&
			sdb.createIndex("playerchat_idx ON players (chatid)") &&
			sdb.createIndex("player_idx ON players (chatid, userid)") &&
			sdb.createTable("usedwords (chatid INTEGER, userid INTEGER, wordindex INTEGER, word TEXT COLLATE NOCASE, points INTEGER)") &&
			sdb.createIndex("usedwordschat_idx ON usedwords (chatid)") &&
			sdb.createIndex("wordcheck_idx ON usedwords (chatid, word)")
	}},
}

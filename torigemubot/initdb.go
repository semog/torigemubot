package main

const gamedbFilename = "torigemu.db"

var gamedb *SQLDb

func initgameDb() bool {
	gamedb = OpenAndPatchDb(gamedbFilename, gamedbPatchFuncs)
	return nil != gamedb
}

// The array of patch functions that will automatically upgrade the database.
var gamedbPatchFuncs = []patchFuncType{
	// Add new patch functions to this array to automatically upgrade the database.
	{1, func(sdb *SQLDb) bool {
		return sdb.CreateTable("players (chatid INTEGER, userid INTEGER, firstname TEXT, lastname TEXT, username TEXT, nickname TEXT COLLATE NOCASE, score INTEGER, numwords INTEGER)") &&
			sdb.CreateIndex("playerchat_idx ON players (chatid)") &&
			sdb.CreateIndex("player_idx ON players (chatid, userid)") &&
			sdb.CreateTable("usedwords (chatid INTEGER, userid INTEGER, wordindex INTEGER, word TEXT COLLATE NOCASE, points INTEGER)") &&
			sdb.CreateIndex("usedwordschat_idx ON usedwords (chatid)") &&
			sdb.CreateIndex("wordcheck_idx ON usedwords (chatid, word)")
	}},
}

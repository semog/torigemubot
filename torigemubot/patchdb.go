package main

// The array of patch functions that will automatically upgrade the database.
var patchFuncs = []struct {
	// patchid is not necessarily sequential. It just needs to be unique.
	patchid   int
	patchFunc func() bool
}{
	// Add new patch functions to this array to automatically upgrade the database.
	{1, func() bool {
		return createTable("version (patchid INTEGER PRIMARY KEY)") &&
			createTable("players (chatid INTEGER, userid INTEGER, firstname TEXT, lastname TEXT, username TEXT, nickname TEXT COLLATE NOCASE, score INTEGER, numwords INTEGER)") &&
			createIndex("chat_idx ON players (chatid)") &&
			createIndex("player_idx ON players (chatid, userid)") &&
			createTable("usedwords (chatid INTEGER, userid INTEGER, wordorder INTEGER, word TEXT COLLATE NOCASE, points INTEGER)") &&
			createIndex("wordcheck_idx ON usedwords (chatid, word)")
	}},
}

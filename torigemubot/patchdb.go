package main

// The array of patch functions that will automatically upgrade the database.
var patchFuncs = []struct {
	// patchid is not necessarily sequential. It just needs to be unique.
	patchid   int
	patchFunc func() bool
}{
	// Add new patch functions to this array to automatically upgrade the database.
	{1, createInitialTables},
}

func createInitialTables() bool {
	// Create a versions table that tracks patches and upgrades applied to the database.
	return createTable("version (patchid INTEGER PRIMARY KEY)") &&
		createTable("players (id INTEGER PRIMARY KEY, chatid INTEGER, userid INTEGER, firstname TEXT, lastname TEXT, username TEXT, nickname TEXT, score INTEGER, numwords INTEGER)") &&
		createTable("usedwords (chatid INTEGER PRIMARY KEY, wordorder INTEGER, playerid INTEGER, word TEXT, points INTEGER)")
}

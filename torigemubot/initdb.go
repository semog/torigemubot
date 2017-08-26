package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const dbFilename = "torigemu.db"
const patchSavePointName = "patchupdate"

var db *sql.DB

func initDb() bool {
	var err error

	db, err = sql.Open("sqlite3", dbFilename)
	if err != nil {
		return false
	}

	for _, patch := range patchFuncs {
		if !patched(patch.patchid) {
			if !beginPatch() {
				log.Printf("ERROR: Could not begin patch database for version %d.\n", patch.patchid)
				return false
			}
			if !(patch.patchFunc() && commitPatch(patch.patchid)) {
				log.Printf("ERROR: Could not patch database for version %d.\n", patch.patchid)
				rollbackPatch()
				return false
			}
			log.Printf("INFO: Patched database version %d.\n", patch.patchid)
		}
	}
	return true
}

func patched(patchid int) bool {
	// Check for the patchid in the version table
	rows, err := db.Query(fmt.Sprintf("SELECT patchid FROM version WHERE patchid = %d", patchid))
	// If we found it, then it has been patched.
	return err == nil && rows.Next()
}

func beginPatch() bool {
	return createSavePoint(patchSavePointName)
}

func commitPatch(patchid int) bool {
	// Add the patchid to the versions table. If it fails, return false.
	return execDb(fmt.Sprintf("INSERT OR FAIL INTO version (patchid) VALUES (%d)", patchid)) &&
		commitSavePoint(patchSavePointName)
}

func rollbackPatch() {
	execDb("ROLLBACK")
}

func createSavePoint(name string) bool {
	return execDb(fmt.Sprintf("SAVEPOINT %s", name))
}

func commitSavePoint(name string) bool {
	return execDb(fmt.Sprintf("RELEASE SAVEPOINT %s", name))
}

func rollbackSavePoint(name string) bool {
	return execDb(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name))
}

func createTable(tableDef string) bool {
	return execDb(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s", tableDef))
}

func execDb(stmt string) bool {
	statement, err := db.Prepare(stmt)
	if err != nil {
		log.Printf("DBERROR: Preparing %s: %v", stmt, err)
		return false
	}
	_, err = statement.Exec()
	if err != nil {
		log.Printf("DBERROR: Executing %s: %v", stmt, err)
	}
	return err == nil
}

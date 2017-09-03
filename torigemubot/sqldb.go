package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const patchSavePointName = "patchupdate"

type sqldb struct {
	db *sql.DB
}

type patchFuncType struct {
	// patchid is not necessarily sequential. It just needs to be unique.
	patchid   int
	patchFunc func(sdb *sqldb) bool
}

func openAndPatchDb(dbFilename string, patchFuncs []patchFuncType) *sqldb {
	var err error
	sdb := &sqldb{}
	sdb.db, err = sql.Open("sqlite3", dbFilename)
	if err != nil {
		return nil
	}
	// Create the patch tables
	if !patchDb(sdb, internalPatchDbFuncs) {
		return nil
	}
	if !patchDb(sdb, patchFuncs) {
		return nil
	}
	return sdb
}

// The array of patch functions that will automatically upgrade the database.
var internalPatchDbFuncs = []patchFuncType{
	{0, func(sdb *sqldb) bool {
		return sdb.createTable("version (patchid INTEGER PRIMARY KEY)")
	}},
}

func patchDb(sdb *sqldb, patchFuncs []patchFuncType) bool {
	// Currently this patching function does not check to see when it is
	// finished whether it is running against a _newer_ database. An additional
	// check would need to be done to see if the final committed patchid matches the
	// expected patchid.
	for _, patch := range patchFuncs {
		if !sdb.patched(patch.patchid) {
			if !sdb.beginPatch() {
				log.Printf("ERROR: Could not begin patch database for version %d.\n", patch.patchid)
				return false
			}
			if !(patch.patchFunc(sdb) && sdb.commitPatch(patch.patchid)) {
				log.Printf("ERROR: Could not patch database for version %d.\n", patch.patchid)
				sdb.rollbackPatch()
				return false
			}
			log.Printf("INFO: Patched database version %d.\n", patch.patchid)
		}
	}
	return true
}

func (sdb *sqldb) beginTrans() {
	sdb.exec("BEGIN")
}

func (sdb *sqldb) commitTrans() {
	sdb.exec("COMMIT")
}

func (sdb *sqldb) rollbackTrans() {
	sdb.exec("ROLLBACK")
}

func (sdb *sqldb) commitTransOnSuccess(success bool) {
	if success {
		sdb.commitTrans()
	} else {
		sdb.rollbackTrans()
	}
}

func (sdb *sqldb) patched(patchid int) bool {
	// Check for the patchid in the version table
	rows, err := sdb.db.Query(fmt.Sprintf("SELECT patchid FROM version WHERE patchid = %d", patchid))
	defer closeRows(rows)
	// If we found it, then it has been patched.
	return err == nil && rows.Next()
}

func (sdb *sqldb) beginPatch() bool {
	return sdb.createSavePoint(patchSavePointName)
}

func (sdb *sqldb) commitPatch(patchid int) bool {
	// Add the patchid to the versions table. If it fails, return false.
	return sdb.exec(fmt.Sprintf("INSERT OR FAIL INTO version (patchid) VALUES (%d)", patchid)) &&
		sdb.commitSavePoint(patchSavePointName)
}

func (sdb *sqldb) rollbackPatch() {
	sdb.rollbackTrans()
}

func (sdb *sqldb) createSavePoint(name string) bool {
	return sdb.exec(fmt.Sprintf("SAVEPOINT %s", name))
}

func (sdb *sqldb) commitSavePoint(name string) bool {
	return sdb.exec(fmt.Sprintf("RELEASE SAVEPOINT %s", name))
}

func (sdb *sqldb) createTable(tableDef string) bool {
	return sdb.exec(fmt.Sprintf("CREATE TABLE %s", tableDef))
}

func (sdb *sqldb) createIndex(indexDef string) bool {
	return sdb.exec(fmt.Sprintf("CREATE INDEX %s", indexDef))
}

func (sdb *sqldb) exec(stmt string, args ...interface{}) bool {
	statement, err := sdb.db.Prepare(stmt)
	if err != nil {
		log.Printf("DBERROR: Preparing %s: %v", stmt, err)
		return false
	}
	_, err = statement.Exec(args...)
	if err != nil {
		log.Printf("DBERROR: Executing %s: %v", stmt, err)
	}
	return err == nil
}

func (sdb *sqldb) query(stmt string, args ...interface{}) bool {
	rows, err := sdb.db.Query(stmt)
	defer closeRows(rows)
	if err != nil {
		log.Printf("DBERROR: Querying %s: %v", stmt, err)
		return false
	}
	if rows.Next() {
		if args != nil {
			rows.Scan(args...)
		}
		return true
	}
	return false
}

func (sdb *sqldb) multiQuery(stmt string, action func(rows *sql.Rows) bool) bool {
	rows, err := sdb.db.Query(stmt)
	defer closeRows(rows)
	if err != nil {
		log.Printf("DBERROR: Querying %s: %v", stmt, err)
		return false
	}
	for rows.Next() {
		if !action(rows) {
			return false
		}
	}
	return true
}

func closeRows(rows *sql.Rows) {
	if nil != rows {
		rows.Close()
	}
}

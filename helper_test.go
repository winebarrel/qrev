package qrev_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/winebarrel/qrev"
	"github.com/winebarrel/qrev/driver"
)

func testDB(t *testing.T, initSQLs ...string) driver.Driver {
	t.Helper()
	initSQLs = append([]string{qrev.CreateTableSQL, qrev.CreateIndexSQL}, initSQLs...)
	return testDBWithoutTable(t, initSQLs...)
}

func testDBWithoutTable(t *testing.T, initSQLs ...string) driver.Driver {
	t.Helper()
	tempDir := t.TempDir()
	testDB := filepath.Join(tempDir, "test.db")
	dsn := "file:" + testDB
	db, err := sql.Open("sqlite", dsn)

	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	defer db.Close()

	for _, initSQL := range initSQLs {
		_, err = db.Exec(initSQL)

		if err != nil {
			t.Log(err)
			t.FailNow()
		}
	}

	return &driver.SQLite{DSN: dsn}
}

func testDumpDB(t *testing.T, dri driver.Driver) []string {
	t.Helper()
	return testDumpDB0(t, dri, true)
}

func testDumpDBWithoutTime(t *testing.T, dri driver.Driver) []string {
	t.Helper()
	return testDumpDB0(t, dri, false)
}

func testDumpDB0(t *testing.T, dri driver.Driver, withTime bool) []string {
	t.Helper()
	db, err := dri.Open()

	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	sel := sq.Select("filename", "hash", "executed_at", "execution_time", "status", "last_error").
		From(qrev.HistoryTable).OrderBy("filename")
	rows, err := sel.RunWith(db).Query()

	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	defer rows.Close()
	output := []string{}

	for rows.Next() {
		var (
			filename      string
			hash          string
			executedAt    string
			executionTime string
			status        string
			lastError     string
		)

		err = rows.Scan(&filename, &hash, &executedAt, &executionTime, &status, &lastError)

		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		var line []string

		if withTime {
			line = []string{filename, hash, executedAt, executionTime, status, lastError}
		} else {
			line = []string{filename, hash, status, lastError}
		}

		output = append(output, strings.Join(line, " "))
	}

	return output
}

// testBrokenDriver fails to open, for the paths that report a connection that
// could not be established.
func testBrokenDriver() driver.Driver {
	return &driver.MySQL{DSN: "not a dsn"}
}

// testDBWithViewTable backs qrev_history with a view, which reads like the real
// table but rejects every write.
func testDBWithViewTable(t *testing.T, rows ...string) driver.Driver {
	t.Helper()
	initSQLs := []string{strings.Replace(qrev.CreateTableSQL, qrev.HistoryTable, "qrev_history_src", 1)}
	initSQLs = append(initSQLs, rows...)
	initSQLs = append(initSQLs, "CREATE VIEW "+qrev.HistoryTable+" AS SELECT * FROM qrev_history_src")

	return testDBWithoutTable(t, initSQLs...)
}

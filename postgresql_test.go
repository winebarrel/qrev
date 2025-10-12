package qrev_test

import (
	"bytes"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev"
	"github.com/winebarrel/qrev/driver"
)

func TestAcc_PostgreSQL(t *testing.T) {
	if !testAcc {
		t.Skip()
	}

	assert := assert.New(t)
	require := require.New(t)

	t.Chdir(t.TempDir())
	os.WriteFile("20251012-delete-old-data.sql", []byte("select 3"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select 2"), 0400)
	os.WriteFile("20251010-init-table.sql", []byte("select 1"), 0400)

	dri, err := driver.New(testDSN_PostgreSQL, false)
	require.NoError(err)

	{
		db, err := dri.Open()
		require.NoError(err)
		db.Exec("DROP TABLE IF EXISTS " + qrev.HistoryTable)
		t.Cleanup(func() { db.Close() })
	}

	var buf bytes.Buffer
	options := &qrev.Options{Driver: dri, Output: &buf, Timeout: 10 * time.Minute}

	tests := []struct {
		cmd      interface{ Run(*qrev.Options) error }
		expected string
		regexp   *regexp.Regexp
	}{
		{
			cmd:      &qrev.InitCmd{},
			expected: "qrev_history table has been created\n",
		},
		{
			cmd:      &qrev.StatusCmd{},
			expected: "No SQL history\n",
		},
		{
			cmd: &qrev.PlanCmd{Path: "*.sql"},
			expected: `20251010-init-table.sql select 1
20251011-update-data.sql select 2
20251012-delete-old-data.sql select 3
`,
		},
		{
			cmd: &qrev.ApplyCmd{Path: "*.sql"},
			expected: `done 20251010-init-table.sql select 1
done 20251011-update-data.sql select 2
done 20251012-delete-old-data.sql select 3
`,
		},
		{
			cmd:      &qrev.PlanCmd{Path: "*.sql"},
			expected: "No SQL file to run\n",
		},
		{
			cmd:      &qrev.ApplyCmd{Path: "*.sql"},
			expected: "No SQL file to run\n",
		},
		{
			cmd: &qrev.StatusCmd{},
			regexp: regexp.MustCompile(`\d+ \w+ \d+:\d+ done 822ae07 20251010-init-table.sql
\d+ \w+ \d+:\d+ done 3cf988c 20251011-update-data.sql
\d+ \w+ \d+:\d+ done 3efaf2f 20251012-delete-old-data.sql
`),
		},
	}

	for _, test := range tests {
		buf.Reset()
		err := test.cmd.Run(options)
		require.NoError(err)

		if test.expected != "" {
			assert.Equal(test.expected, buf.String())
		}

		if test.regexp != nil {
			assert.Regexp(test.regexp, buf.String())
		}
	}
}

package util_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev/util"
)

func TestPathsToFiles_OK(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tempDir := t.TempDir()
	t.Chdir(tempDir)
	os.WriteFile("20251012-delete-old-data.sql", []byte("select 3"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select 2"), 0400)
	os.WriteFile("20251010-init-table.sql", []byte("select 1"), 0400)

	targets, err := util.PathsToFiles([]string{
		"20251012-delete-old-data.sql",
		"20251011-update-data.sql",
		"20251010-init-table.sql",
	})

	require.NoError(err)

	assert.Equal([]*util.File{
		{
			Path:  "20251010-init-table.sql",
			Name:  "20251010-init-table.sql",
			Hash:  "822ae07d4783158bc1912bb623e5107cc9002d519e1143a9c200ed6ee18b6d0f",
			Rerun: false,
		},
		{
			Path:  "20251011-update-data.sql",
			Name:  "20251011-update-data.sql",
			Hash:  "3cf988cc782b44bc24ceb17e445d9c3cfd06b6c848ecfff949e1bfdb9f705a61",
			Rerun: false,
		},
		{
			Path:  "20251012-delete-old-data.sql",
			Name:  "20251012-delete-old-data.sql",
			Hash:  "3efaf2f2e7527fc540b26c1517859b5446dff36e946f40b56f25f6941b170cc2",
			Rerun: false,
		},
	}, targets)
}

func TestFile_Read(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tempDir := t.TempDir()
	t.Chdir(tempDir)
	os.WriteFile("20251012-delete-old-data.sql", []byte("select 3"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select 2"), 0400)
	os.WriteFile("20251010-init-table.sql", []byte("select 1"), 0400)

	tests := []struct {
		name     string
		expected string
	}{
		{name: "20251012-delete-old-data.sql", expected: "select 3"},
		{name: "20251011-update-data.sql", expected: "select 2"},
		{name: "20251010-init-table.sql", expected: "select 1"},
	}

	for _, test := range tests {
		f := &util.File{Path: test.name}
		b, err := f.Read()
		require.NoError(err)
		assert.Equal(test.expected, b)
	}
}

func TestFile_Head(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tempDir := t.TempDir()
	t.Chdir(tempDir)
	os.WriteFile("20251012-delete-old-data.sql", []byte("select\n3"), 0400)
	os.WriteFile("20251011-update-data.sql", []byte("select\n2"), 0400)
	os.WriteFile("20251010-init-table.sql", []byte("select\n1"), 0400)

	tests := []struct {
		name     string
		expected string
	}{
		{name: "20251012-delete-old-data.sql", expected: "select 3"},
		{name: "20251011-update-data.sql", expected: "select 2"},
		{name: "20251010-init-table.sql", expected: "select 1"},
	}

	for _, test := range tests {
		f := &util.File{Path: test.name}
		b, err := f.Head()
		require.NoError(err)
		assert.Equal(test.expected, b)
	}
}

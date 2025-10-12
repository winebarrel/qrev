package util_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/winebarrel/qrev/util"
)

func TestHash(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	tempDir := t.TempDir()

	tests := []struct {
		data   string
		expect string
	}{
		{data: "hello", expect: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{data: "world", expect: "486ea46224d1bb4fb680f34f7c9ad96a8f24ec88be73ea8e5a6c65260e9cb8a7"},
	}

	for _, test := range tests {
		testDat := filepath.Join(tempDir, test.data+".dat")
		os.WriteFile(testDat, []byte(test.data), 0400)
		hash, err := util.Hash(testDat)
		require.NoError(err)
		assert.Equal(test.expect, hash)
	}
}

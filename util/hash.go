package util

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func Hash(path string) (string, error) {
	f, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer f.Close()
	hash := sha256.New()
	_, err = io.Copy(hash, f)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

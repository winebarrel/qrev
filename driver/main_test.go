package driver_test

import (
	"os"
	"testing"
)

var (
	testAcc = false
)

const (
	testDSN_MySQL      = "root@tcp(127.0.0.1:23307)/mysql"
	testDSN_PostgreSQL = "postgres://postgres@localhost:25433"
)

func TestMain(m *testing.M) {
	if v := os.Getenv("TEST_ACC"); v == "1" {
		testAcc = true
	}

	m.Run()
}

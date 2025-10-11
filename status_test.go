package qrev_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/winebarrel/qrev"
)

func TestStatus_Color(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		status   qrev.Status
		expected string
	}{
		{status: qrev.StatusDone, expected: "done"},
		{status: qrev.StatusFail, expected: "fail"},
		{status: qrev.StatusSkip, expected: "skip"},
	}

	for _, test := range tests {
		assert.Equal(test.expected, test.status.Color())
	}
}

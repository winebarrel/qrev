package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/winebarrel/qrev/util"
)

func TestFormatError(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		s        string
		expected string
	}{
		{s: "aaa", expected: "│ aaa"},
		{s: "\naaa\n", expected: "│ aaa"},
		{s: "aaa\nbbb", expected: "│ aaa\n│ bbb"},
		{s: "\naaa\nbbb\nccc\n", expected: "│ aaa\n│ bbb\n│ ccc"},
	}

	for _, test := range tests {
		output := util.FormatError(test.s)
		assert.Equal(test.expected, output)
	}
}

func TestHeadContent(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		s        string
		expected string
	}{
		{s: "aaa", expected: "aaa"},
		{s: "\naaa\n", expected: "aaa"},
		{s: "aaa \n\n bbb", expected: "aaa bbb"},
		{s: " \n aaa\n\nbbb\n\n\nccc \n ", expected: "aaa bbb ccc"},
		{s: "London Bridge is \n falling down, Falling down, falling down,", expected: "London Bridge is falling down,"},
	}

	for _, test := range tests {
		output := util.HeadContent(test.s)
		assert.Equal(test.expected, output)
	}
}

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateOpts(t *testing.T) {
	cases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "TestCreateOptsWithEmptySlice",
			input:    []string{},
			expected: []string{backOpt},
		},
		{
			name:     "TestCreateOptsWithSingleOption",
			input:    []string{"option1"},
			expected: []string{backOpt, "option1"},
		},
		{
			name:     "TestCreateOptsWithMultipleOptions",
			input:    []string{"option1", "option2", "option3"},
			expected: []string{backOpt, "option1", "option2", "option3"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := createOpts(c.input)
			assert.Equal(t, c.expected, result)
		})
	}
}

func TestInputLocalPort(t *testing.T) {
	port, err := inputLocalPort()
	assert.NoError(t, err)
	assert.Equal(t, "42069", port)
}

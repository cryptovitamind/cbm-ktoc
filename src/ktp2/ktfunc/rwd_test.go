package ktfunc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStartEndBlocks(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedStart uint64
		expectedEnd   uint64
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "valid input",
			input:         "100:200",
			expectedStart: 100,
			expectedEnd:   200,
			expectError:   false,
		},
		{
			name:          "zero start",
			input:         "0:100",
			expectedStart: 0,
			expectedEnd:   100,
			expectError:   false,
		},
		{
			name:        "non-numeric values",
			input:       "abc:def",
			expectError: true,
			errorMsg:    "invalid start block: strconv.ParseUint: parsing \"abc\": invalid syntax",
		},
		{
			name:        "missing end",
			input:       "100:",
			expectError: true,
			errorMsg:    "invalid end block: strconv.ParseUint: parsing \"\": invalid syntax",
		},
		{
			name:        "missing start",
			input:       ":200",
			expectError: true,
			errorMsg:    "invalid start block: strconv.ParseUint: parsing \"\": invalid syntax",
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
			errorMsg:    "invalid start:end blocks format, expected 'start:end'",
		},
		{
			name:        "extra parts",
			input:       "100:200:300",
			expectError: true,
			errorMsg:    "invalid start:end blocks format, expected 'start:end'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ParseStartEndBlocks(tt.input)
			if tt.expectError {
				assert.Error(t, err, "expected an error for input: %s", tt.input)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error(), "error message mismatch")
				}
			} else {
				assert.NoError(t, err, "unexpected error for input: %s", tt.input)
				assert.Equal(t, tt.expectedStart, start, "start block mismatch")
				assert.Equal(t, tt.expectedEnd, end, "end block mismatch")
			}
		})
	}
}

func TestParseWithdrawBlocks(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []uint32
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid input",
			input:       "100,200,300",
			expected:    []uint32{100, 200, 300},
			expectError: false,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
			errorMsg:    "block string cannot be empty",
		},
		{
			name:        "non-numeric value",
			input:       "100,abc,200",
			expectError: true,
			errorMsg:    "invalid block number 'abc' at position 2: strconv.ParseUint: parsing \"abc\": invalid syntax",
		},
		{
			name:        "below range",
			input:       "0",
			expectError: true,
			errorMsg:    "block number 0 at position 1 is out of range (1-5000000)",
		},
		{
			name:        "above range",
			input:       "5000001",
			expectError: true,
			errorMsg:    "block number 5000001 at position 1 is out of range (1-5000000)",
		},
		{
			name:        "empty part",
			input:       "100,,200",
			expectError: true,
			errorMsg:    "block 2 is empty",
		},
		{
			name:        "spaces in input",
			input:       " 100 , 200 ",
			expected:    []uint32{100, 200},
			expectError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseWithdrawBlocks(tt.input)
			if tt.expectError {
				assert.Error(t, err, "expected an error for input: %s", tt.input)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error(), "error message mismatch")
				}
			} else {
				assert.NoError(t, err, "unexpected error for input: %s", tt.input)
				assert.Equal(t, tt.expected, result, "result mismatch")
			}
		})
	}
}

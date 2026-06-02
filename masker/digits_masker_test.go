package masker

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestDigitsMasker_Mask(t *testing.T) {
    masker := DigitsMasker{}
    ctx := context.Background()

    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"only digits", "12345", "*****"},
        {"only letters", "abcde", "abcde"},
        {"mixed", "a1b2c3", "a*b*c*"},
        {"empty", "", ""},
        {"with spaces", "abc 123", "abc ***"},
        {"russian letters", "абв123", "абв***"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := masker.Mask(ctx, tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
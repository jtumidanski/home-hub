package entry

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateValue_Sentiment(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr error
	}{
		{"positive", `{"rating":"positive"}`, nil},
		{"neutral", `{"rating":"neutral"}`, nil},
		{"negative", `{"rating":"negative"}`, nil},
		{"invalid rating", `{"rating":"bad"}`, ErrInvalidSentiment},
		{"empty", `{}`, ErrInvalidSentiment},
		{"null", `null`, ErrValueRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateValue("sentiment", json.RawMessage(tt.value), nil)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateValue_Numeric(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr error
	}{
		{"zero", `{"count":0}`, nil},
		{"positive", `{"count":5}`, nil},
		{"negative", `{"count":-1}`, ErrInvalidNumeric},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateValue("numeric", json.RawMessage(tt.value), nil)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateValue_Range(t *testing.T) {
	cfg := json.RawMessage(`{"min":0,"max":100}`)
	tests := []struct {
		name    string
		value   string
		wantErr error
	}{
		{"in bounds", `{"value":50}`, nil},
		{"min bound", `{"value":0}`, nil},
		{"max bound", `{"value":100}`, nil},
		{"below min", `{"value":-1}`, ErrInvalidRange},
		{"above max", `{"value":101}`, ErrInvalidRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateValue("range", json.RawMessage(tt.value), cfg)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

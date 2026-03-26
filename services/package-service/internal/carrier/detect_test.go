package carrier

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name           string
		trackingNumber string
		wantCarrier    string
		wantConfidence string
	}{
		{
			name:           "UPS tracking number",
			trackingNumber: "1Z999AA10123456784",
			wantCarrier:    "ups",
			wantConfidence: ConfidenceHigh,
		},
		{
			name:           "UPS lowercase normalized",
			trackingNumber: "1z999aa10123456784",
			wantCarrier:    "ups",
			wantConfidence: ConfidenceHigh,
		},
		{
			name:           "FedEx 12-digit",
			trackingNumber: "123456789012",
			wantCarrier:    "fedex",
			wantConfidence: ConfidenceHigh,
		},
		{
			name:           "FedEx 15-digit",
			trackingNumber: "123456789012345",
			wantCarrier:    "fedex",
			wantConfidence: ConfidenceHigh,
		},
		{
			name:           "FedEx 20-digit also matches USPS",
			trackingNumber: "94001234567890123456",
			wantCarrier:    "fedex",
			wantConfidence: ConfidenceMedium, // matches FedEx 20-digit and USPS patterns
		},
		{
			name:           "USPS 22-digit",
			trackingNumber: "9400123456789012345678",
			wantCarrier:    "usps",
			wantConfidence: ConfidenceHigh, // 22 digits matches USPS only
		},
		{
			name:           "unknown tracking number",
			trackingNumber: "ABCXYZ",
			wantCarrier:    "",
			wantConfidence: ConfidenceLow,
		},
		{
			name:           "empty tracking number",
			trackingNumber: "",
			wantCarrier:    "",
			wantConfidence: ConfidenceLow,
		},
		{
			name:           "whitespace trimmed",
			trackingNumber: "  1Z999AA10123456784  ",
			wantCarrier:    "ups",
			wantConfidence: ConfidenceHigh,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Detect(tc.trackingNumber)
			require.Equal(t, tc.wantCarrier, result.DetectedCarrier)
			require.Equal(t, tc.wantConfidence, result.Confidence)
		})
	}
}

func TestDetect_AmbiguousNumber(t *testing.T) {
	// A 20-digit number could match both FedEx and USPS patterns
	result := Detect("12345678901234567890")
	require.Equal(t, ConfidenceMedium, result.Confidence)
	require.NotEmpty(t, result.DetectedCarrier)
}

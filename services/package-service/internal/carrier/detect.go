package carrier

import (
	"regexp"
	"strings"
)

const (
	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
	ConfidenceLow    = "low"
)

type DetectionResult struct {
	TrackingNumber  string
	DetectedCarrier string
	Confidence      string
}

// RestDetectionModel is the JSON:API representation for carrier detection.
type RestDetectionModel struct {
	TrackingNumber  string `json:"trackingNumber"`
	DetectedCarrier string `json:"detectedCarrier"`
	Confidence      string `json:"confidence"`
}

func (r RestDetectionModel) GetName() string       { return "carrierDetections" }
func (r RestDetectionModel) GetID() string          { return "detect" }
func (r *RestDetectionModel) SetID(_ string) error  { return nil }
func (r *RestDetectionModel) SetToOneReferenceID(_ string, _ string) error {
	return nil
}

// Carrier patterns
var (
	// UPS: starts with 1Z followed by 16 alphanumeric chars
	upsPattern = regexp.MustCompile(`^1Z[0-9A-Z]{16}$`)

	// FedEx: 12, 15, or 20 digits
	fedexPattern12 = regexp.MustCompile(`^\d{12}$`)
	fedexPattern15 = regexp.MustCompile(`^\d{15}$`)
	fedexPattern20 = regexp.MustCompile(`^\d{20}$`)

	// USPS: 20-22 digits, or starts with specific prefixes
	uspsPattern20_22 = regexp.MustCompile(`^\d{20,22}$`)
	uspsPrefixPattern = regexp.MustCompile(`^(94|92|93|70|23)\d{18,20}$`)
)

// Detect analyzes a tracking number and returns the detected carrier and confidence.
func Detect(trackingNumber string) DetectionResult {
	tn := strings.TrimSpace(strings.ToUpper(trackingNumber))

	result := DetectionResult{
		TrackingNumber: trackingNumber,
	}

	var matches []string

	if upsPattern.MatchString(tn) {
		matches = append(matches, "ups")
	}

	if fedexPattern12.MatchString(tn) || fedexPattern15.MatchString(tn) || fedexPattern20.MatchString(tn) {
		matches = append(matches, "fedex")
	}

	if uspsPattern20_22.MatchString(tn) || uspsPrefixPattern.MatchString(tn) {
		matches = append(matches, "usps")
	}

	switch len(matches) {
	case 0:
		result.Confidence = ConfidenceLow
	case 1:
		result.DetectedCarrier = matches[0]
		result.Confidence = ConfidenceHigh
	default:
		// Multiple matches — pick most specific, mark medium confidence
		result.DetectedCarrier = matches[0]
		result.Confidence = ConfidenceMedium
	}

	return result
}

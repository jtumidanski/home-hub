package planner

import (
	"testing"
)

func ptrInt(v int) *int { return &v }

func buildConfig(classification string, servingsYield *int) *Model {
	b := NewBuilder().SetClassification(classification)
	if servingsYield != nil {
		b.SetServingsYield(servingsYield)
	}
	m, _ := b.Build()
	return &m
}

func TestComputeReadiness(t *testing.T) {
	tests := []struct {
		name           string
		config         *Model
		recipeServings *int
		wantReady      bool
		wantIssues     []string
	}{
		{
			name:           "nil config returns not ready with missing config issue",
			config:         nil,
			recipeServings: nil,
			wantReady:      false,
			wantIssues:     []string{"planner configuration is missing"},
		},
		{
			name:           "empty classification returns not ready",
			config:         buildConfig("", ptrInt(4)),
			recipeServings: nil,
			wantReady:      false,
			wantIssues:     []string{"classification is not set"},
		},
		{
			name:           "no servings returns not ready",
			config:         buildConfig("dinner", nil),
			recipeServings: nil,
			wantReady:      false,
			wantIssues:     []string{"servings is not set"},
		},
		{
			name:           "classification set and servingsYield set returns ready",
			config:         buildConfig("dinner", ptrInt(4)),
			recipeServings: nil,
			wantReady:      true,
			wantIssues:     []string{},
		},
		{
			name:           "classification set and recipe servings set returns ready",
			config:         buildConfig("dinner", nil),
			recipeServings: ptrInt(6),
			wantReady:      true,
			wantIssues:     []string{},
		},
		{
			name:           "multiple issues combined",
			config:         buildConfig("", nil),
			recipeServings: nil,
			wantReady:      false,
			wantIssues:     []string{"classification is not set", "servings is not set"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ComputeReadiness(tt.config, tt.recipeServings)

			if r.Ready != tt.wantReady {
				t.Errorf("Ready = %v, want %v", r.Ready, tt.wantReady)
			}

			if len(r.Issues) != len(tt.wantIssues) {
				t.Fatalf("Issues = %v (len %d), want %v (len %d)", r.Issues, len(r.Issues), tt.wantIssues, len(tt.wantIssues))
			}

			for i, issue := range r.Issues {
				if issue != tt.wantIssues[i] {
					t.Errorf("Issues[%d] = %q, want %q", i, issue, tt.wantIssues[i])
				}
			}
		})
	}
}

package hcl

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
)

func TestParseFiles(t *testing.T) {
	tests := []struct {
		name          string
		paths         []string
		expectedBody  hcl.Body
		expectedError string
	}{
		{
			name: "valid configurations",
			paths: []string{
				"./testdata/valid1.hcl",
				"./testdata/valid2.hcl",
			},
		},
		{
			name: "invalid configurations",
			paths: []string{
				"./testdata/valid1.hcl",
				"./testdata/invalid.hcl",
			},
			expectedError: "invalid.hcl", // Invalid file must be reported.
		},
		{
			name: "non-existent file",
			paths: []string{
				"./testdata/valid1.hcl",
				"./testdata/non-existent.hcl",
			},
			expectedError: "Cannot read file ./testdata/non-existent.hcl", // Non-existent file must be reported.
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, diags := ParseFiles(tt.paths, nil)
			if len(tt.expectedError) > 0 {
				assert.NotNil(t, diags)
				assert.True(t, diags.HasErrors())
				assert.Contains(t, diags.Error(), tt.expectedError)

			} else {
				assert.Nil(t, diags)
				assert.False(t, diags.HasErrors())
				assert.NotNil(t, body)
			}
		})
	}
}

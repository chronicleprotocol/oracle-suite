package config

import (
	netURL "net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestURL_MarshalHCL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedCty cty.Value
		wantErr     bool
	}{
		{
			name:        "valid URL",
			url:         "https://example.com",
			expectedCty: cty.StringVal("https://example.com"),
			wantErr:     false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url, _ := netURL.Parse(test.url)
			u := URL(*url)
			cty, err := u.MarshalHCL()
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedCty, cty)
			}
		})
	}
}

func TestURL_UnmarshalHCL(t *testing.T) {
	tests := []struct {
		name        string
		value       cty.Value
		expectedURL string
		wantErr     bool
	}{
		{
			name:        "valid URL",
			value:       cty.StringVal("https://example.com"),
			expectedURL: "https://example.com",
			wantErr:     false,
		},
		{
			name:    "missing-scheme",
			value:   cty.StringVal("example.com"),
			wantErr: true,
		},
		{
			name:    "missing-host",
			value:   cty.StringVal("/test"),
			wantErr: true,
		},
		{
			name:    "non-string value",
			value:   cty.NumberIntVal(42),
			wantErr: true,
		},
		{
			name:        "unknown value",
			value:       cty.UnknownVal(cty.String),
			expectedURL: "",
			wantErr:     false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var url URL
			err := url.UnmarshalHCL(test.value)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedURL, url.String())
			}
		})
	}
}

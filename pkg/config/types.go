package config

import (
	"fmt"
	netURL "net/url"

	"github.com/zclconf/go-cty/cty"
)

// URL accepts a URL as a string.
// It returns an error if the string is not a valid URL.
type URL netURL.URL

// String implements the fmt.Stringer interface.
func (u *URL) String() string {
	return (*netURL.URL)(u).String()
}

func (u URL) MarshalHCL() (cty.Value, error) {
	return cty.StringVal(u.String()), nil
}

// UnmarshalHCL implements the hcl.Unmarshaler interface.
func (u *URL) UnmarshalHCL(value cty.Value) error {
	if !value.IsKnown() {
		return nil
	}
	if value.Type() != cty.String {
		return fmt.Errorf("URL must be a string")
	}
	url, err := netURL.Parse(value.AsString())
	if err != nil {
		return err
	}
	if url.Scheme == "" {
		return fmt.Errorf("URL must have a scheme")
	}
	if url.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	*u = URL(*url)
	return nil
}

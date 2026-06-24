package main

import "strings"

// scopeFlags is a repeatable --scope flag value.
// Each invocation appends one selector to the slice.
type scopeFlags []string

func (sf *scopeFlags) String() string { return strings.Join(*sf, ", ") }
func (sf *scopeFlags) Set(s string) error {
	*sf = append(*sf, s)
	return nil
}

package main

import (
	"strconv"

	"github.com/urfave/cli"
)

// optionalBool is a boolean with a separate presence flag.
type optionalBool struct {
	present bool
	value   bool
}

// optionalBool is a cli.Generic == flag.Value implementation equivalent to
// the one underlying flag.Bool, except that it records whether the flag has been set.
// This is distinct from optionalBool to (pretend to) force callers to use
// newOptionalBool
type optionalBoolValue optionalBool

func newOptionalBoolValue(p *optionalBool) cli.Generic {
	p.present = false
	return (*optionalBoolValue)(p)
}

func (ob *optionalBoolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	ob.value = v
	ob.present = true
	return nil
}

func (ob *optionalBoolValue) String() string {
	if !ob.present {
		return "" // This is, sadly, not round-trip safe: --flag is interpreted as --flag=true
	}
	return strconv.FormatBool(ob.value)
}

func (ob *optionalBoolValue) IsBoolFlag() bool {
	return true
}

// optionalString is a string with a separate presence flag.
type optionalString struct {
	present bool
	value   string
}

// optionalString is a cli.Generic == flag.Value implementation equivalent to
// the one underlying flag.String, except that it records whether the flag has been set.
// This is distinct from optionalString to (pretend to) force callers to use
// newoptionalString
type optionalStringValue optionalString

func newOptionalStringValue(p *optionalString) cli.Generic {
	p.present = false
	return (*optionalStringValue)(p)
}

func (ob *optionalStringValue) Set(s string) error {
	ob.value = s
	ob.present = true
	return nil
}

func (ob *optionalStringValue) String() string {
	if !ob.present {
		return "" // This is, sadly, not round-trip safe: --flag= is interpreted as {present:true, value:""}
	}
	return ob.value
}

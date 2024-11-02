package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestOptionalBoolSet(t *testing.T) {
	for _, c := range []struct {
		input    string
		accepted bool
		value    bool
	}{
		// Valid inputs documented for strconv.ParseBool == flag.BoolVar
		{"1", true, true},
		{"t", true, true},
		{"T", true, true},
		{"TRUE", true, true},
		{"true", true, true},
		{"True", true, true},
		{"0", true, false},
		{"f", true, false},
		{"F", true, false},
		{"FALSE", true, false},
		{"false", true, false},
		{"False", true, false},
		// A few invalid inputs
		{"", false, false},
		{"yes", false, false},
		{"no", false, false},
		{"2", false, false},
	} {
		var ob optionalBool
		v := newOptionalBoolValue(&ob)
		require.False(t, ob.present)
		err := v.Set(c.input)
		if c.accepted {
			assert.NoError(t, err, c.input)
			assert.Equal(t, c.value, ob.value)
		} else {
			assert.Error(t, err, c.input)
			assert.False(t, ob.present) // Just to be extra paranoid.
		}
	}

	// Nothing actually explicitly says that .Set() is never called when the flag is not present on the command line;
	// so, check that it is not being called, at least in the straightforward case (it's not possible to test that it
	// is not called in any possible situation).
	var globalOB, commandOB optionalBool
	actionRun := false
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.GenericFlag{
			Name:  "global-OB",
			Value: newOptionalBoolValue(&globalOB),
		},
	}
	app.Commands = []cli.Command{{
		Name: "cmd",
		Flags: []cli.Flag{
			cli.GenericFlag{
				Name:  "command-OB",
				Value: newOptionalBoolValue(&commandOB),
			},
		},
		Action: func(*cli.Context) error {
			assert.False(t, globalOB.present)
			assert.False(t, commandOB.present)
			actionRun = true
			return nil
		},
	}}
	err := app.Run([]string{"app", "cmd"})
	require.NoError(t, err)
	assert.True(t, actionRun)
}

func TestOptionalBoolString(t *testing.T) {
	for _, c := range []struct {
		input    optionalBool
		expected string
	}{
		{optionalBool{present: true, value: true}, "true"},
		{optionalBool{present: true, value: false}, "false"},
		{optionalBool{present: false, value: true}, ""},
		{optionalBool{present: false, value: false}, ""},
	} {
		var ob optionalBool
		v := newOptionalBoolValue(&ob)
		ob = c.input
		res := v.String()
		assert.Equal(t, c.expected, res)
	}
}

func TestOptionalBoolIsBoolFlag(t *testing.T) {
	// IsBoolFlag means that the argument value must either be part of the same argument, with =;
	// if there is no =, the value is set to true.
	// This differs form other flags, where the argument is required and may be either separated with = or supplied in the next argument.
	for _, c := range []struct {
		input        []string
		expectedOB   optionalBool
		expectedArgs []string
	}{
		{[]string{"1", "2"}, optionalBool{present: false}, []string{"1", "2"}},                                       // Flag not present
		{[]string{"--OB=true", "1", "2"}, optionalBool{present: true, value: true}, []string{"1", "2"}},              // --OB=true
		{[]string{"--OB=false", "1", "2"}, optionalBool{present: true, value: false}, []string{"1", "2"}},            // --OB=false
		{[]string{"--OB", "true", "1", "2"}, optionalBool{present: true, value: true}, []string{"true", "1", "2"}},   // --OB true
		{[]string{"--OB", "false", "1", "2"}, optionalBool{present: true, value: true}, []string{"false", "1", "2"}}, // --OB false
	} {
		var ob optionalBool
		actionRun := false
		app := cli.NewApp()
		app.Commands = []cli.Command{{
			Name: "cmd",
			Flags: []cli.Flag{
				cli.GenericFlag{
					Name:  "OB",
					Value: newOptionalBoolValue(&ob),
				},
			},
			Action: func(ctx *cli.Context) error {
				assert.Equal(t, c.expectedOB, ob)
				assert.Equal(t, c.expectedArgs, ([]string)(ctx.Args()))
				actionRun = true
				return nil
			},
		}}
		err := app.Run(append([]string{"app", "cmd"}, c.input...))
		require.NoError(t, err)
		assert.True(t, actionRun)
	}
}

func TestOptionalStringSet(t *testing.T) {
	// Really just a smoke test, but differentiating between not present and empty.
	for _, c := range []string{"", "hello"} {
		var os optionalString
		v := newOptionalStringValue(&os)
		require.False(t, os.present)
		err := v.Set(c)
		assert.NoError(t, err, c)
		assert.Equal(t, c, os.value)
	}

	// Nothing actually explicitly says that .Set() is never called when the flag is not present on the command line;
	// so, check that it is not being called, at least in the straightforward case (it's not possible to test that it
	// is not called in any possible situation).
	var globalOS, commandOS optionalString
	actionRun := false
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.GenericFlag{
			Name:  "global-OS",
			Value: newOptionalStringValue(&globalOS),
		},
	}
	app.Commands = []cli.Command{{
		Name: "cmd",
		Flags: []cli.Flag{
			cli.GenericFlag{
				Name:  "command-OS",
				Value: newOptionalStringValue(&commandOS),
			},
		},
		Action: func(*cli.Context) error {
			assert.False(t, globalOS.present)
			assert.False(t, commandOS.present)
			actionRun = true
			return nil
		},
	}}
	err := app.Run([]string{"app", "cmd"})
	require.NoError(t, err)
	assert.True(t, actionRun)
}

func TestOptionalStringString(t *testing.T) {
	for _, c := range []struct {
		input    optionalString
		expected string
	}{
		{optionalString{present: true, value: "hello"}, "hello"},
		{optionalString{present: true, value: ""}, ""},
		{optionalString{present: false, value: "hello"}, ""},
		{optionalString{present: false, value: ""}, ""},
	} {
		var os optionalString
		v := newOptionalStringValue(&os)
		os = c.input
		res := v.String()
		assert.Equal(t, c.expected, res)
	}
}

func TestOptionalStringIsBoolFlag(t *testing.T) {
	// NOTE: optionalStringValue does not implement IsBoolFlag!
	// IsBoolFlag means that the argument value must either be part of the same argument, with =;
	// if there is no =, the value is set to true.
	// This differs form other flags, where the argument is required and may be either separated with = or supplied in the next argument.
	for _, c := range []struct {
		input        []string
		expectedOS   optionalString
		expectedArgs []string
	}{
		{[]string{"1", "2"}, optionalString{present: false}, []string{"1", "2"}},                                 // Flag not present
		{[]string{"--OS=hello", "1", "2"}, optionalString{present: true, value: "hello"}, []string{"1", "2"}},    // --OS=true
		{[]string{"--OS=", "1", "2"}, optionalString{present: true, value: ""}, []string{"1", "2"}},              // --OS=false
		{[]string{"--OS", "hello", "1", "2"}, optionalString{present: true, value: "hello"}, []string{"1", "2"}}, // --OS true
		{[]string{"--OS", "", "1", "2"}, optionalString{present: true, value: ""}, []string{"1", "2"}},           // --OS false
	} {
		var os optionalString
		actionRun := false
		app := cli.NewApp()
		app.Commands = []cli.Command{{
			Name: "cmd",
			Flags: []cli.Flag{
				cli.GenericFlag{
					Name:  "OS",
					Value: newOptionalStringValue(&os),
				},
			},
			Action: func(ctx *cli.Context) error {
				assert.Equal(t, c.expectedOS, os)
				assert.Equal(t, c.expectedArgs, ([]string)(ctx.Args()))
				actionRun = true
				return nil
			},
		}}
		err := app.Run(append([]string{"app", "cmd"}, c.input...))
		require.NoError(t, err)
		assert.True(t, actionRun)
	}
}

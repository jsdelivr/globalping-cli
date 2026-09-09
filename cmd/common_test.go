package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/version"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func Test_ParseTargets(t *testing.T) {
	for _, test := range []struct {
		name    string
		input   string
		want    []string
		wantErr string
	}{
		{name: "single", input: " example.com ", want: []string{"example.com"}},
		{name: "comparison whitespace", input: " first.example , second.example ", want: []string{"first.example", "second.example"}},
		{name: "encoded comma", input: "https://example.com/a%2Cb", want: []string{"https://example.com/a%2Cb"}},
		{name: "empty first", input: ",example.com", wantErr: "provided target is empty"},
		{name: "empty second", input: "example.com,", wantErr: "provided target is empty"},
		{name: "duplicate", input: "example.com,example.com", wantErr: "comparison targets must be distinct"},
		{name: "too many", input: "one,two,three", wantErr: "maximum of two targets"},
		{name: "URL comma", input: "https://example.com/a,b", want: []string{"https://example.com/a", "b"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			targets, err := parseTargets(test.input)

			if test.wantErr != "" {
				assert.ErrorContains(t, err, test.wantErr)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.want, targets)
		})
	}
}

func Test_UpdateContext_ComparisonModes(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"globalping", "ping", "one.example,two.example"}

	for _, test := range []struct {
		name    string
		flags   map[string]string
		wantErr string
	}{
		{name: "automatic table"},
		{name: "explicit table", flags: map[string]string{"table": "true"}},
		{name: "CI", flags: map[string]string{"ci": "true"}},
		{name: "inactive flags", flags: map[string]string{"json": "false", "latency": "false", "infinite": "false"}},
		{name: "table false", flags: map[string]string{"table": "false"}, wantErr: "table output cannot be disabled"},
		{name: "JSON", flags: map[string]string{"json": "true"}, wantErr: "json flag is not supported"},
		{name: "latency", flags: map[string]string{"latency": "true"}, wantErr: "latency flag is not supported"},
		{name: "infinite", flags: map[string]string{"infinite": "true"}, wantErr: "infinite flag is not supported"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := createDefaultContext()
			root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, nil, nil, nil, nil, nil)
			cmd, _, err := root.Cmd.Find([]string{"ping"})
			assert.NoError(t, err)

			for name, value := range test.flags {
				flags := cmd.Flags()

				if flags.Lookup(name) == nil {
					flags = root.Cmd.PersistentFlags()
				}

				assert.NoError(t, flags.Set(name, value))
			}

			err = root.updateContext(cmd, []string{"one.example,two.example"})

			if test.wantErr != "" {
				assert.ErrorContains(t, err, test.wantErr)

				return
			}

			assert.NoError(t, err)
			assert.True(t, ctx.Comparison)
			assert.True(t, ctx.Table)
			assert.Equal(t, "one.example", ctx.Target)
			assert.Equal(t, []string{"one.example", "two.example"}, ctx.Targets)
		})
	}
}

func Test_UpdateContext_ComparisonValidatesEveryTargetIPVersion(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"globalping", "ping", "example.com,1.1.1.1"}

	ctx := createDefaultContext()
	ctx.Ipv4 = true
	root := NewRoot(view.NewPrinter(nil, new(bytes.Buffer), new(bytes.Buffer)), ctx, nil, nil, nil, nil, nil)
	cmd, _, err := root.Cmd.Find([]string{"ping"})
	assert.NoError(t, err)

	err = root.updateContext(cmd, []string{"example.com,1.1.1.1"})

	assert.ErrorIs(t, err, ErrTargetIPVersionNotAllowed)
	assert.True(t, isIPTarget("http", "https://[2001:db8::1]/path"))
}

func Test_UpdateContext(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"no_arg":                test_updateContext_NoArg,
		"country":               test_updateContext_Country,
		"country_whitespace":    test_updateContext_CountryWhitespace,
		"no_target":             test_updateContext_NoTarget,
		"limit_below_minimum":   test_updateContext_LimitBelowMinimum,
		"ci_env":                test_updateContext_CIEnv,
		"target_not_hostname":   test_updateContext_TargetIsNotAHostname,
		"resolver_not_hostname": test_updateContext_ResolverIsNotAHostname,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

func test_updateContext_NoArg(t *testing.T) {
	ctx := createDefaultContext()
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	cmd := &cobra.Command{Use: "ping"}
	assert.NoError(t, cmd.Execute())

	err := root.updateContext(cmd, []string{"1.1.1.1"})
	assert.Equal(t, "ping", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "world", ctx.From)
	assert.NoError(t, err)
}

func test_updateContext_Country(t *testing.T) {
	ctx := createDefaultContext()
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	cmd := &cobra.Command{Use: "ping"}
	assert.NoError(t, cmd.Execute())

	err := root.updateContext(cmd, []string{"1.1.1.1", "from", "Germany"})
	assert.Equal(t, "ping", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "Germany", ctx.From)
	assert.NoError(t, err)
}

// Check if country with whitespace is parsed correctly
func test_updateContext_CountryWhitespace(t *testing.T) {
	ctx := createDefaultContext()
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	cmd := &cobra.Command{Use: "ping"}
	assert.NoError(t, cmd.Execute())

	err := root.updateContext(cmd, []string{"1.1.1.1", "from", " Germany, France"})
	assert.Equal(t, "ping", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "Germany, France", ctx.From)
	assert.NoError(t, err)
}

func test_updateContext_NoTarget(t *testing.T) {
	ctx := createDefaultContext()
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	cmd := &cobra.Command{Use: "ping"}
	assert.NoError(t, cmd.Execute())

	err := root.updateContext(cmd, []string{})
	assert.Error(t, err)
}

func test_updateContext_LimitBelowMinimum(t *testing.T) {
	ctx := createDefaultContext()
	ctx.Limit = 0
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	cmd := &cobra.Command{Use: "ping"}
	assert.NoError(t, cmd.Execute())

	err := root.updateContext(cmd, []string{"1.1.1.1"})
	assert.EqualError(t, err, "limit must be at least 1")
}

func test_updateContext_CIEnv(t *testing.T) {
	oldCI := os.Getenv("CI")
	t.Setenv("CI", "true")
	defer t.Setenv("CI", oldCI)

	ctx := createDefaultContext()
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	cmd := &cobra.Command{Use: "ping"}
	assert.NoError(t, cmd.Execute())

	err := root.updateContext(cmd, []string{"1.1.1.1"})
	assert.Equal(t, "ping", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "world", ctx.From)
	assert.True(t, ctx.CIMode)
	assert.NoError(t, err)
}

func test_updateContext_TargetIsNotAHostname(t *testing.T) {
	ctx := createDefaultContext()
	ctx.Ipv4 = true
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	cmd := &cobra.Command{Use: "ping"}
	assert.NoError(t, cmd.Execute())

	err := root.updateContext(cmd, []string{"1.1.1.1"})
	assert.EqualError(t, err, ErrTargetIPVersionNotAllowed.Error())

	ctx.Ipv4 = false
	ctx.Ipv6 = true
	err = root.updateContext(cmd, []string{"1.1.1.1"})
	assert.EqualError(t, err, ErrTargetIPVersionNotAllowed.Error())
}

func test_updateContext_ResolverIsNotAHostname(t *testing.T) {
	ctx := createDefaultContext()
	ctx.Ipv4 = true
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil, nil)

	cmd := &cobra.Command{Use: "dns"}
	assert.NoError(t, cmd.Execute())

	err := root.updateContext(cmd, []string{"example.com", "@1.1.1.1"})
	assert.EqualError(t, err, ErrResolverIPVersionNotAllowed.Error())

	ctx.Ipv4 = false
	ctx.Ipv6 = true
	err = root.updateContext(cmd, []string{"example.com", "@1.1.1.1"})
	assert.EqualError(t, err, ErrResolverIPVersionNotAllowed.Error())
}

func Test_ParseTargetQuery_Simple(t *testing.T) {
	cmd := "ping"
	args := []string{"example.com"}

	q, err := parseTargetQuery(cmd, args)
	assert.NoError(t, err)

	assert.Equal(t, TargetQuery{Target: "example.com", From: ""}, *q)
}

func Test_ParseTargetQuery_SimpleWithResolver(t *testing.T) {
	cmd := "dns"
	args := []string{"example.com", "@1.1.1.1"}

	q, err := parseTargetQuery(cmd, args)
	assert.NoError(t, err)

	assert.Equal(t, TargetQuery{Target: "example.com", From: "", Resolver: "1.1.1.1"}, *q)
}

func Test_ParseTargetQuery_ResolverNotAllowed(t *testing.T) {
	cmd := "ping"
	args := []string{"example.com", "@1.1.1.1"}

	_, err := parseTargetQuery(cmd, args)
	assert.ErrorContains(t, err, "does not accept a resolver argument")
}

func Test_ParseTargetQuery_TargetFromX(t *testing.T) {
	cmd := "ping"
	args := []string{"example.com", "from", "London"}

	q, err := parseTargetQuery(cmd, args)
	assert.NoError(t, err)

	assert.Equal(t, TargetQuery{Target: "example.com", From: "London"}, *q)
}

func Test_ParseTargetQuery_TargetFromXWithResolver(t *testing.T) {
	cmd := "http"
	args := []string{"example.com", "from", "London", "@1.1.1.1"}

	q, err := parseTargetQuery(cmd, args)
	assert.NoError(t, err)

	assert.Equal(t, TargetQuery{Target: "example.com", From: "London", Resolver: "1.1.1.1"}, *q)
}

func Test_FindAndRemoveResolver_SimpleNoResolver(t *testing.T) {
	args := []string{"example.com"}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	assert.Equal(t, "", resolver)
	assert.Equal(t, args, argsWithoutResolver)
}

func Test_FindAndRemoveResolver_NoResolver(t *testing.T) {
	args := []string{"example.com", "from", "London"}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	assert.Equal(t, "", resolver)
	assert.Equal(t, args, argsWithoutResolver)
}

func Test_FindAndRemoveResolver_ResolverAndFrom(t *testing.T) {
	args := []string{"example.com", "@1.1.1.1", "from", "London"}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	assert.Equal(t, "1.1.1.1", resolver)
	assert.Equal(t, []string{"example.com", "from", "London"}, argsWithoutResolver)
}

func Test_FindAndRemoveResolver_ResolverOnly(t *testing.T) {
	args := []string{"example.com", "@1.1.1.1"}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	assert.Equal(t, "1.1.1.1", resolver)
	assert.Equal(t, []string{"example.com"}, argsWithoutResolver)
}

func TestUserAgent(t *testing.T) {
	version.Version = "x.y.z"
	assert.Equal(t, "globalping-cli/vx.y.z (https://github.com/jsdelivr/globalping-cli)", getUserAgent())
}

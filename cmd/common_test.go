package cmd

import (
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/version"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
)

func Test_UpdateContext(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"no_arg":                test_updateContext_NoArg,
		"country":               test_updateContext_Country,
		"country_whitespace":    test_updateContext_CountryWhitespace,
		"no_target":             test_updateContext_NoTarget,
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
	ctx := createDefaultContext("ping")
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

	err := root.updateContext("test", []string{"1.1.1.1"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "world", ctx.From)
	assert.NoError(t, err)
}

func test_updateContext_Country(t *testing.T) {
	ctx := createDefaultContext("ping")
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

	err := root.updateContext("test", []string{"1.1.1.1", "from", "Germany"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "Germany", ctx.From)
	assert.NoError(t, err)
}

// Check if country with whitespace is parsed correctly
func test_updateContext_CountryWhitespace(t *testing.T) {
	ctx := createDefaultContext("ping")
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

	err := root.updateContext("test", []string{"1.1.1.1", "from", " Germany, France"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "Germany, France", ctx.From)
	assert.NoError(t, err)
}

func test_updateContext_NoTarget(t *testing.T) {
	ctx := createDefaultContext("ping")
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

	err := root.updateContext("test", []string{})
	assert.Error(t, err)
}

func test_updateContext_CIEnv(t *testing.T) {
	oldCI := os.Getenv("CI")
	t.Setenv("CI", "true")
	defer t.Setenv("CI", oldCI)

	ctx := createDefaultContext("ping")
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

	err := root.updateContext("test", []string{"1.1.1.1"})
	assert.Equal(t, "test", ctx.Cmd)
	assert.Equal(t, "1.1.1.1", ctx.Target)
	assert.Equal(t, "world", ctx.From)
	assert.True(t, ctx.CIMode)
	assert.NoError(t, err)
}

func test_updateContext_TargetIsNotAHostname(t *testing.T) {
	ctx := createDefaultContext("ping")
	ctx.Ipv4 = true
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

	err := root.updateContext("ping", []string{"1.1.1.1"})
	assert.EqualError(t, err, ErrTargetIPVersionNotAllowed.Error())

	ctx.Ipv4 = false
	ctx.Ipv6 = true
	err = root.updateContext("ping", []string{"1.1.1.1"})
	assert.EqualError(t, err, ErrTargetIPVersionNotAllowed.Error())
}

func test_updateContext_ResolverIsNotAHostname(t *testing.T) {
	ctx := createDefaultContext("dns")
	ctx.Ipv4 = true
	printer := view.NewPrinter(nil, nil, nil)
	root := NewRoot(printer, ctx, nil, nil, nil, nil)

	err := root.updateContext("dns", []string{"example.com", "@1.1.1.1"})
	assert.EqualError(t, err, ErrResolverIPVersionNotAllowed.Error())

	ctx.Ipv4 = false
	ctx.Ipv6 = true
	err = root.updateContext("dns", []string{"example.com", "@1.1.1.1"})
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

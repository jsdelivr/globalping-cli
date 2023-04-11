package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTargetQuery_Simple(t *testing.T) {
	cmd := "ping"
	args := []string{"example.com"}

	q, err := ParseTargetQuery(cmd, args)
	assert.NoError(t, err)

	assert.Equal(t, TargetQuery{Target: "example.com", From: ""}, *q)
}

func TestParseTargetQuery_SimpleWithResolver(t *testing.T) {
	cmd := "dns"
	args := []string{"example.com", "@1.1.1.1"}

	q, err := ParseTargetQuery(cmd, args)
	assert.NoError(t, err)

	assert.Equal(t, TargetQuery{Target: "example.com", From: "", Resolver: "1.1.1.1"}, *q)
}

func TestParseTargetQuery_ResolverNotAllowed(t *testing.T) {
	cmd := "ping"
	args := []string{"example.com", "@1.1.1.1"}

	_, err := ParseTargetQuery(cmd, args)
	assert.ErrorContains(t, err, "does not accept a resolver argument")
}

func TestParseTargetQuery_TargetFromX(t *testing.T) {
	cmd := "ping"
	args := []string{"example.com", "from", "London"}

	q, err := ParseTargetQuery(cmd, args)
	assert.NoError(t, err)

	assert.Equal(t, TargetQuery{Target: "example.com", From: "London"}, *q)
}

func TestParseTargetQuery_TargetFromXWithResolver(t *testing.T) {
	cmd := "http"
	args := []string{"example.com", "from", "London", "@1.1.1.1"}

	q, err := ParseTargetQuery(cmd, args)
	assert.NoError(t, err)

	assert.Equal(t, TargetQuery{Target: "example.com", From: "London", Resolver: "1.1.1.1"}, *q)
}

func TestFindAndRemoveResolver_SimpleNoResolver(t *testing.T) {
	args := []string{"example.com"}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	assert.Equal(t, "", resolver)
	assert.Equal(t, args, argsWithoutResolver)
}

func TestFindAndRemoveResolver_NoResolver(t *testing.T) {
	args := []string{"example.com", "from", "London"}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	assert.Equal(t, "", resolver)
	assert.Equal(t, args, argsWithoutResolver)
}

func TestFindAndRemoveResolver_ResolverAndFrom(t *testing.T) {
	args := []string{"example.com", "@1.1.1.1", "from", "London"}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	assert.Equal(t, "1.1.1.1", resolver)
	assert.Equal(t, []string{"example.com", "from", "London"}, argsWithoutResolver)
}

func TestFindAndRemoveResolver_ResolverOnly(t *testing.T) {
	args := []string{"example.com", "@1.1.1.1"}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	assert.Equal(t, "1.1.1.1", resolver)
	assert.Equal(t, []string{"example.com"}, argsWithoutResolver)
}

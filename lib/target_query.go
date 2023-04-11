package lib

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

type TargetQuery struct {
	Target   string
	From     string
	Resolver string
}

var commandsWithResolver = []string{
	"dns",
	"http",
}

func ParseTargetQuery(cmd string, args []string) (*TargetQuery, error) {
	targetQuery := &TargetQuery{}
	// Target
	if len(args) == 0 {
		return nil, errors.New("provided target is empty")
	}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)

	if resolver != "" {
		if !slices.Contains(commandsWithResolver, cmd) {
			return nil, fmt.Errorf("command %s does not accept a resolver argument. @%s was provided", cmd, resolver)
		}

		targetQuery.Resolver = resolver
	}

	targetQuery.Target = argsWithoutResolver[0]

	if len(argsWithoutResolver) > 1 {
		if argsWithoutResolver[1] == "from" {
			targetQuery.From = strings.TrimSpace(strings.Join(argsWithoutResolver[2:], " "))
		} else {
			return nil, errors.New("invalid command format")
		}
	}

	return targetQuery, nil
}

func findAndRemoveResolver(args []string) (string, []string) {
	var resolver string
	resolverIndex := -1
	for i := 0; i < len(args); i++ {
		if len(args[i]) > 0 && args[i][0] == '@' {
			resolver = args[i][1:]
			resolverIndex = i
			break
		}
	}

	if resolverIndex == -1 {
		// resolver was not found
		return "", args
	}

	argsClone := slices.Clone(args)
	argsWithoutResolver := slices.Delete(argsClone, resolverIndex, resolverIndex+1)

	return resolver, argsWithoutResolver
}

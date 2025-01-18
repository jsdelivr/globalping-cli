package cmd

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/storage"
	"github.com/jsdelivr/globalping-cli/version"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/spf13/cobra"
)

var (
	ErrTargetIPVersionNotAllowed   = errors.New("ipVersion is not allowed when target is not a domain")
	ErrResolverIPVersionNotAllowed = errors.New("ipVersion is not allowed when resolver is not a domain")
)

func (r *Root) updateContext(cmd *cobra.Command, args []string) error {
	r.ctx.Cmd = cmd.CalledAs() // Get the command name

	// if the command does not have any arguments or flags, show help
	if len(os.Args) == 2 {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		cmd.Help()
		return errors.New("")
	}

	targetQuery, err := parseTargetQuery(r.ctx.Cmd, args)
	if err != nil {
		return err
	}

	r.ctx.Target = targetQuery.Target

	if targetQuery.From != "" {
		r.ctx.From = targetQuery.From
	}

	if targetQuery.Resolver != "" {
		r.ctx.Resolver = targetQuery.Resolver
	}

	if r.ctx.Ipv4 || r.ctx.Ipv6 {
		if net.ParseIP(r.ctx.Target) != nil {
			return ErrTargetIPVersionNotAllowed
		}
		if r.ctx.Resolver != "" && net.ParseIP(r.ctx.Resolver) != nil {
			return ErrResolverIPVersionNotAllowed
		}
	}

	// Check env for CI
	if os.Getenv("CI") != "" {
		r.ctx.CIMode = true
	}

	// Check if it is a terminal or being piped/redirected
	// We want to disable realtime updates if that is the case
	f, ok := r.printer.OutWriter.(*os.File)
	if ok {
		stdoutFileInfo, err := f.Stat()
		if err != nil {
			return fmt.Errorf("stdout stat failed: %s", err)
		}
		if (stdoutFileInfo.Mode() & os.ModeCharDevice) == 0 {
			// stdout is piped, run in ci mode
			r.ctx.CIMode = true
		}
	} else {
		r.ctx.CIMode = true
	}

	if r.ctx.CIMode {
		r.printer.DisableStyling()
	}

	return nil
}

func (r *Root) getLocations() ([]globalping.Locations, error) {
	fromArr := strings.Split(r.ctx.From, ",")
	if len(fromArr) == 1 {
		mId, err := r.mapFromSession(fromArr[0])
		if err != nil {
			return nil, err
		}
		if mId == "" {
			mId = strings.TrimSpace(fromArr[0])
		} else {
			r.ctx.IsLocationFromSession = true
			r.ctx.RecordToSession = false
		}
		return []globalping.Locations{{Magic: mId}}, nil
	}
	locations := make([]globalping.Locations, len(fromArr))
	for i, v := range fromArr {
		locations[i] = globalping.Locations{
			Magic: strings.TrimSpace(v),
		}
	}
	return locations, nil
}

func (r *Root) evaluateError(err error) {
	if err == nil {
		return
	}
	e, ok := err.(*globalping.MeasurementError)
	if !ok {
		return
	}
	if e.Code == globalping.StatusUnauthorizedWithTokenRefreshed {
		r.Cmd.SilenceErrors = true
		r.printer.ErrPrintln("Access token successfully refreshed. Try repeating the measurement.")
		return
	}
	if e.Code == http.StatusTooManyRequests && r.ctx.MeasurementsCreated > 0 {
		r.Cmd.SilenceErrors = true
		r.printer.ErrPrintln(r.printer.Color("> "+e.Message, view.FGBrightYellow))
		return
	}
}

type TargetQuery struct {
	Target   string
	From     string
	Resolver string
}

var commandsWithResolver = []string{
	"dns",
	"http",
}

func parseTargetQuery(cmd string, args []string) (*TargetQuery, error) {
	targetQuery := &TargetQuery{}
	if len(args) == 0 {
		return nil, errors.New("provided target is empty")
	}

	resolver, argsWithoutResolver := findAndRemoveResolver(args)
	if resolver != "" {
		// resolver was found
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
		if len(args[i]) > 0 && args[i][0] == '@' && args[i-1] != "from" {
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

// Maps a location to a measurement ID from history, if possible.
func (r *Root) mapFromSession(location string) (string, error) {
	if location == "" {
		return "", nil
	}
	if location[0] == '@' {
		index, err := strconv.Atoi(location[1:])
		if err != nil {
			return "", storage.ErrInvalidIndex
		}
		return r.storage.GetIdFromSession(index)
	}
	if location == "first" {
		return r.storage.GetIdFromSession(1)
	}
	if location == "last" || location == "previous" {
		return r.storage.GetIdFromSession(-1)
	}
	return "", nil
}

func silenceUsageOnCreateMeasurementError(err error) bool {
	e, ok := err.(*globalping.MeasurementError)
	if ok {
		switch e.Code {
		case http.StatusBadRequest:
			return false
		default:
			return true
		}
	}
	return true
}

func getUserAgent() string {
	return fmt.Sprintf("globalping-cli/v%s (https://github.com/jsdelivr/globalping-cli)", version.Version)
}

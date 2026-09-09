package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jsdelivr/globalping-cli/api"
	"github.com/jsdelivr/globalping-cli/storage"
	"github.com/jsdelivr/globalping-cli/version"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/jsdelivr/globalping-go"
	"github.com/spf13/cobra"
)

var (
	ErrTargetIPVersionNotAllowed   = errors.New("ipVersion is not allowed when target is not a domain")
	ErrResolverIPVersionNotAllowed = errors.New("ipVersion is not allowed when resolver is not a domain")
)

const (
	measurementAwaitDefaultTimeout = 45 * time.Second
	measurementAwaitTimeoutBuffer  = 10 * time.Second
)

type measurementAwaitState struct {
	id          string
	startedAt   time.Time
	maxDuration time.Duration
}

func newMeasurementAwaitState(id string, startedAt time.Time, timeout int) *measurementAwaitState {
	maxDuration := measurementAwaitDefaultTimeout

	if timeout != 0 {
		maxDuration = time.Duration(timeout)*time.Second + measurementAwaitTimeoutBuffer
	}

	return &measurementAwaitState{
		id:          id,
		startedAt:   startedAt,
		maxDuration: maxDuration,
	}
}

func (s *measurementAwaitState) checkTimeout(now time.Time) error {
	if now.Sub(s.startedAt) > s.maxDuration {
		return fmt.Errorf("timed out waiting for measurement %s to finish: %w", s.id, context.DeadlineExceeded)
	}

	return nil
}

func (r *Root) handleMeasurement(ctx context.Context, id string, opts *globalping.MeasurementCreate) (err error) {
	defer func() {
		if err != nil {
			r.Cmd.SilenceUsage = true
		}
	}()

	if !r.ctx.Table && (r.ctx.CIMode || r.ctx.ToJSON || r.ctx.ToLatency) {
		res, err := r.client.AwaitMeasurement(ctx, id)

		if err != nil {
			return err
		}

		if r.ctx.ToLatency {
			return r.viewer.OutputLatency(id, res)
		}

		if r.ctx.ToJSON {
			b, err := r.client.GetMeasurementRaw(ctx, id)

			if err != nil {
				return err
			}

			r.viewer.OutputJSON(id, b)

			return nil
		}

		if r.ctx.CIMode {
			r.viewer.OutputDefault(id, res, opts)

			return nil
		}
	}

	if r.ctx.Table {
		defer r.viewer.OutputShare()
	}

	startedAt := r.utils.Now()
	res, err := r.client.GetMeasurement(ctx, id)

	if err != nil {
		return err
	}

	awaitState := newMeasurementAwaitState(id, startedAt, res.Timeout)

	if r.ctx.Table {
		for {
			if !r.ctx.CIMode || res.Status != globalping.MeasurementStatusInProgress {
				_, err = r.viewer.OutputTable(res)

				if err != nil {
					return err
				}
			}

			if res.Status != globalping.MeasurementStatusInProgress {
				return nil
			}

			if err := awaitState.checkTimeout(r.utils.Now()); err != nil {
				return err
			}

			timer := time.NewTimer(r.ctx.APIMinInterval)

			select {
			case <-ctx.Done():
				timer.Stop()

				return ctx.Err()
			case <-timer.C:
			}

			res, err = r.client.GetMeasurement(ctx, id)

			if err != nil {
				return err
			}
		}
	}

	w, h := r.printer.GetSize()
	// Poll API until the measurement is complete
	for res.Status == globalping.MeasurementStatusInProgress {
		if err := awaitState.checkTimeout(r.utils.Now()); err != nil {
			return err
		}

		timer := time.NewTimer(r.ctx.APIMinInterval)

		select {
		case <-ctx.Done():
			timer.Stop()

			return ctx.Err()
		case <-timer.C:
		}

		res, err = r.client.GetMeasurement(ctx, id)

		if err != nil {
			return err
		}

		r.viewer.OutputLive(res, opts, w, h)
	}

	r.printer.AreaClear()

	r.viewer.OutputDefault(id, res, opts)

	return nil
}

func (r *Root) createAndHandleMeasurements(ctx context.Context, opts []*globalping.MeasurementCreate) error {
	first, err := r.createMeasurement(ctx, opts[0])

	if err != nil {
		r.evaluateError(err)

		return err
	}

	if len(opts) == 1 {
		return r.handleMeasurement(ctx, first.Id, opts[0])
	}

	opts[1].Locations = globalping.PreviousMeasurementID(first.Id)
	second, createErr := r.createMeasurement(ctx, opts[1])

	if createErr != nil {
		r.evaluateError(createErr)
		displayErr := r.handleMeasurement(ctx, first.Id, opts[0])

		if displayErr != nil && r.Cmd.SilenceErrors {
			r.printer.ErrPrintf("Error: %v\n", displayErr)
		}

		return errors.Join(createErr, displayErr)
	}

	return r.handleComparisonMeasurements(ctx, first.Id, second.Id)
}

func (r *Root) createMeasurement(ctx context.Context, opts *globalping.MeasurementCreate) (*view.HistoryItem, error) {
	res, err := r.client.CreateMeasurement(ctx, opts)

	if err != nil {
		r.Cmd.SilenceUsage = silenceUsageOnCreateMeasurementError(err)

		return nil, err
	}

	r.ctx.MeasurementsCreated++
	hm := &view.HistoryItem{
		Id:        res.ID,
		Status:    globalping.MeasurementStatusInProgress,
		StartedAt: r.utils.Now(),
	}
	r.ctx.History.Push(hm)

	if r.ctx.RecordToSession {
		r.ctx.RecordToSession = false
		err := r.storage.SaveIdToSession(res.ID)

		if err != nil {
			if r.ctx.Cmd == "ping" {
				r.printer.ErrPrintf("Warning: %s\n", err)
			} else {
				r.printer.Printf("Warning: %s\n", err)
			}
		}
	}

	return hm, nil
}

func (r *Root) comparisonRequests(first *globalping.MeasurementCreate) []*globalping.MeasurementCreate {
	requests := []*globalping.MeasurementCreate{first}

	if !r.ctx.Comparison {
		return requests
	}

	second := *first
	second.Target = r.ctx.Targets[1]
	second.Locations = nil

	return append(requests, &second)
}

func (r *Root) handleComparisonMeasurements(ctx context.Context, firstID, secondID string) (err error) {
	defer func() {
		if err != nil {
			r.Cmd.SilenceUsage = true
		}
	}()
	defer r.viewer.OutputShare()

	measurements := make([]*globalping.Measurement, 2)
	states := make([]*measurementAwaitState, 2)
	ids := []string{firstID, secondID}

	for i, id := range ids {
		startedAt := r.utils.Now()
		measurements[i], err = r.client.GetMeasurement(ctx, id)

		if err != nil {
			return err
		}

		states[i] = newMeasurementAwaitState(id, startedAt, measurements[i].Timeout)
	}

	for {
		inProgress := measurements[0].Status == globalping.MeasurementStatusInProgress ||
			measurements[1].Status == globalping.MeasurementStatusInProgress

		if !r.ctx.CIMode || !inProgress {
			if _, err = r.viewer.OutputComparisonTable(measurements[0], measurements[1]); err != nil {
				return err
			}
		}

		if !inProgress {
			return nil
		}

		for i := range measurements {
			if measurements[i].Status == globalping.MeasurementStatusInProgress {
				if err := states[i].checkTimeout(r.utils.Now()); err != nil {
					return err
				}
			}
		}

		timer := time.NewTimer(r.ctx.APIMinInterval)

		select {
		case <-ctx.Done():
			timer.Stop()

			return ctx.Err()
		case <-timer.C:
		}

		for i, id := range ids {
			if measurements[i].Status != globalping.MeasurementStatusInProgress {
				continue
			}

			measurements[i], err = r.client.GetMeasurement(ctx, id)

			if err != nil {
				return err
			}
		}
	}
}

func (r *Root) updateContext(cmd *cobra.Command, args []string) error {
	r.ctx.Cmd = cmd.CalledAs() // Get the command name

	// if the command does not have any arguments or flags, show help
	if len(os.Args) == 2 {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		if err := cmd.Help(); err != nil {
			return err
		}

		return errors.New("")
	}

	r.ctx.Protocol, _ = cmd.Flags().GetString("protocol")
	r.ctx.Protocol = strings.ToUpper(r.ctx.Protocol)
	r.ctx.Port, _ = cmd.Flags().GetUint16("port")

	targetQuery, err := parseTargetQuery(r.ctx.Cmd, args)

	if err != nil {
		return err
	}

	r.ctx.Targets, err = parseTargets(targetQuery.Target)

	if err != nil {
		return err
	}

	r.ctx.Target = r.ctx.Targets[0]
	r.ctx.Comparison = len(r.ctx.Targets) == 2

	if r.ctx.Comparison {
		switch {
		case r.ctx.ToJSON:
			return errors.New("the json flag is not supported when comparing targets")
		case r.ctx.ToLatency:
			return errors.New("the latency flag is not supported when comparing targets")
		case r.ctx.Infinite:
			return errors.New("the infinite flag is not supported when comparing targets")
		case cmd.Flags().Changed("table") && !r.ctx.Table:
			return errors.New("table output cannot be disabled when comparing targets")
		}

		r.ctx.Table = true
	}

	if r.ctx.Table {
		r.ctx.ToLatency = false
		r.ctx.ToJSON = false
	}

	if targetQuery.From != "" {
		r.ctx.From = targetQuery.From
	}

	if targetQuery.Resolver != "" {
		r.ctx.Resolver = targetQuery.Resolver
	}

	if r.ctx.Ipv4 || r.ctx.Ipv6 {
		for _, target := range r.ctx.Targets {
			if isIPTarget(r.ctx.Cmd, target) {
				return ErrTargetIPVersionNotAllowed
			}
		}

		if r.ctx.Resolver != "" && net.ParseIP(r.ctx.Resolver) != nil {
			return ErrResolverIPVersionNotAllowed
		}
	}

	if r.ctx.Limit < 1 {
		return errors.New("limit must be at least 1")
	}

	if cmd.Flags().Changed("timeout") {
		if r.ctx.Timeout < 5 || r.ctx.Timeout > 30 {
			return errors.New("timeout must be between 5 and 30 seconds")
		}
	} else {
		r.ctx.Timeout = 0
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
			return fmt.Errorf("stdout stat failed: %w", err)
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

func isIPTarget(command, target string) bool {
	if net.ParseIP(target) != nil {
		return true
	}

	if command != "http" {
		return false
	}

	urlData, err := parseUrlData(target)

	return err == nil && net.ParseIP(urlData.Host) != nil
}

func parseTargets(input string) ([]string, error) {
	parts := strings.Split(input, ",")

	if len(parts) > 2 {
		return nil, errors.New("a maximum of two targets is supported")
	}

	targets := make([]string, len(parts))
	seen := map[string]struct{}{}

	for i, part := range parts {
		target := strings.TrimSpace(part)

		if target == "" {
			return nil, errors.New("provided target is empty")
		}

		if _, ok := seen[target]; ok {
			return nil, errors.New("comparison targets must be distinct")
		}

		seen[target] = struct{}{}
		targets[i] = target
	}

	return targets, nil
}

func (r *Root) getLocations() (globalping.LocationSelection, error) {
	fromArr := strings.Split(r.ctx.From, ",")

	if len(fromArr) == 1 {
		mId, err := r.mapFromSession(fromArr[0])

		if err != nil {
			return nil, err
		}

		if mId == "" {
			return globalping.LocationOptions{{Magic: strings.TrimSpace(fromArr[0])}}, nil
		}

		r.ctx.IsLocationFromSession = true
		r.ctx.RecordToSession = false

		return globalping.PreviousMeasurementID(mId), nil
	}

	locations := make(globalping.LocationOptions, len(fromArr))

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

	var measurementErr *globalping.MeasurementError

	if !errors.As(err, &measurementErr) {
		return
	}

	if measurementErr.StatusCode == api.StatusUnauthorizedWithTokenRefreshed {
		r.Cmd.SilenceErrors = true
		r.printer.ErrPrintln("Access token successfully refreshed. Try repeating the measurement.")

		return
	}

	if measurementErr.StatusCode == http.StatusTooManyRequests && r.ctx.MeasurementsCreated > 0 {
		r.Cmd.SilenceErrors = true
		r.printer.ErrPrintln(r.printer.Color("> "+measurementErr.Message, view.FGBrightYellow))

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

	for i := range args {
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
	var measurementErr *globalping.MeasurementError

	if errors.As(err, &measurementErr) {
		switch measurementErr.StatusCode {
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

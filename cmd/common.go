package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/icza/backscanner"
	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/shirou/gopsutil/process"
)

var (
	ErrNoPreviousMeasurements = errors.New("no previous measurements found")
	ErrInvalidIndex           = errors.New("invalid index")
	ErrIndexOutOfRange        = errors.New("index out of range")
)
var (
	saveIdToSessionErr = "failed to save measurement ID: %s"
	readMeasuremetsErr = "failed to read previous measurements: %s"
)

var SESSION_PATH string

func (r *Root) updateContext(cmd string, args []string) error {
	r.ctx.Cmd = cmd // Get the command name

	targetQuery, err := parseTargetQuery(cmd, args)
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

	return nil
}

func (r *Root) getLocations() ([]globalping.Locations, error) {
	fromArr := strings.Split(r.ctx.From, ",")
	if len(fromArr) == 1 {
		mId, err := mapFromSession(fromArr[0])
		if err != nil {
			return nil, err
		}
		if mId == "" {
			mId = strings.TrimSpace(fromArr[0])
		} else {
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
func mapFromSession(location string) (string, error) {
	if location == "" {
		return "", nil
	}
	if location[0] == '@' {
		index, err := strconv.Atoi(location[1:])
		if err != nil {
			return "", ErrInvalidIndex
		}
		return getIdFromSession(index)
	}
	if location == "first" {
		return getIdFromSession(1)
	}
	if location == "last" || location == "previous" {
		return getIdFromSession(-1)
	}
	return "", nil
}

// Returns the measurement ID at the given index from the session history
func getIdFromSession(index int) (string, error) {
	if index == 0 {
		return "", ErrInvalidIndex
	}
	f, err := os.Open(getMeasurementsPath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", ErrNoPreviousMeasurements
		}
		return "", fmt.Errorf(readMeasuremetsErr, err)
	}
	defer f.Close()
	// Read ids from the end of the file
	if index < 0 {
		fStats, err := f.Stat()
		if err != nil {
			return "", fmt.Errorf(readMeasuremetsErr, err)
		}
		if fStats.Size() == 0 {
			return "", ErrNoPreviousMeasurements
		}
		scanner := backscanner.New(f, int(fStats.Size()-1)) // -1 to skip last newline
		for {
			index++
			b, _, err := scanner.LineBytes()
			if err != nil {
				if err == io.EOF {
					return "", ErrIndexOutOfRange
				}
				return "", fmt.Errorf(readMeasuremetsErr, err)
			}
			if index == 0 {
				return string(b), nil
			}
		}
	}
	// Read ids from the beginning of the file
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		index--
		if index == 0 {
			return scanner.Text(), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read previous measurements: %s", err)
	}
	return "", ErrIndexOutOfRange
}

// Saves the measurement ID to the session history
func saveIdToSession(id string) error {
	_, err := os.Stat(getSessionPath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err := os.Mkdir(getSessionPath(), 0755)
			if err != nil {
				return fmt.Errorf(saveIdToSessionErr, err)
			}
		} else {
			return fmt.Errorf(saveIdToSessionErr, err)
		}
	}
	f, err := os.OpenFile(getMeasurementsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf(saveIdToSessionErr, err)
	}
	defer f.Close()
	_, err = f.WriteString(id + "\n")
	if err != nil {
		return fmt.Errorf(saveIdToSessionErr, err)
	}
	return nil
}

func getSessionPath() string {
	if SESSION_PATH != "" {
		return SESSION_PATH
	}
	SESSION_PATH = filepath.Join(os.TempDir(), getSessionId())
	return SESSION_PATH
}

func getSessionId() string {
	p, err := process.NewProcess(int32(os.Getppid()))
	if err != nil {
		return "globalping"
	}
	// Workaround for bash.exe on Windows
	// PPID is different on each run.
	// https://cygwin.com/git/gitweb.cgi?p=newlib-cygwin.git;a=commit;h=448cf5aa4b429d5a9cebf92a0da4ab4b5b6d23fe
	if runtime.GOOS == "windows" {
		name, _ := p.Name()
		if name == "bash.exe" {
			p, err = p.Parent()
			if err != nil {
				return "globalping"
			}
		}
	}
	createTime, _ := p.CreateTime()
	return fmt.Sprintf("globalping_%d_%d", p.Pid, createTime)
}

func getMeasurementsPath() string {
	return filepath.Join(getSessionPath(), "measurements")
}

func getHistoryPath() string {
	return filepath.Join(getSessionPath(), "history")
}

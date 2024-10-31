package storage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jsdelivr/globalping-cli/utils"
	"github.com/shirou/gopsutil/process"
)

type LocalStorage struct {
	utils      utils.Utils
	configName string
	configDir  string

	tempDir           string
	sessionsDir       string
	currentSessionDir string
	config            *Config
}

func NewLocalStorage(utils utils.Utils) *LocalStorage {
	return &LocalStorage{
		utils:      utils,
		configName: "config.json",
	}
}

func (s *LocalStorage) Init(dirName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	s.configDir = filepath.Join(homeDir, dirName)
	err = os.MkdirAll(s.configDir, 0755)
	if err != nil {
		return err
	}
	s.tempDir = filepath.Join(os.TempDir(), dirName)
	s.sessionsDir = filepath.Join(s.tempDir, "sessions")
	s.currentSessionDir = filepath.Join(s.sessionsDir, getSessionId())
	err = os.MkdirAll(s.currentSessionDir, 0755)
	if err != nil {
		return err
	}
	_, err = s.LoadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			s.config = &Config{
				Profile:  "default",
				Profiles: make(map[string]*Profile),
			}
			s.SaveConfig()
		}
	}
	s.Migrate()
	return nil
}

func (s *LocalStorage) Remove() error {
	err := os.RemoveAll(s.tempDir)
	if err != nil {
		return err
	}
	err = os.RemoveAll(s.configDir)
	if err != nil {
		return err
	}
	return nil
}

func (s *LocalStorage) Cleanup() error {
	// Remove old session directories
	entries, err := os.ReadDir(s.sessionsDir)
	if err != nil {
		return err
	}
	maxEntries := 100
	l := len(entries)
	for i, e := range entries {
		name := e.Name()
		if l-i > maxEntries {
			os.RemoveAll(filepath.Join(s.sessionsDir, name))
			continue
		}
		info, _ := e.Info()
		if info.ModTime().Before(s.utils.Now().AddDate(0, 0, -7)) {
			os.RemoveAll(filepath.Join(s.sessionsDir, name))
		}
	}
	// Truncate files
	truncateFile(s.historyPath(), 1<<23)      // 8 MB
	truncateFile(s.measurementsPath(), 1<<20) // 1 MB
	return nil
}

func (s *LocalStorage) joinConfigDir(name string) (string, error) {
	return filepath.Join(s.configDir, name), nil
}

func (s *LocalStorage) joinSessionDir(name string) string {
	return filepath.Join(s.currentSessionDir, name)
}

func getSessionId() string {
	p, err := process.NewProcess(int32(os.Getppid()))
	if err != nil {
		return "session"
	}
	// Workaround for bash.exe on Windows
	// PPID is different on each run.
	// https://cygwin.com/git/gitweb.cgi?p=newlib-cygwin.git;a=commit;h=448cf5aa4b429d5a9cebf92a0da4ab4b5b6d23fe
	if runtime.GOOS == "windows" {
		name, _ := p.Name()
		if name == "bash.exe" {
			p, err = p.Parent()
			if err != nil {
				return "session"
			}
		}
	}
	createTime, _ := p.CreateTime()
	return fmt.Sprintf("%d_%d", createTime, p.Pid)
}

func truncateFile(file string, maxSize int64) error {
	f, err := os.OpenFile(file, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	fileSize := stat.Size()
	if fileSize <= maxSize {
		return nil
	}
	// Determine start position
	startPos := fileSize - maxSize
	_, err = f.Seek(startPos, io.SeekStart)
	if err != nil {
		return err
	}
	b, err := bufio.NewReader(f).ReadBytes('\n')
	if err != nil {
		return err
	}
	startPos += int64(len(b))

	// Truncate
	f.Seek(startPos, io.SeekStart)
	b, err = io.ReadAll(f)
	if err != nil {
		return err
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}
	err = f.Truncate(fileSize - startPos)
	if err != nil {
		return err
	}
	return nil
}

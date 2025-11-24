package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MigrationFunc func() error

func (s *LocalStorage) Migrate() error {
	if s.config.LastMigration >= len(s.migrations) {
		return nil
	}
	for i := s.config.LastMigration; i < len(s.migrations); i++ {
		err := s.migrations[i]()
		if err != nil {
			fmt.Printf("Warning: migration %d failed: %v", i, err)
		}
	}
	s.config.LastMigration = len(s.migrations)
	return s.SaveConfig()
}

func (s *LocalStorage) UpdateSessionDir() error {
	oldDir := filepath.Join(os.TempDir())
	entries, err := os.ReadDir(oldDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() && strings.HasPrefix(name, "globalping_") {
			info, _ := e.Info()
			if info.ModTime().Before(time.Now().AddDate(0, 0, -7)) {
				os.RemoveAll(filepath.Join(oldDir, name))
				continue
			}
			parts := strings.Split(name, "_")
			newName := parts[2] + "_" + parts[1]
			err := os.Rename(filepath.Join(oldDir, name, "measurements"), filepath.Join(s.sessionsDir, newName, measurementsFileName))
			if err != nil {
				return err
			}
			err = os.Rename(filepath.Join(oldDir, name, "history"), filepath.Join(s.sessionsDir, newName, historyFileName))
			if err != nil {
				return err
			}
			err = os.RemoveAll(filepath.Join(oldDir, name))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *LocalStorage) MoveSessionsToUserDir() error {
	// Old location: TempDir() + dirName + "sessions"
	// New location: TempDir() + dirName + userId + "sessions"
	oldSessionsDir := filepath.Join(filepath.Dir(s.tempDir), "sessions")

	if _, err := os.Stat(oldSessionsDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(oldSessionsDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(s.sessionsDir, 0755)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		oldPath := filepath.Join(oldSessionsDir, e.Name())
		newPath := filepath.Join(s.sessionsDir, e.Name())

		if _, err := os.Stat(newPath); err == nil {
			continue
		}

		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Printf("Warning: failed to move session %s: %v\n", e.Name(), err)
			continue
		}
	}

	os.RemoveAll(oldSessionsDir)

	return nil
}

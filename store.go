package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Message struct {
	Text     string    `json:"text"`
	Time     time.Time `json:"time"`
	NoHyphen bool      `json:"no_hyphen,omitempty"`
}

type Store struct {
	Messages  []Message `json:"messages"`
	MainIndex int       `json:"main_index"`
}

func storePath(gitDir string) string {
	return filepath.Join(gitDir, "jot", "log.json")
}

func loadStore(gitDir string) (Store, error) {
	path := storePath(gitDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			legacy := filepath.Join(gitDir, "gut", "log.json")
			data, err = os.ReadFile(legacy)
			if err != nil {
				if os.IsNotExist(err) {
					return Store{MainIndex: 0}, nil
				}
				return Store{}, fmt.Errorf("read store: %w", err)
			}
		} else {
			return Store{}, fmt.Errorf("read store: %w", err)
		}
	}

	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return Store{}, fmt.Errorf("parse store: %w", err)
	}
	return store, nil
}

func saveStore(gitDir string, store Store) error {
	path := storePath(gitDir)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create store directory: %w", err)
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("encode store: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "log-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return fmt.Errorf("rename store file: %w", err)
	}
	return nil
}

func clearStore(gitDir string) error {
	return saveStore(gitDir, Store{MainIndex: 0})
}

package main

import (
	"os"
	"path/filepath"
	"strings"
)

type ConfigOpts struct {
	Symbol   string
	Interval int
	Backfill int
}

// find the filename = .env
// Root => base of the repo
func loadDotEnv(path string) error {
	data, err := os.ReadFile(filepath.Join(path, ".env"))
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if _, set := os.LookupEnv(k); set {
			continue
		}
		os.Setenv(strings.TrimSpace(k), strings.TrimSpace(v))
	}
	return nil
}

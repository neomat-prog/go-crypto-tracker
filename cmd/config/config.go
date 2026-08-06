package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ConfigOpts struct {
	Symbol   string
	Interval string
	Backfill int
}

var validIntervals = map[string]struct{}{
	"1m": {}, "3m": {}, "5m": {}, "15m": {}, "30m": {},
	"1h": {}, "2h": {}, "4h": {}, "6h": {}, "8h": {}, "12h": {},
	"1d": {}, "3d": {}, "1w": {}, "1M": {},
}

func LoadDotEnv(path string) (ConfigOpts, error) {
	values, _, err := readDotEnv(path)
	if err != nil {
		return ConfigOpts{}, err
	}

	get := func(key string) string {
		if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
		return strings.TrimSpace(values[key])
	}

	cfg := ConfigOpts{
		Symbol:   strings.ToUpper(get("SYMBOL")),
		Interval: get("INTERVAL"),
		Backfill: 100, // default
	}

	if raw := get("BACKFILL"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return ConfigOpts{}, fmt.Errorf("BACKFILL %q is not a number: %w", raw, err)
		}
		cfg.Backfill = n
	}

	if err := cfg.Validate(); err != nil {
		return ConfigOpts{}, err
	}
	return cfg, nil
}

func (c ConfigOpts) Validate() error {
	var errs []error

	if strings.TrimSpace(c.Symbol) == "" {
		errs = append(errs, errors.New("SYMBOL is required for live market display"))
	}
	if _, ok := validIntervals[c.Interval]; !ok {
		errs = append(errs, fmt.Errorf("INTERVAL %q is not valid", c.Interval))
	}
	if c.Backfill < 0 || c.Backfill > 1000 {
		errs = append(errs, fmt.Errorf("BACKFILL must be 0..1000, got %d", c.Backfill))
	}

	return errors.Join(errs...)
}

func readDotEnv(path string) (map[string]string, string, error) {
	resolvedPath, exists, err := resolveDotEnvPath(path)
	if err != nil {
		return nil, "", err
	}
	if !exists {
		return map[string]string{}, "", nil
	}

	file, err := os.Open(resolvedPath)
	if err != nil {
		return nil, "", fmt.Errorf("open %s: %w", resolvedPath, err)
	}
	defer file.Close()

	values := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, "", fmt.Errorf("invalid .env line: %q", line)
		}

		key = strings.TrimSpace(strings.TrimPrefix(key, "export "))
		if key == "" {
			return nil, "", fmt.Errorf("invalid .env line: %q", line)
		}

		value = strings.Trim(strings.TrimSpace(value), `"'`)
		values[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, "", fmt.Errorf("read %s: %w", resolvedPath, err)
	}

	return values, filepath.Dir(resolvedPath), nil
}

func resolveDotEnvPath(path string) (string, bool, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", false, nil
	}

	if filepath.IsAbs(path) {
		if _, err := os.Stat(path); err == nil {
			return filepath.Clean(path), true, nil
		} else if os.IsNotExist(err) {
			return "", false, nil
		} else {
			return "", false, fmt.Errorf("stat %s: %w", path, err)
		}
	}

	if filepath.Base(path) != path {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", false, fmt.Errorf("resolve %s: %w", path, err)
		}

		if _, err := os.Stat(abs); err == nil {
			return abs, true, nil
		} else if os.IsNotExist(err) {
			return "", false, nil
		} else {
			return "", false, fmt.Errorf("stat %s: %w", abs, err)
		}
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("getwd: %w", err)
	}

	for {
		candidate := filepath.Join(dir, path)

		if _, err := os.Stat(candidate); err == nil {
			abs, absErr := filepath.Abs(candidate)
			if absErr != nil {
				return "", false, fmt.Errorf("resolve %s: %w", candidate, absErr)
			}
			return abs, true, nil
		} else if !os.IsNotExist(err) {
			return "", false, fmt.Errorf("stat %s: %w", candidate, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", false, nil
}

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeEnv(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	return path
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"SYMBOL", "INTERVAL", "BACKFILL"} {
		t.Setenv(k, "")
	}
}

func TestLoadDotEnv(t *testing.T) {
	clearEnv(t)

	path := writeEnv(t, `
	SYMBOL="btcusdt"
	INTERVAL="1m"
	BACKFILL="100
	`)

	cfg, err := loadDotEnv(path)
	if err != nil {
		t.Fatalf("loadDotEnv: %v", err)
	}
	if cfg.Symbol != "BTCUSDT" {
		t.Errorf("Symbol = %q, want BTCUSDT", cfg.Symbol)
	}
	if cfg.Interval != "1m" {
		t.Errorf("Interval = %q, want 1m", cfg.Interval)
	}
	if cfg.Backfill != 100 {
		t.Errorf("Backfill = %d, want 100", cfg.Backfill)
	}
}

func TestLoadDotEnvErrors(t *testing.T) {
	cases := []struct {
		name     string
		contents string
		wantIn   string
	}{
		{"missing symbol", "INTERVAL=\"1m\"\n", "SYMBOL"},
		{"bad interval", "SYMBOL=\"BTCUSDT\"\nINTERVAL=\"7m\"\n", "INTERVAL"},
		{"backfill not a number", "SYMBOL=\"BTCUSDT\"\nINTERVAL=\"1m\"\nBACKFILL=\"lots\"\n", "BACKFILL"},
		{"backfill out of range", "SYMBOL=\"BTCUSDT\"\nINTERVAL=\"1m\"\nBACKFILL=\"5000\"\n", "BACKFILL"},
		{"malformed line", "SYMBOL\n", "invalid .env line"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clearEnv(t)

			_, err := loadDotEnv(writeEnv(t, tc.contents))
			if err == nil {
				t.Fatal("want error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantIn) {
				t.Errorf("error %q does not mention %q", err, tc.wantIn)
			}
		})
	}
}

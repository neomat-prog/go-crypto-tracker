package main

import (
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	if err := loadDotEnv("../.."); err != nil {
		t.Fatalf("repo root .env: %v", err)
	}
}

package config

import "testing"

func TestConfigValidation(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatalf("Default().Validate() error = %v", err)
	}
	if err := (Config{}).Validate(); err == nil {
		t.Fatal("empty API key was accepted")
	}
	if err := (Config{APIKey: " secret "}).Validate(); err == nil {
		t.Fatal("API key with surrounding whitespace was accepted")
	}
}

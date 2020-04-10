package main

import "testing"

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig("testdata/config.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	if cfg.Port != 1234 {
		t.Errorf("cfg.Port = %v, expected %v", cfg.Port, 1234)
	}
}

// When an error occurs, the default values should be used.
func TestLoadConfigNotExist(t *testing.T) {
	cfg, err := LoadConfig("testdata/notexist.yaml")
	if err == nil {
		t.Errorf("err = nil, expected not nil")
	}
	if cfg.Port != 8080 {
		t.Errorf("cfg.Port = %v, expected %v", cfg.Port, 8080)
	}
}

// Unset configuration entries should be set to their default value.
func TestLoadConfigPartial(t *testing.T) {
	cfg, err := LoadConfig("testdata/config_empty.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	if cfg.Port != 8080 {
		t.Errorf("cfg.Port = %v, expected %v", cfg.Port, 8080)
	}
}

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

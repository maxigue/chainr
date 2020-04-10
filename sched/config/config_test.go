package config

import "testing"

func TestLoad(t *testing.T) {
	cfg, err := Load("testdata/config.yaml")
	if err != nil {
		t.Fatal(err.Error())
	}
	if cfg.Port != 1234 {
		t.Errorf("cfg.Port = %v, expected %v", cfg.Port, 1234)
	}
}

// When an error occurs, the default values should be used.
func TestLoadNotExist(t *testing.T) {
	cfg, err := Load("testdata/notexist.yaml")
	if err == nil {
		t.Fatal("err = nil, expected not nil")
	}
	if cfg.Port != 8080 {
		t.Errorf("cfg.Port = %v, expected %v", cfg.Port, 8080)
	}
}

// Unset configuration entries should be set to their default value.
func TestLoadPartial(t *testing.T) {
	cfg, err := Load("testdata/config_empty.yaml")
	if err != nil {
		t.Fatal(err.Error())
	}
	if cfg.Port != 8080 {
		t.Errorf("cfg.Port = %v, expected %v", cfg.Port, 8080)
	}
}

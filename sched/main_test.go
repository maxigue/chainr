package main

import (
	"testing"

	"io/ioutil"
	"log"
	"os"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func TestAddrDefault(t *testing.T) {
	addr := addr()
	if addr != ":8080" {
		t.Errorf("addr = %v, expected :8080", addr)
	}
}

func TestAddrEnv(t *testing.T) {
	if err := os.Setenv("PORT", "1234"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("PORT")

	addr := addr()
	if addr != ":1234" {
		t.Errorf("addr = %v, expected :1234", addr)
	}
}

func TestAddrInvalidPort(t *testing.T) {
	if err := os.Setenv("PORT", "test"); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("PORT")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Addr did not panic")
		}
	}()

	_ = addr()
}

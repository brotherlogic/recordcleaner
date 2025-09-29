package main

import (
	"testing"
)

func TestInit(t *testing.T) {
	s := Init()
	if s == nil {
		t.Errorf("Init() returned a nil server")
	}

	if s.lastUpdate == nil {
		t.Errorf("Init() did not initialize lastUpdate map")
	}
}

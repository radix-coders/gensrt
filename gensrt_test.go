package gensrt

import (
	"testing"
)

func TestSuccessCase(t *testing.T) {
	cfg, err := NewConfig("test.wav", "key.json")
	err = cfg.ProcessRequest()
	if err != nil {
		t.Errorf("ProcessRequest Err:%+v", err)
	}
}

func TestFailureCase1(t *testing.T) {
	_, err := NewConfig("", "key.json")
	if err == nil {
		t.Errorf("NewConfig Err:%+v", err)
	}
}

func TestFailureCase2(t *testing.T) {
	_, err := NewConfig("test.wav", "")
	if err == nil {
		t.Errorf("NewConfig Err:%+v", err)
	}
}

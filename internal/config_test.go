package scribe

import (
	"testing"
)

func TestBoost_InitLogger_Debug(t *testing.T) {
	InitLogger(true)
	if Logger == nil {
		t.Fatal("InitLogger(true): Logger is nil")
	}
}

func TestBoost_InitLogger_Info(t *testing.T) {
	InitLogger(false)
	if Logger == nil {
		t.Fatal("InitLogger(false): Logger is nil")
	}
}

func TestBoost_InitLoggerCLI_Debug(t *testing.T) {
	InitLoggerCLI(true)
	if Logger == nil {
		t.Fatal("InitLoggerCLI(true): Logger is nil")
	}
}

func TestBoost_InitLoggerCLI_NondDebug(t *testing.T) {
	InitLoggerCLI(false)
	if Logger == nil {
		t.Fatal("InitLoggerCLI(false): Logger is nil")
	}
}

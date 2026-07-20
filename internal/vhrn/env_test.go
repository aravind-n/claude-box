package vhrn

import (
	"reflect"
	"testing"
)

func TestTerminalEnvDefaults(t *testing.T) {
	t.Setenv("TERM", "")
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM_PROGRAM_VERSION", "")
	got := terminalEnv()
	want := []string{"--env", "TERM=xterm-256color"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("terminalEnv() = %v, want %v", got, want)
	}
}

func TestTerminalEnvForwards(t *testing.T) {
	t.Setenv("TERM", "screen-256color")
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TERM_PROGRAM", "Apple_Terminal")
	t.Setenv("TERM_PROGRAM_VERSION", "")
	got := terminalEnv()
	want := []string{
		"--env", "TERM=screen-256color",
		"--env", "COLORTERM=truecolor",
		"--env", "TERM_PROGRAM=Apple_Terminal",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("terminalEnv() = %v, want %v", got, want)
	}
}

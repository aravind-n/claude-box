package vhrn

import (
	"reflect"
	"testing"
)

func TestParseRunFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		openNet bool
		allow   []string
		rest    []string
		wantErr bool
	}{
		{name: "empty", args: nil, rest: nil},
		{name: "agent flags pass through", args: []string{"--model", "opus"}, rest: []string{"--model", "opus"}},
		{name: "open-net then dashdash", args: []string{"--open-net", "--", "--help"}, openNet: true, rest: []string{"--help"}},
		{name: "allow comma list", args: []string{"--allow", "a.com,b.com", "arg"}, allow: []string{"a.com", "b.com"}, rest: []string{"arg"}},
		{name: "allow equals form", args: []string{"--allow=x.com"}, allow: []string{"x.com"}},
		{name: "repeated allow", args: []string{"--allow", "a.com", "--allow", "b.com"}, allow: []string{"a.com", "b.com"}},
		{name: "allow missing value", args: []string{"--allow"}, wantErr: true},
		{name: "bare dashdash", args: []string{"--"}, rest: nil},
		{name: "first unknown stops parsing", args: []string{"positional", "--open-net"}, rest: []string{"positional", "--open-net"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := parseRunFlags(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if f.openNet != tt.openNet {
				t.Errorf("openNet = %v, want %v", f.openNet, tt.openNet)
			}
			if !reflect.DeepEqual(f.extraAllow, tt.allow) {
				t.Errorf("extraAllow = %v, want %v", f.extraAllow, tt.allow)
			}
			if !reflect.DeepEqual(f.rest, tt.rest) {
				t.Errorf("rest = %v, want %v", f.rest, tt.rest)
			}
		})
	}
}

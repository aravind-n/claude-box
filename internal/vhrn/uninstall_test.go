package vhrn

import (
	"reflect"
	"testing"
)

func TestRemoveImageArgv(t *testing.T) {
	if got, want := removeImageArgv("docker", "vhrn-claude"), []string{"image", "rm", "vhrn-claude"}; !reflect.DeepEqual(got, want) {
		t.Errorf("docker argv = %v, want %v", got, want)
	}
	if got, want := removeImageArgv("container", "vhrn-claude"), []string{"image", "delete", "vhrn-claude"}; !reflect.DeepEqual(got, want) {
		t.Errorf("container argv = %v, want %v", got, want)
	}
}

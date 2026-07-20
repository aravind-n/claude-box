package vhrn

import (
	"fmt"
	"os"
	"os/exec"
)

// detectEngine mirrors vhrn.sh: an explicit VHRN_ENGINE (then ENGINE) wins, else
// auto-detect `container` first, then `docker` — matching the Makefile so build
// and run agree on the same engine.
func detectEngine() (string, error) {
	engine := os.Getenv("VHRN_ENGINE")
	if engine == "" {
		engine = os.Getenv("ENGINE")
	}
	if engine == "" {
		switch {
		case lookPath("container"):
			engine = "container"
		case lookPath("docker"):
			engine = "docker"
		default:
			return "", fmt.Errorf("no container engine found; install Apple container or Docker")
		}
	}
	if !lookPath(engine) {
		return "", fmt.Errorf("engine %q not found", engine)
	}
	return engine, nil
}

func lookPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

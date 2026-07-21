package vhrn

import "strings"

// Image names and the registry release images are pulled from. VHRN_REGISTRY
// overrides the registry (a fork, or a local mirror).
const (
	baseImageName   = "vhrn-base"
	proxyImageName  = "vhrn-proxy"
	defaultRegistry = "ghcr.io/aravind-n"
	// localVersion marks a make-built image used as-is (bare name, no registry),
	// rather than one pulled from the registry.
	localVersion = "local"
)

func registryBase() string { return envOr("VHRN_REGISTRY", defaultRegistry) }

// parseHarnessArg splits "claude" or "claude@v0.2.0" into name and version,
// defaulting to "latest" when no @tag is given.
func parseHarnessArg(arg string) (name, version string) {
	if i := strings.IndexByte(arg, '@'); i >= 0 {
		if name, version = arg[:i], arg[i+1:]; version == "" {
			version = "latest"
		}
		return name, version
	}
	return arg, "latest"
}

// harnessImageRef is the image to run for a harness at an installed version: the
// bare local image for a make-built install, else the versioned registry ref. The
// proxy is pinned to the same version, so a box and its proxy are always a set.
func harnessImageRef(h Harness, version string) string {
	if version == localVersion {
		return h.Image
	}
	return registryBase() + "/" + h.Image + ":" + version
}

// proxyImageRef pins the egress proxy to the same version as the harness it serves.
func proxyImageRef(version string) string {
	if version == localVersion {
		return proxyImageName
	}
	return registryBase() + "/" + proxyImageName + ":" + version
}

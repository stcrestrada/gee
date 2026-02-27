package tui

import "os/exec"

// Capabilities records which external CLIs are available.
type Capabilities struct {
	HasGH   bool
	HasGLab bool
}

// DetectCapabilities checks the PATH for gh and glab.
func DetectCapabilities() Capabilities {
	_, ghErr := exec.LookPath("gh")
	_, glabErr := exec.LookPath("glab")
	return Capabilities{
		HasGH:   ghErr == nil,
		HasGLab: glabErr == nil,
	}
}

// DiscoveryProvider returns "gh", "glab", or "" based on what's available.
func DiscoveryProvider(caps Capabilities) string {
	if caps.HasGH {
		return "gh"
	}
	if caps.HasGLab {
		return "glab"
	}
	return ""
}

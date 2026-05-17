package cmd

import "os"

// ActorEnvChain is the environment variable fallback chain for automatic actor
// resolution. Earlier entries take precedence.
var ActorEnvChain = []string{"TL_ACTOR", "ACTOR_NAME", "BEADS_ACTOR"}

// DefaultDetectActor is the default auto-detection function, exported so
// tests can restore DetectedActor after overriding it.
var DefaultDetectActor = defaultDetectActor

// DetectedActor is called when no explicit actor is provided via CLI flag or
// environment variable. Tests may override this to simulate agent detection.
var DetectedActor = DefaultDetectActor

// ResolveActor returns the actor identity using this priority:
//  1. CLI --actor flag (when non-empty)
//  2. Environment variables in ActorEnvChain order
//  3. Auto-detection via DetectedActor
func ResolveActor(cliFlag string) string {
	if cliFlag != "" {
		return cliFlag
	}
	for _, env := range ActorEnvChain {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	return DetectedActor()
}

func defaultDetectActor() string {
	// Known agent markers.
	if v := os.Getenv("CLAUDE_CODE_SESSION_ID"); v != "" {
		return "claude-code"
	}
	if _, err := os.Stat(".codex"); err == nil {
		return "codex"
	}
	if v := os.Getenv("PI_AGENT_ID"); v != "" {
		return "pi"
	}
	// Fallback: use hostname.
	if host, err := os.Hostname(); err == nil && host != "" {
		return host
	}
	return "unknown"
}

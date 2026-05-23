package bdd

import (
	"fmt"
	"os"
	"time"

	"github.com/cucumber/godog"

	"github.com/aholbreich/tl/cmd"
)

// --- actor.feature support ------------------------------------------------

func initializeActorSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^environment variable "([^"]*)" is "([^"]*)"$`, w.setEnv)
	ctx.Step(`^the detected agent is "([^"]*)"$`, w.setDetectedAgent)
	ctx.Step(`^the claim expiry for "([^"]*)" is extended$`, w.claimExpiryIsExtended)
}

func (w *world) setEnv(key, value string) error {
	w.envOverrides = append(w.envOverrides, key)
	return os.Setenv(key, value)
}

func (w *world) setDetectedAgent(agent string) error {
	cmd.DetectedActor = func() string { return agent }
	return nil
}

func (w *world) claimExpiryIsExtended(id string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Claim.ExpiresAt == nil {
		return fmt.Errorf("task %s has no claim expiry", id)
	}
	// Fixture tasks set UpdatedAt to fixtureTime (2026-05-16T12:00Z).
	// A successful renewal bumps UpdatedAt to time.Now().
	if t.UpdatedAt.Equal(fixtureTime) {
		return fmt.Errorf("task %s was not renewed (updated_at still at fixture time)", id)
	}
	if !t.Claim.ExpiresAt.After(time.Now().UTC()) {
		return fmt.Errorf("task %s claim expiry %v is not in the future", id, t.Claim.ExpiresAt)
	}
	return nil
}

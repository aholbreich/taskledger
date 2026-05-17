package bdd

import (
	"bytes"
	"fmt"
	"github.com/cucumber/godog"
	"time"

	"encoding/json"
	"github.com/aholbreich/taskledger/internal/task"
)

// --- claim.feature support -----------------------------------------------

func initializeClaimSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^a ready task "([^"]*)" titled "([^"]*)"$`, w.readyTaskTitled)
	ctx.Step(`^a ready task "([^"]*)"$`, w.readyTask)
	ctx.Step(`^a task "([^"]*)" claimed by "([^"]*)" with an active lease$`, w.taskClaimedByWithActiveLease)
	ctx.Step(`^a task "([^"]*)" claimed by "([^"]*)"$`, w.taskClaimedByWithActiveLease)
	ctx.Step(`^a task "([^"]*)" with status "([^"]*)"$`, w.taskWithStatus)
	ctx.Step(`^"([^"]*)" is claimed by "([^"]*)"$`, w.taskIsClaimedBy)
	ctx.Step(`^"([^"]*)" is still claimed by "([^"]*)"$`, w.taskIsClaimedBy)
	ctx.Step(`^"([^"]*)" is no longer claimed by "([^"]*)"$`, w.taskIsNoLongerClaimedBy)
	ctx.Step(`^"([^"]*)" has status "([^"]*)"$`, w.taskHasSpecificStatus)
	ctx.Step(`^"([^"]*)" still has status "([^"]*)"$`, w.taskHasSpecificStatus)
	ctx.Step(`^"([^"]*)" has a non-empty claim expiry$`, w.taskHasNonEmptyClaimExpiry)
	ctx.Step(`^"([^"]*)" is not claimed$`, w.taskIsNotClaimed)
	ctx.Step(`^"([^"]*)" was closed by "([^"]*)"$`, w.taskWasClosedBy)
	ctx.Step(`^"([^"]*)" does not have status "([^"]*)"$`, w.taskDoesNotHaveStatus)
	ctx.Step(`^the command reports the task is blocked$`, w.outputReportsTaskBlocked)
	ctx.Step(`^the command reports the claim is held by a different actor$`, w.outputReportsClaimHeldByDifferentActor)
	ctx.Step(`^the command reports the task is already closed$`, w.outputReportsTaskAlreadyClosed)
	ctx.Step(`^the JSON output contains actor "([^"]*)"$`, w.jsonClaimActor)
	ctx.Step(`^the JSON output contains a claim expiry (\d+) minutes after the claim time$`, w.jsonClaimExpiryAfter)
}

func (w *world) readyTaskTitled(id, title string) error {
	return writeFixtureTask(&task.Task{
		ID:        id,
		Title:     title,
		Status:    "open",
		Priority:  "medium",
		CreatedAt: fixtureTime,
		UpdatedAt: fixtureTime,
		CreatedBy: "human",
		DependsOn: []string{},
		Tags:      []string{},
	})
}

func (w *world) readyTask(id string) error {
	return w.readyTaskTitled(id, id)
}

func (w *world) taskClaimedByWithActiveLease(id, actor string) error {
	now := time.Now().UTC().Truncate(time.Second)
	expires := now.Add(1 * time.Hour)
	a := actor
	return writeFixtureTask(&task.Task{
		ID:        id,
		Title:     id,
		Status:    "in_progress",
		Priority:  "medium",
		CreatedAt: fixtureTime,
		UpdatedAt: fixtureTime,
		CreatedBy: "human",
		DependsOn: []string{},
		Tags:      []string{},
		Claim: task.Claim{
			Actor:       &a,
			ClaimedAt:   &now,
			ExpiresAt:   &expires,
			HeartbeatAt: &now,
		},
	})
}

func (w *world) taskWithStatus(id, status string) error {
	t := &task.Task{
		ID:        id,
		Title:     id,
		Status:    status,
		Priority:  "medium",
		CreatedAt: fixtureTime,
		UpdatedAt: fixtureTime,
		CreatedBy: "human",
		DependsOn: []string{},
		Tags:      []string{},
	}
	return writeFixtureTask(t)
}

func (w *world) taskIsClaimedBy(id, actor string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Claim.Actor == nil {
		return fmt.Errorf("task %s has no claim", id)
	}
	if *t.Claim.Actor != actor {
		return fmt.Errorf("task %s claimed by %q, expected %q", id, *t.Claim.Actor, actor)
	}
	return nil
}

func (w *world) taskIsNoLongerClaimedBy(id, actor string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Claim.Actor != nil && *t.Claim.Actor == actor {
		return fmt.Errorf("task %s is still claimed by %q", id, actor)
	}
	return nil
}

func (w *world) taskHasSpecificStatus(id, status string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Status != status {
		return fmt.Errorf("task %s status is %q, expected %q", id, t.Status, status)
	}
	return nil
}

func (w *world) taskHasNonEmptyClaimExpiry(id string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Claim.ExpiresAt == nil {
		return fmt.Errorf("task %s has no claim expiry", id)
	}
	if t.Claim.ExpiresAt.IsZero() {
		return fmt.Errorf("task %s claim expiry is zero", id)
	}
	return nil
}

func (w *world) taskIsNotClaimed(id string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Claim.Actor != nil {
		return fmt.Errorf("task %s is claimed by %q, expected none", id, *t.Claim.Actor)
	}
	return nil
}

func (w *world) taskWasClosedBy(id, actor string) error {
	return assertEventRecordedBy("closed", id, actor)
}

func (w *world) taskDoesNotHaveStatus(id, status string) error {
	t, err := loadFixtureTask(id)
	if err != nil {
		return err
	}
	if t.Status == status {
		return fmt.Errorf("task %s status is %q, expected a different status", id, status)
	}
	return nil
}

func (w *world) outputReportsTaskBlocked() error {
	return w.outputContainsAll("blocked")
}

func (w *world) outputReportsClaimHeldByDifferentActor() error {
	return w.outputContainsAll("claimed", "different")
}

func (w *world) outputReportsTaskAlreadyClosed() error {
	return w.outputContainsAll("already", "closed")
}

func (w *world) jsonClaimActor(expected string) error {
	dec := json.NewDecoder(bytes.NewReader(w.stdout.Bytes()))

	// Try single object.
	var single struct {
		Claim struct {
			Actor *string `json:"actor"`
		} `json:"claim"`
	}
	if err := json.Unmarshal(w.stdout.Bytes(), &single); err == nil && single.Claim.Actor != nil {
		if *single.Claim.Actor != expected {
			return fmt.Errorf("JSON claim actor = %v, expected %q", single.Claim.Actor, expected)
		}
		return nil
	}

	// Try array.
	dec = json.NewDecoder(bytes.NewReader(w.stdout.Bytes()))
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('[') {
		return fmt.Errorf("stdout is neither JSON object nor array; got: %s", w.stdout.String())
	}
	for dec.More() {
		var t task.Task
		if err := dec.Decode(&t); err != nil {
			return fmt.Errorf("parse JSON array element: %w", err)
		}
		if t.Claim.Actor != nil && *t.Claim.Actor == expected {
			return nil
		}
	}
	return fmt.Errorf("JSON array does not contain a claim by %q; got: %s", expected, w.stdout.String())
}

func (w *world) jsonClaimExpiryAfter(minutes int) error {
	var data struct {
		Claim struct {
			ClaimedAt time.Time `json:"claimed_at"`
			ExpiresAt time.Time `json:"expires_at"`
		} `json:"claim"`
	}
	if err := json.Unmarshal(w.stdout.Bytes(), &data); err != nil {
		return fmt.Errorf("stdout is not JSON (%v); got: %s", err, w.stdout.String())
	}
	diff := data.Claim.ExpiresAt.Sub(data.Claim.ClaimedAt)
	expected := time.Duration(minutes) * time.Minute
	if diff != expected {
		return fmt.Errorf("claim expiry is %v after claim time, expected %d minutes (%v)", diff, minutes, expected)
	}
	return nil
}

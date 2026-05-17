package bdd

import (
	"fmt"
	"strings"
	"time"

	"encoding/json"
	"github.com/aholbreich/taskledger/internal/task"
	"github.com/cucumber/godog"
)

// --- list.feature support -------------------------------------------------

func initializeListSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^the following tasks exist:$`, w.followingTasksExist)
	ctx.Step(`^the output lists "([^"]*)"$`, w.outputListsTaskID)
	ctx.Step(`^the output lists "([^"]*)" with status "([^"]*)" and title "([^"]*)"$`, w.outputListsTask)
	ctx.Step(`^the output lists "([^"]*)" with status "([^"]*)", priority "([^"]*)", claimed by "([^"]*)", and title "([^"]*)"$`, w.outputListsTaskWithColumns)
	ctx.Step(`^the output does not list "([^"]*)"$`, w.outputDoesNotListTask)
	ctx.Step(`^the list output columns are exactly:$`, w.listOutputColumnsAreExactly)
	ctx.Step(`^the JSON output is an array of (\d+) tasks$`, w.jsonOutputIsArrayOfTasks)
	ctx.Step(`^the JSON output contains a task with identifier "([^"]*)"$`, w.jsonArrayContainsTaskID)
	ctx.Step(`^the JSON output does not contain a task with identifier "([^"]*)"$`, w.jsonArrayDoesNotContainTaskID)
	ctx.Step(`^the listed task identifiers appear in this order:$`, w.listedTaskIDsAppearInOrder)
	ctx.Step(`^the JSON output contains identifier "([^"]*)"$`, w.jsonStringField("id"))
}

func (w *world) followingTasksExist(table *godog.Table) error {
	allowedColumns := map[string]bool{
		"id": true, "status": true, "priority": true, "claimed by": true, "title": true, "tags": true,
	}
	for _, header := range table.Rows[0].Cells {
		if !allowedColumns[header.Value] {
			return fmt.Errorf("unsupported task fixture column %q", header.Value)
		}
	}

	for rowIdx, row := range table.Rows[1:] {
		values := map[string]string{}
		for i, cell := range row.Cells {
			values[table.Rows[0].Cells[i].Value] = strings.TrimSpace(cell.Value)
		}
		id := values["id"]
		if id == "" {
			return fmt.Errorf("task row %d is missing id", rowIdx+1)
		}
		status := values["status"]
		if status == "" {
			status = "open"
		}
		priority := values["priority"]
		if priority == "" {
			priority = "medium"
		}
		title := values["title"]
		if title == "" {
			title = id
		}

		fixture := &task.Task{
			ID:        id,
			Title:     title,
			Status:    status,
			Priority:  priority,
			CreatedAt: fixtureTime,
			UpdatedAt: fixtureTime,
			CreatedBy: "human",
			DependsOn: []string{},
			Tags:      []string{},
		}
		if actor := values["claimed by"]; actor != "" {
			claimedAt := fixtureTime
			expiresAt := fixtureTime.Add(time.Hour)
			fixture.Claim = task.Claim{
				Actor:       &actor,
				ClaimedAt:   &claimedAt,
				ExpiresAt:   &expiresAt,
				HeartbeatAt: &claimedAt,
			}
		}
		if tagsVal := values["tags"]; tagsVal != "" {
			for _, tag := range strings.Split(tagsVal, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					fixture.Tags = append(fixture.Tags, tag)
				}
			}
		}
		if err := writeFixtureTask(fixture); err != nil {
			return err
		}
	}
	return nil
}

func (w *world) outputListsTaskID(id string) error {
	if _, ok := lineContaining(w.stdout.String(), id); !ok {
		return fmt.Errorf("output does not list %q; got:\n%s", id, w.stdout.String())
	}
	return nil
}

func (w *world) outputListsTask(id, status, title string) error {
	line, ok := lineContaining(w.stdout.String(), id)
	if !ok {
		return fmt.Errorf("output does not list %q; got:\n%s", id, w.stdout.String())
	}
	if !strings.Contains(line, status) {
		return fmt.Errorf("line for %q does not contain status %q: %s", id, status, line)
	}
	if !strings.Contains(line, title) {
		return fmt.Errorf("line for %q does not contain title %q: %s", id, title, line)
	}
	return nil
}

func (w *world) outputListsTaskWithColumns(id, status, priority, claimedBy, title string) error {
	line, ok := lineContaining(w.stdout.String(), id)
	if !ok {
		return fmt.Errorf("output does not list %q; got:\n%s", id, w.stdout.String())
	}
	columns := splitListLine(line)
	expected := []string{id, status, priority, claimedBy, title}
	if len(columns) != len(expected) {
		return fmt.Errorf("line for %q has columns %v, expected %v; line: %s", id, columns, expected, line)
	}
	for i := range expected {
		if columns[i] != expected[i] {
			return fmt.Errorf("line for %q column %d = %q, expected %q; columns: %v", id, i+1, columns[i], expected[i], columns)
		}
	}
	return nil
}

func (w *world) outputDoesNotListTask(id string) error {
	if line, ok := lineContaining(w.stdout.String(), id); ok {
		return fmt.Errorf("output unexpectedly lists %q in line: %s\nfull output:\n%s", id, line, w.stdout.String())
	}
	return nil
}

func (w *world) listOutputColumnsAreExactly(table *godog.Table) error {
	lines := nonEmptyLines(w.stdout.String())
	if len(lines) == 0 {
		return fmt.Errorf("list output is empty")
	}
	actual := splitListLine(lines[0])
	var expected []string
	for _, row := range table.Rows[1:] {
		expected = append(expected, row.Cells[0].Value)
	}
	if strings.Join(actual, "|") != strings.Join(expected, "|") {
		return fmt.Errorf("list columns = %v, expected %v; output:\n%s", actual, expected, w.stdout.String())
	}
	return nil
}

func (w *world) jsonOutputIsArrayOfTasks(count int) error {
	tasks, err := w.jsonTaskArray()
	if err != nil {
		return err
	}
	if len(tasks) != count {
		return fmt.Errorf("JSON array has %d tasks, expected %d; got: %s", len(tasks), count, w.stdout.String())
	}
	return nil
}

func (w *world) jsonArrayContainsTaskID(id string) error {
	tasks, err := w.jsonTaskArray()
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if t.ID == id {
			return nil
		}
	}
	return fmt.Errorf("JSON array does not contain task %q; got: %s", id, w.stdout.String())
}

func (w *world) jsonArrayDoesNotContainTaskID(id string) error {
	tasks, err := w.jsonTaskArray()
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if t.ID == id {
			return fmt.Errorf("JSON array unexpectedly contains task %q; got: %s", id, w.stdout.String())
		}
	}
	return nil
}

func (w *world) listedTaskIDsAppearInOrder(table *godog.Table) error {
	var expected []string
	for _, row := range table.Rows[1:] {
		expected = append(expected, row.Cells[0].Value)
	}
	var actual []string
	for _, line := range nonEmptyLines(w.stdout.String()) {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "ID" {
			continue
		}
		actual = append(actual, fields[0])
	}
	if strings.Join(actual, ",") != strings.Join(expected, ",") {
		return fmt.Errorf("listed task ids = %v, expected %v; output:\n%s", actual, expected, w.stdout.String())
	}
	return nil
}

func (w *world) jsonTaskArray() ([]task.Task, error) {
	var tasks []task.Task
	if err := json.Unmarshal(w.stdout.Bytes(), &tasks); err != nil {
		return nil, fmt.Errorf("stdout is not a JSON task array (%v); got: %s", err, w.stdout.String())
	}
	return tasks, nil
}

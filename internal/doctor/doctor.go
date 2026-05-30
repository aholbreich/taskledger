// Package doctor scans a ledger for structural and data-integrity problems.
//
// It deliberately avoids store.List: that helper aborts on the first
// unparseable task file, but doctor's whole job is to find and report those
// files. So doctor reads the tasks directory entry by entry and tolerates
// per-file failures, turning each into a diagnostic instead of an error.
//
// Diagnose is read-only. Fix mutates the ledger and must be called while the
// caller holds the ledger lock.
package doctor

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aholbreich/tl/internal/events"
	"github.com/aholbreich/tl/internal/repo"
	"github.com/aholbreich/tl/internal/store"
	"github.com/aholbreich/tl/internal/task"
)

// Severities.
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
)

// Categories.
const (
	CategoryFrontmatter = "frontmatter"
	CategoryIdentity    = "identity"
	CategoryDependency  = "dependency"
	CategoryEvents      = "events"
	CategoryClaims      = "claims"
	CategoryTimestamps  = "timestamps"
	CategoryFilesystem  = "filesystem"
	CategoryBody        = "body"
	CategoryConfig      = "config"
	CategoryScale       = "scale"
	CategoryReferences  = "references"
)

// Scale thresholds: where filesystem scans and journal reads start to become
// noticeable on slower disks. Tune as the tool's performance evolves.
const (
	scaleTaskThreshold  = 100
	scaleEventThreshold = 1000
)

// fix kinds — drive the repair Fix performs and the verb it reports.
const (
	fixSelfDep   = "self-dependency"
	fixOrphanTmp = "orphan-tmp"
	fixOpenClaim = "open-claim"
	fixExpired   = "expired-claim"
	fixDeadRef   = "dead-reference"
)

// Diagnostic is one finding. The exported fields form the stable JSON shape;
// the unexported fields let Fix repair the issue and are never serialized.
type Diagnostic struct {
	Severity string `json:"severity"`
	Category string `json:"category"`
	TaskID   string `json:"task_id,omitempty"`
	Message  string `json:"message"`
	Fixable  bool   `json:"fixable"`

	fixKind   string // how Fix repairs this (empty when not fixable)
	fixTarget string // path or reference value the repair acts on
}

// Repair records a fix that was applied, with the verb to report it.
type Repair struct {
	Diagnostic Diagnostic
	Verb       string // fixed | removed | cleared | released
}

var (
	urlSchemeRE     = regexp.MustCompile(`^[a-z][a-z0-9+.-]*:`)
	conflictStartRE = regexp.MustCompile(`^<{7}`)
	conflictMidRE   = regexp.MustCompile(`^={7,}$`)
	conflictEndRE   = regexp.MustCompile(`^>{7}`)
	canonicalNoteRE = regexp.MustCompile(`^-\s+(\S+)\s+\[([^\]]*)\]\s+([A-Za-z_][A-Za-z0-9_-]*):\s*(.*)$`)
	notesHeadingRE  = regexp.MustCompile(`(?m)^## Notes\s*$`)
	validStatuses   = map[string]bool{"open": true, "in_progress": true, "blocked": true, "done": true, "cancelled": true, "pending_human": true}
	validPriorities = map[string]bool{"high": true, "medium": true, "low": true}
)

// Diagnose scans the ledger and returns every finding, sorted for stable
// output. It returns an error only when the ledger itself cannot be scanned
// (e.g. the tasks directory is unreadable) — individual bad files become
// diagnostics, not errors. The returned slice is never nil.
func Diagnose(ledger string) ([]Diagnostic, error) {
	now := time.Now().UTC()
	var diags []Diagnostic

	diags = append(diags, checkConfig(ledger)...)

	tasksDir := filepath.Join(ledger, repo.TasksDir)
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return nil, err
	}

	// Collect successfully-parsed tasks keyed by filename stem for the
	// cross-file checks (identity, dependencies, events).
	parsed := map[string]*task.Task{}
	idToFiles := map[string][]string{}
	taskFileCount := 0

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(name, ".md.tmp") {
			diags = append(diags, Diagnostic{
				Severity:  SeverityWarning,
				Category:  CategoryFilesystem,
				Message:   "orphaned temp file " + name + " (interrupted write?)",
				Fixable:   true,
				fixKind:   fixOrphanTmp,
				fixTarget: filepath.Join(tasksDir, name),
			})
			continue
		}
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		taskFileCount++
		stem := strings.TrimSuffix(name, ".md")

		data, err := os.ReadFile(filepath.Join(tasksDir, name))
		if err != nil {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Category: CategoryFilesystem,
				TaskID:   stem,
				Message:  "task file cannot be read: " + err.Error(),
			})
			continue
		}
		t, err := task.UnmarshalMarkdown(data)
		if err != nil {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Category: CategoryFrontmatter,
				TaskID:   stem,
				Message:  "invalid frontmatter: " + err.Error(),
			})
			continue
		}

		parsed[stem] = t
		if t.ID != "" {
			idToFiles[t.ID] = append(idToFiles[t.ID], stem)
		}

		diags = append(diags, checkFrontmatter(stem, t)...)
		diags = append(diags, checkClaims(stem, t, now)...)
		diags = append(diags, checkTimestamps(stem, t, now)...)
		diags = append(diags, checkBody(stem, t)...)
		diags = append(diags, checkReferences(ledger, stem, t)...)
	}

	diags = append(diags, checkIdentity(idToFiles)...)
	diags = append(diags, checkDependencies(parsed)...)

	eventCount, orphans := scanEvents(ledger, parsed)
	diags = append(diags, orphans...)

	diags = append(diags, checkScale(taskFileCount, eventCount)...)

	sortDiagnostics(diags)
	if diags == nil {
		diags = []Diagnostic{}
	}
	return diags, nil
}

// Fix scans the ledger, repairs every fixable diagnostic, and returns the
// repairs applied alongside the diagnostics it could not fix. The caller must
// hold the ledger lock.
func Fix(ledger string) (applied []Repair, unfixable []Diagnostic, err error) {
	diags, err := Diagnose(ledger)
	if err != nil {
		return nil, nil, err
	}
	for _, d := range diags {
		if !d.Fixable {
			unfixable = append(unfixable, d)
			continue
		}
		verb, ferr := applyRepair(ledger, d)
		if ferr != nil {
			// Surface a failed repair as a still-unfixable finding rather
			// than aborting the whole run.
			d.Message += " (repair failed: " + ferr.Error() + ")"
			d.Fixable = false
			unfixable = append(unfixable, d)
			continue
		}
		applied = append(applied, Repair{Diagnostic: d, Verb: verb})
	}
	return applied, unfixable, nil
}

func applyRepair(ledger string, d Diagnostic) (string, error) {
	switch d.fixKind {
	case fixOrphanTmp:
		return "removed", os.Remove(d.fixTarget)
	case fixSelfDep:
		t, err := store.Read(ledger, d.TaskID)
		if err != nil {
			return "", err
		}
		t.DependsOn = removeString(t.DependsOn, t.ID)
		t.UpdatedAt = time.Now().UTC().Truncate(time.Second)
		if err := store.Write(ledger, t); err != nil {
			return "", err
		}
		return "fixed", events.Append(ledger, events.Event{Event: "dependency_removed", TaskID: t.ID})
	case fixOpenClaim:
		t, err := store.Read(ledger, d.TaskID)
		if err != nil {
			return "", err
		}
		t.Claim = task.Claim{}
		t.UpdatedAt = time.Now().UTC().Truncate(time.Second)
		if err := store.Write(ledger, t); err != nil {
			return "", err
		}
		return "cleared", events.Append(ledger, events.Event{Event: "released", TaskID: t.ID, Actor: "tl-doctor"})
	case fixExpired:
		t, err := store.Read(ledger, d.TaskID)
		if err != nil {
			return "", err
		}
		t.Status = "open"
		t.Claim = task.Claim{}
		t.UpdatedAt = time.Now().UTC().Truncate(time.Second)
		if err := store.Write(ledger, t); err != nil {
			return "", err
		}
		return "released", events.Append(ledger, events.Event{Event: "released", TaskID: t.ID, Actor: "tl-doctor"})
	case fixDeadRef:
		t, err := store.Read(ledger, d.TaskID)
		if err != nil {
			return "", err
		}
		t.References = removeString(t.References, d.fixTarget)
		t.UpdatedAt = time.Now().UTC().Truncate(time.Second)
		if err := store.Write(ledger, t); err != nil {
			return "", err
		}
		return "fixed", events.Append(ledger, events.Event{Event: "reference_removed", TaskID: t.ID, Value: d.fixTarget})
	default:
		return "", os.ErrInvalid
	}
}

// --- per-task checks ------------------------------------------------------

func checkFrontmatter(id string, t *task.Task) []Diagnostic {
	var out []Diagnostic
	add := func(msg string) {
		out = append(out, Diagnostic{Severity: SeverityError, Category: CategoryFrontmatter, TaskID: id, Message: msg})
	}
	if strings.TrimSpace(t.Title) == "" {
		add("missing required field: title")
	}
	if !validStatuses[t.Status] {
		add("unknown status value: " + quote(t.Status))
	}
	if !validPriorities[t.Priority] {
		add("unknown priority value: " + quote(t.Priority))
	}
	if strings.TrimSpace(t.Type) == "" {
		add("missing required field: type")
	}
	return out
}

func checkClaims(id string, t *task.Task, now time.Time) []Diagnostic {
	switch {
	case t.Status == "open" && t.Claim.Actor != nil:
		return []Diagnostic{{
			Severity: SeverityWarning, Category: CategoryClaims, TaskID: id,
			Message: "open task still carries claim data",
			Fixable: true, fixKind: fixOpenClaim,
		}}
	case t.Claim.Actor != nil && t.Claim.ExpiresAt != nil && t.Claim.ExpiresAt.Before(now):
		return []Diagnostic{{
			Severity: SeverityWarning, Category: CategoryClaims, TaskID: id,
			Message: "claim lease expired and was never released",
			Fixable: true, fixKind: fixExpired,
		}}
	case t.Status == "in_progress" && t.Claim.Actor == nil:
		return []Diagnostic{{
			Severity: SeverityError, Category: CategoryClaims, TaskID: id,
			Message: "status is in_progress but the task has no claim",
		}}
	}
	return nil
}

func checkTimestamps(id string, t *task.Task, now time.Time) []Diagnostic {
	var out []Diagnostic
	warn := func(msg string) {
		out = append(out, Diagnostic{Severity: SeverityWarning, Category: CategoryTimestamps, TaskID: id, Message: msg})
	}
	if t.CreatedAt.After(now) {
		warn("created_at is in the future")
	}
	if t.CreatedAt.After(t.UpdatedAt) {
		warn("created_at is after updated_at")
	}
	if t.Claim.ExpiresAt != nil && t.Claim.ClaimedAt != nil && t.Claim.ExpiresAt.Before(*t.Claim.ClaimedAt) {
		warn("claim expiry is before the claim time")
	}
	return out
}

func checkBody(id string, t *task.Task) []Diagnostic {
	var out []Diagnostic
	for _, line := range strings.Split(t.Body, "\n") {
		if conflictStartRE.MatchString(line) || conflictMidRE.MatchString(line) || conflictEndRE.MatchString(line) {
			out = append(out, Diagnostic{
				Severity: SeverityError, Category: CategoryBody, TaskID: id,
				Message: "merge conflict marker in body",
			})
			break
		}
	}
	for _, line := range notesSectionLines(t.Body) {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			continue // only bullet lines are expected to be canonical notes
		}
		if !canonicalNoteRE.MatchString(trimmed) {
			out = append(out, Diagnostic{
				Severity: SeverityWarning, Category: CategoryBody, TaskID: id,
				Message: "note line does not match the canonical format",
			})
			break
		}
	}
	return out
}

func checkReferences(ledger, id string, t *task.Task) []Diagnostic {
	repoRoot := filepath.Dir(ledger)
	var out []Diagnostic
	for _, ref := range t.References {
		if urlSchemeRE.MatchString(ref) {
			continue // URL-shaped — skip, no network calls
		}
		if !strings.Contains(ref, "/") {
			continue // bare identifier / free text — not checkable
		}
		// Path-shaped: treat as repo-relative file path.
		if _, err := os.Stat(filepath.Join(repoRoot, ref)); err != nil {
			out = append(out, Diagnostic{
				Severity: SeverityWarning, Category: CategoryReferences, TaskID: id,
				Message: "referenced file does not exist: " + ref,
				Fixable: true, fixKind: fixDeadRef, fixTarget: ref,
			})
		}
	}
	return out
}

// --- cross-file checks ----------------------------------------------------

func checkIdentity(idToFiles map[string][]string) []Diagnostic {
	var out []Diagnostic
	for id, files := range idToFiles {
		if len(files) > 1 {
			sort.Strings(files)
			out = append(out, Diagnostic{
				Severity: SeverityError, Category: CategoryIdentity, TaskID: id,
				Message: "duplicate task id across files: " + strings.Join(files, ", "),
			})
		}
	}
	return out
}

func checkDependencies(parsed map[string]*task.Task) []Diagnostic {
	var out []Diagnostic
	graph := map[string][]string{}
	for stem, t := range parsed {
		id := t.ID
		if id == "" {
			id = stem
		}
		for _, dep := range t.DependsOn {
			if dep == id {
				out = append(out, Diagnostic{
					Severity: SeverityError, Category: CategoryDependency, TaskID: id,
					Message: "task depends on itself",
					Fixable: true, fixKind: fixSelfDep,
				})
				continue
			}
			if !taskExists(parsed, dep) {
				out = append(out, Diagnostic{
					Severity: SeverityError, Category: CategoryDependency, TaskID: id,
					Message: "depends on nonexistent task " + dep,
				})
				continue
			}
			graph[id] = append(graph[id], dep)
		}
	}
	for _, id := range cycleNodes(graph) {
		out = append(out, Diagnostic{
			Severity: SeverityError, Category: CategoryDependency, TaskID: id,
			Message: "task is part of a dependency cycle",
		})
	}
	return out
}

// taskExists reports whether dep matches a parsed task by its id field.
func taskExists(parsed map[string]*task.Task, dep string) bool {
	for _, t := range parsed {
		if t.ID == dep {
			return true
		}
	}
	return false
}

// cycleNodes returns every node that belongs to a dependency cycle, via
// Tarjan's strongly-connected-components algorithm. Any SCC larger than one
// node is a cycle; every member is reported.
func cycleNodes(graph map[string][]string) []string {
	index := 0
	indices := map[string]int{}
	lowlink := map[string]int{}
	onStack := map[string]bool{}
	var stack []string
	var result []string

	var strongConnect func(v string)
	strongConnect = func(v string) {
		indices[v] = index
		lowlink[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for _, w := range graph[v] {
			if _, seen := indices[w]; !seen {
				strongConnect(w)
				if lowlink[w] < lowlink[v] {
					lowlink[v] = lowlink[w]
				}
			} else if onStack[w] {
				if indices[w] < lowlink[v] {
					lowlink[v] = indices[w]
				}
			}
		}

		if lowlink[v] == indices[v] {
			var scc []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			if len(scc) > 1 {
				result = append(result, scc...)
			}
		}
	}

	// Deterministic traversal order.
	nodes := make([]string, 0, len(graph))
	for v := range graph {
		nodes = append(nodes, v)
	}
	sort.Strings(nodes)
	for _, v := range nodes {
		if _, seen := indices[v]; !seen {
			strongConnect(v)
		}
	}
	sort.Strings(result)
	return result
}

func scanEvents(ledger string, parsed map[string]*task.Task) (int, []Diagnostic) {
	p := filepath.Join(ledger, repo.EventsJournal)
	f, err := os.Open(p)
	if err != nil {
		return 0, nil
	}
	defer f.Close()

	count := 0
	seenOrphan := map[string]bool{}
	var out []Diagnostic
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		count++
		var e events.Event
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		if e.TaskID == "" || seenOrphan[e.TaskID] {
			continue
		}
		if _, ok := parsed[store.NormalizeID(e.TaskID)]; ok {
			continue
		}
		if taskExists(parsed, e.TaskID) {
			continue
		}
		seenOrphan[e.TaskID] = true
		out = append(out, Diagnostic{
			Severity: SeverityError, Category: CategoryEvents, TaskID: e.TaskID,
			Message: "event journal references task with no file",
		})
	}
	return count, out
}

func checkConfig(ledger string) []Diagnostic {
	if _, err := repo.LoadConfig(ledger); err != nil {
		if os.IsNotExist(err) {
			return []Diagnostic{{Severity: SeverityError, Category: CategoryConfig, Message: "config.yaml is missing"}}
		}
		return []Diagnostic{{Severity: SeverityError, Category: CategoryConfig, Message: "config.yaml is invalid: " + err.Error()}}
	}
	return nil
}

func checkScale(tasks, eventCount int) []Diagnostic {
	var out []Diagnostic
	if tasks > scaleTaskThreshold {
		out = append(out, Diagnostic{
			Severity: SeverityWarning, Category: CategoryScale,
			Message: "large ledger: many tasks (scans may slow down)",
		})
	}
	if eventCount > scaleEventThreshold {
		out = append(out, Diagnostic{
			Severity: SeverityWarning, Category: CategoryScale,
			Message: "large ledger: many events (journal reads may slow down)",
		})
	}
	return out
}

// --- helpers --------------------------------------------------------------

// notesSectionLines returns the lines of the body's ## Notes section, up to
// the next H2 heading (or end of body).
func notesSectionLines(body string) []string {
	loc := notesHeadingRE.FindStringIndex(body)
	if loc == nil {
		return nil
	}
	rest := body[loc[1]:]
	if next := regexp.MustCompile(`(?m)^## `).FindStringIndex(rest); next != nil {
		rest = rest[:next[0]]
	}
	return strings.Split(rest, "\n")
}

func removeString(list []string, v string) []string {
	out := make([]string, 0, len(list))
	for _, s := range list {
		if s == v {
			continue
		}
		out = append(out, s)
	}
	return out
}

func quote(s string) string { return "\"" + s + "\"" }

// sortDiagnostics orders findings by category, then severity (errors first),
// then task id, for stable human and JSON output.
func sortDiagnostics(d []Diagnostic) {
	sevRank := func(s string) int {
		if s == SeverityError {
			return 0
		}
		return 1
	}
	sort.SliceStable(d, func(i, j int) bool {
		if d[i].Category != d[j].Category {
			return d[i].Category < d[j].Category
		}
		if d[i].Severity != d[j].Severity {
			return sevRank(d[i].Severity) < sevRank(d[j].Severity)
		}
		if d[i].TaskID != d[j].TaskID {
			return d[i].TaskID < d[j].TaskID
		}
		return d[i].Message < d[j].Message
	})
}

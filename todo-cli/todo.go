package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
)

// Task is a single todo item.
type Task struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// TodoList is the set of tasks persisted to disk.
type TodoList struct {
	Tasks []Task `json:"tasks"`
}

// markResult reports the outcome of MarkDone.
type markResult int

const (
	markNotFound markResult = iota
	markAlreadyDone
	markChanged
)

// Add appends a new task with the next available ID and returns it.
func (t *TodoList) Add(text string) Task {
	id := 1
	for _, task := range t.Tasks {
		if task.ID >= id {
			id = task.ID + 1
		}
	}
	task := Task{ID: id, Text: text}
	t.Tasks = append(t.Tasks, task)
	return task
}

// MarkDone marks the task with the given ID as done, reporting whether the task
// was found and whether this call actually changed its state.
func (t *TodoList) MarkDone(id int) markResult {
	i := slices.IndexFunc(t.Tasks, func(task Task) bool { return task.ID == id })
	if i < 0 {
		return markNotFound
	}
	if t.Tasks[i].Done {
		return markAlreadyDone
	}
	t.Tasks[i].Done = true
	return markChanged
}

// Remove deletes the task with the given ID, reporting whether it existed.
func (t *TodoList) Remove(id int) bool {
	i := slices.IndexFunc(t.Tasks, func(task Task) bool { return task.ID == id })
	if i < 0 {
		return false
	}
	t.Tasks = slices.Delete(t.Tasks, i, i+1)
	return true
}

// Render writes the human-readable task list to w.
func (t *TodoList) Render(w io.Writer) {
	if len(t.Tasks) == 0 {
		fmt.Fprintln(w, "no tasks found")
		return
	}
	for _, task := range t.Tasks {
		status := "[ ]"
		if task.Done {
			status = "[x]"
		}
		fmt.Fprintf(w, "%d %s %s\n", task.ID, status, task.Text)
	}
}

// validate enforces the invariants the rest of the code relies on: positive,
// unique task IDs. A corrupt or hand-edited file that breaks these is rejected
// loudly rather than silently misbehaving (e.g. done/remove acting on only the
// first of two tasks sharing an ID).
func (t *TodoList) validate() error {
	seen := make(map[int]struct{}, len(t.Tasks))
	for _, task := range t.Tasks {
		if task.ID <= 0 {
			return fmt.Errorf("task id must be positive, got %d", task.ID)
		}
		if _, dup := seen[task.ID]; dup {
			return fmt.Errorf("duplicate task id %d", task.ID)
		}
		seen[task.ID] = struct{}{}
	}
	return nil
}

// Load reads the todo list from path. A missing or empty file yields an empty
// list; a present-but-corrupt file returns an error naming the path and how to
// recover, so the user is never left stuck.
func Load(path string) (*TodoList, error) {
	list := &TodoList{}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return list, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	if len(data) == 0 {
		return list, nil
	}
	if err := json.Unmarshal(data, list); err != nil {
		return nil, fmt.Errorf("parsing %s: %w (back up and delete this file to recover)", path, err)
	}
	if err := list.validate(); err != nil {
		return nil, fmt.Errorf("%s is corrupt: %w", path, err)
	}
	return list, nil
}

// Save writes the list to path atomically and durably: it writes to a temp file
// in the same directory, fsyncs it, then renames over the target. A crash can
// therefore never leave a truncated or partial todo file.
func (t *TodoList) Save(path string) (err error) {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding todo list: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".todo-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	// On any failure below, remove the temp file. After a successful rename
	// tmpName no longer exists, so this Remove is a harmless no-op.
	defer func() {
		if err != nil {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing %s: %w", tmpName, err)
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing %s: %w", tmpName, err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("closing %s: %w", tmpName, err)
	}
	if err = os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}

// storagePath resolves the data-file location. TODO_FILE overrides everything;
// otherwise the file lives in a stable per-user config directory so the list is
// the same no matter which directory the command is run from.
func storagePath() (string, error) {
	if p := os.Getenv("TODO_FILE"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving user config dir: %w", err)
	}
	return filepath.Join(dir, "todo", "todo.json"), nil
}

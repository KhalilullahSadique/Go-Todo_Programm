package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddAssignsSequentialIDs(t *testing.T) {
	list := &TodoList{}
	a := list.Add("first")
	b := list.Add("second")
	if a.ID != 1 || b.ID != 2 {
		t.Fatalf("got IDs %d,%d, want 1,2", a.ID, b.ID)
	}
	if len(list.Tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(list.Tasks))
	}
}

func TestAddIDIsMaxPlusOne(t *testing.T) {
	list := &TodoList{Tasks: []Task{{ID: 5, Text: "x"}}}
	if got := list.Add("y"); got.ID != 6 {
		t.Fatalf("got ID %d, want 6", got.ID)
	}
}

func TestMarkDoneStates(t *testing.T) {
	list := &TodoList{Tasks: []Task{{ID: 1, Text: "x"}}}
	if got := list.MarkDone(1); got != markChanged {
		t.Fatalf("first MarkDone = %v, want markChanged", got)
	}
	if !list.Tasks[0].Done {
		t.Fatal("task was not marked done")
	}
	if got := list.MarkDone(1); got != markAlreadyDone {
		t.Fatalf("second MarkDone = %v, want markAlreadyDone", got)
	}
	if got := list.MarkDone(99); got != markNotFound {
		t.Fatalf("missing MarkDone = %v, want markNotFound", got)
	}
}

func TestRemove(t *testing.T) {
	list := &TodoList{Tasks: []Task{{ID: 1}, {ID: 2}, {ID: 3}}}
	if !list.Remove(2) {
		t.Fatal("Remove(2) = false, want true")
	}
	if len(list.Tasks) != 2 || list.Tasks[0].ID != 1 || list.Tasks[1].ID != 3 {
		t.Fatalf("unexpected tasks after remove: %+v", list.Tasks)
	}
	if list.Remove(2) {
		t.Fatal("Remove(2) a second time = true, want false")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todo.json")
	orig := &TodoList{Tasks: []Task{{ID: 1, Text: "a"}, {ID: 2, Text: "b", Done: true}}}
	if err := orig.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Tasks) != 2 || got.Tasks[1].Text != "b" || !got.Tasks[1].Done {
		t.Fatalf("round-trip mismatch: %+v", got.Tasks)
	}
}

func TestSaveIsAtomicOverExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todo.json")
	first := &TodoList{Tasks: []Task{{ID: 1, Text: "a"}}}
	if err := first.Save(path); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	second := &TodoList{Tasks: []Task{{ID: 1, Text: "a"}, {ID: 2, Text: "b"}}}
	if err := second.Save(path); err != nil {
		t.Fatalf("second Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Tasks) != 2 {
		t.Fatalf("got %d tasks after overwrite, want 2", len(got.Tasks))
	}
	// No leftover temp files should remain in the directory.
	entries, _ := os.ReadDir(filepath.Dir(path))
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("leftover temp file: %s", e.Name())
		}
	}
}

func TestLoadMissingFile(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatalf("Load missing file: %v", err)
	}
	if len(got.Tasks) != 0 {
		t.Fatalf("want empty list, got %+v", got.Tasks)
	}
}

func TestLoadEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todo.json")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load empty file: %v", err)
	}
	if len(got.Tasks) != 0 {
		t.Fatalf("want empty list, got %+v", got.Tasks)
	}
}

func TestLoadCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todo.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load corrupt file = nil error, want error")
	}
}

func TestLoadRejectsDuplicateIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todo.json")
	if err := os.WriteFile(path, []byte(`{"tasks":[{"id":1},{"id":1}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load with duplicate IDs = nil error, want error")
	}
}

func TestLoadRejectsNonPositiveIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todo.json")
	if err := os.WriteFile(path, []byte(`{"tasks":[{"id":0}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load with id 0 = nil error, want error")
	}
}

func TestStoragePathEnvOverride(t *testing.T) {
	t.Setenv("TODO_FILE", "/custom/todo.json")
	got, err := storagePath()
	if err != nil {
		t.Fatal(err)
	}
	if got != "/custom/todo.json" {
		t.Fatalf("got %q, want /custom/todo.json", got)
	}
}

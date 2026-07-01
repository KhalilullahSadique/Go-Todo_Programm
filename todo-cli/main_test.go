package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tempStore points TODO_FILE at a throwaway file for the duration of a test.
func tempStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "todo.json")
	t.Setenv("TODO_FILE", path)
	return path
}

func TestRunAddListDone(t *testing.T) {
	tempStore(t)
	var out, errBuf bytes.Buffer

	if code := run([]string{"add", "buy", "milk"}, &out, &errBuf); code != 0 {
		t.Fatalf("add exit %d, stderr=%q", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "added task 1") {
		t.Fatalf("add output = %q", out.String())
	}

	out.Reset()
	run([]string{"list"}, &out, &errBuf)
	if !strings.Contains(out.String(), "1 [ ] buy milk") {
		t.Fatalf("list output = %q", out.String())
	}

	out.Reset()
	if code := run([]string{"done", "1"}, &out, &errBuf); code != 0 {
		t.Fatalf("done exit %d", code)
	}
	out.Reset()
	run([]string{"list"}, &out, &errBuf)
	if !strings.Contains(out.String(), "1 [x] buy milk") {
		t.Fatalf("list after done = %q", out.String())
	}
}

func TestRunAddEmptyIsRejected(t *testing.T) {
	tempStore(t)
	var out, errBuf bytes.Buffer
	if code := run([]string{"add", "   "}, &out, &errBuf); code == 0 {
		t.Fatal("add of whitespace-only text exited 0, want non-zero")
	}
	if !strings.Contains(errBuf.String(), "cannot be empty") {
		t.Fatalf("stderr = %q", errBuf.String())
	}
}

func TestRunDoneAlreadyDoneSkipsError(t *testing.T) {
	tempStore(t)
	var out, errBuf bytes.Buffer
	run([]string{"add", "task"}, &out, &errBuf)
	run([]string{"done", "1"}, &out, &errBuf)

	out.Reset()
	errBuf.Reset()
	if code := run([]string{"done", "1"}, &out, &errBuf); code != 0 {
		t.Fatalf("second done exit %d, want 0", code)
	}
	if !strings.Contains(out.String(), "already done") {
		t.Fatalf("expected 'already done' feedback, got %q", out.String())
	}
}

func TestRunDoneNotFound(t *testing.T) {
	tempStore(t)
	var out, errBuf bytes.Buffer
	if code := run([]string{"done", "5"}, &out, &errBuf); code != 1 {
		t.Fatalf("done on missing id exit %d, want 1", code)
	}
}

func TestRunDoneNonNumeric(t *testing.T) {
	tempStore(t)
	var out, errBuf bytes.Buffer
	if code := run([]string{"done", "abc"}, &out, &errBuf); code != 1 {
		t.Fatalf("done abc exit %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "must be a number") {
		t.Fatalf("stderr = %q", errBuf.String())
	}
}

func TestRunRemove(t *testing.T) {
	tempStore(t)
	var out, errBuf bytes.Buffer
	run([]string{"add", "one"}, &out, &errBuf)
	out.Reset()
	if code := run([]string{"remove", "1"}, &out, &errBuf); code != 0 {
		t.Fatalf("remove exit %d", code)
	}
	out.Reset()
	run([]string{"list"}, &out, &errBuf)
	if !strings.Contains(out.String(), "no tasks found") {
		t.Fatalf("list after remove = %q", out.String())
	}
}

func TestRunHelpIgnoresCorruptFile(t *testing.T) {
	path := tempStore(t)
	if err := os.WriteFile(path, []byte("{garbage"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	if code := run([]string{"help"}, &out, &errBuf); code != 0 {
		t.Fatalf("help with corrupt file exit %d, want 0", code)
	}
	if !strings.Contains(out.String(), "todo - simple task manager") {
		t.Fatalf("help output = %q", out.String())
	}
}

func TestRunDataCommandReportsCorruptFile(t *testing.T) {
	path := tempStore(t)
	if err := os.WriteFile(path, []byte("{garbage"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	if code := run([]string{"list"}, &out, &errBuf); code != 1 {
		t.Fatalf("list with corrupt file exit %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), path) {
		t.Fatalf("error should name the file path, got %q", errBuf.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	tempStore(t)
	var out, errBuf bytes.Buffer
	if code := run([]string{"frobnicate"}, &out, &errBuf); code != 1 {
		t.Fatalf("unknown command exit %d, want 1", code)
	}
}

func TestRunNoArgs(t *testing.T) {
	tempStore(t)
	var out, errBuf bytes.Buffer
	if code := run(nil, &out, &errBuf); code != 1 {
		t.Fatalf("no args exit %d, want 1", code)
	}
}

package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes one CLI invocation and returns the process exit code. Keeping the
// logic here (rather than in main with scattered os.Exit calls) makes every
// command path unit-testable with in-memory writers.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 1
	}

	cmd := args[0]

	// Usage/help must never depend on the data file, so a corrupt or unreadable
	// todo file can never stop the user from getting help.
	switch cmd {
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	}

	path, err := storagePath()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	list, err := Load(path)
	if err != nil {
		fmt.Fprintf(stderr, "error loading todo list: %v\n", err)
		return 1
	}

	switch cmd {
	case "list":
		list.Render(stdout)
		return 0

	case "add":
		if len(args) < 2 {
			fmt.Fprintln(stderr, "usage: todo add <task description>")
			return 1
		}
		text := strings.TrimSpace(strings.Join(args[1:], " "))
		if text == "" {
			fmt.Fprintln(stderr, "task description cannot be empty")
			return 1
		}
		task := list.Add(text)
		if err := list.Save(path); err != nil {
			fmt.Fprintf(stderr, "error saving todo list: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "added task %d\n", task.ID)
		return 0

	case "done":
		id, ok := parseID(cmd, args, stderr)
		if !ok {
			return 1
		}
		switch list.MarkDone(id) {
		case markNotFound:
			fmt.Fprintf(stderr, "task %d not found\n", id)
			return 1
		case markAlreadyDone:
			// Nothing changed: report it and skip the needless file rewrite.
			fmt.Fprintf(stdout, "task %d already done\n", id)
			return 0
		default: // markChanged
			if err := list.Save(path); err != nil {
				fmt.Fprintf(stderr, "error saving todo list: %v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "marked task %d done\n", id)
			return 0
		}

	case "remove":
		id, ok := parseID(cmd, args, stderr)
		if !ok {
			return 1
		}
		if !list.Remove(id) {
			fmt.Fprintf(stderr, "task %d not found\n", id)
			return 1
		}
		if err := list.Save(path); err != nil {
			fmt.Fprintf(stderr, "error saving todo list: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "removed task %d\n", id)
		return 0

	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", cmd)
		printUsage(stderr)
		return 1
	}
}

// parseID validates that a "<cmd> <id>" invocation received exactly one numeric
// argument, writing a usage/error message to stderr otherwise.
func parseID(cmd string, args []string, stderr io.Writer) (int, bool) {
	if len(args) != 2 {
		fmt.Fprintf(stderr, "usage: todo %s <task id>\n", cmd)
		return 0, false
	}
	id, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintln(stderr, "task id must be a number")
		return 0, false
	}
	return id, true
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "todo - simple task manager")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  todo list                    list all tasks")
	fmt.Fprintln(w, "  todo add <task description>  add a task")
	fmt.Fprintln(w, "  todo done <task id>          mark a task done")
	fmt.Fprintln(w, "  todo remove <task id>        delete a task")
	fmt.Fprintln(w, "  todo help                    show this help")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Storage: $TODO_FILE if set, otherwise <user-config-dir>/todo/todo.json")
}

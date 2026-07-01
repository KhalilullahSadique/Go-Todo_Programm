package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type Task struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

type TodoList struct {
	Tasks []Task `json:"tasks"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	path, err := storagePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	list, err := loadTodoList(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading todo list: %v\n", err)
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "list":
		list.print()
	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: todo add <task description>")
			os.Exit(1)
		}
		taskText := joinArgs(os.Args[2:])
		list.add(taskText)
		if err := list.save(path); err != nil {
			fmt.Fprintf(os.Stderr, "error saving todo list: %v\n", err)
			os.Exit(1)
		}
	case "done":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "usage: todo done <task id>")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "task id must be a number")
			os.Exit(1)
		}
		if !list.markDone(id) {
			fmt.Fprintf(os.Stderr, "task %d not found\n", id)
			os.Exit(1)
		}
		if err := list.save(path); err != nil {
			fmt.Fprintf(os.Stderr, "error saving todo list: %v\n", err)
			os.Exit(1)
		}
	case "remove":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "usage: todo remove <task id>")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "task id must be a number")
			os.Exit(1)
		}
		if !list.remove(id) {
			fmt.Fprintf(os.Stderr, "task %d not found\n", id)
			os.Exit(1)
		}
		if err := list.save(path); err != nil {
			fmt.Fprintf(os.Stderr, "error saving todo list: %v\n", err)
			os.Exit(1)
		}
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("todo - simple task manager")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  todo list")
	fmt.Println("  todo add <task description>")
	fmt.Println("  todo done <task id>")
	fmt.Println("  todo remove <task id>")
	fmt.Println("  todo help")
}

func storagePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, ".todo.json"), nil
}

func loadTodoList(path string) (*TodoList, error) {
	list := &TodoList{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return list, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return list, nil
	}
	if err := json.Unmarshal(data, list); err != nil {
		return nil, err
	}
	return list, nil
}

func (t *TodoList) save(path string) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0o600)
}

func (t *TodoList) add(text string) {
	id := 1
	for _, task := range t.Tasks {
		if task.ID >= id {
			id = task.ID + 1
		}
	}
	t.Tasks = append(t.Tasks, Task{ID: id, Text: text})
	fmt.Printf("added task %d\n", id)
}

func (t *TodoList) markDone(id int) bool {
	for i, task := range t.Tasks {

		if task.ID == id {

			if task.Done {
				return true
			}
			t.Tasks[i].Done = true
			// fmt.Printf("marked task %d done\n", id)
			fmt.Println("marked task done")
			return true
		}
	}
	return false
}
func (t *TodoList) remove(id int) bool {
	for i, task := range t.Tasks {
		if task.ID == id {
			t.Tasks = append(t.Tasks[:i], t.Tasks[i+1:]...)
			fmt.Printf("removed task %d\n", id)
			return true
		}
	}
	return false
}

func (t *TodoList) print() {
	if len(t.Tasks) == 0 {
		fmt.Println("no tasks found")
		return
	}
	for _, task := range t.Tasks {
		status := "[ ]"
		if task.Done {
			status = "[x]"
		}
		fmt.Printf("%d %s %s\n", task.ID, status, task.Text)
	}
}

func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}

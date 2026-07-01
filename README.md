# todo

A small, dependency-free command-line task manager written in Go. Tasks are
stored as JSON and every write is atomic, so an interrupted save can never
corrupt or lose your list.

## Install

```sh
go install github.com/KhalilullahSadique/Go-Todo_Programm/todo-cli@latest
```

Or build from a checkout:

```sh
cd todo-cli
go build -o todo .
```

## Usage

```
todo list                    list all tasks
todo add <task description>   add a task
todo done <task id>           mark a task done
todo remove <task id>         delete a task
todo help                     show help
```

Example:

```sh
$ todo add buy milk
added task 1
$ todo add call the plumber
added task 2
$ todo done 1
marked task 1 done
$ todo list
1 [x] buy milk
2 [ ] call the plumber
```

## Storage

By default the list is stored at `<user-config-dir>/todo/todo.json`
(`~/.config/todo/todo.json` on Linux), so it is the same no matter which
directory you run `todo` from. Override the location with the `TODO_FILE`
environment variable:

```sh
TODO_FILE=./project-tasks.json todo add ship the release
```

## Development

The module lives in [`todo-cli/`](todo-cli/):

```sh
cd todo-cli
go test ./...      # run the test suite
go vet ./...       # static checks
gofmt -l .         # should print nothing
```

## License

Released under the [MIT License](LICENSE).

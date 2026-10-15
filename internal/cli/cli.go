package cli

import (
	"bufio"
	"cli-task-traker/internal/domain"
	"cli-task-traker/internal/service"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"uuid"
)

type CLI struct {
	service *service.Service
	input   io.Reader
	output  io.Writer
	scanner *bufio.Scanner
}

func NewCLI(service *service.Service, input io.Reader, output io.Writer) *CLI {
	return &CLI{
		service: service,
		input:   input,
		output:  output,
		scanner: bufio.NewScanner(input),
	}
}

func (cli *CLI) Run() error {
	fmt.Printf("TaskTracker. Type 'help' for commands\n")
	for {
		fmt.Printf("> ")

		if !cli.scanner.Scan() {
			if err := cli.scanner.Err(); err != nil {
				return fmt.Errorf("error reading input %w", err)
			}
			return nil
		}

		line := strings.TrimSpace(cli.scanner.Text())

		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		command := strings.ToLower(fields[0])
		args := fields[1:]

		stop, err := cli.dispatch(command, args)

		if err != nil {
			fmt.Printf("error: %v\n", err)
		}

		if stop {
			return nil
		}
	}
}

func (cli *CLI) dispatch(command string, args []string) (bool, error) {
	switch command {
	case "help":
		cli.help()
		return false, nil
	case "exit", "quit":
		fmt.Printf("bye\n")
		return true, nil
	case "add":
		return false, cli.add(args)
	case "list":
		return false, cli.list(args)
	case "get":
		return false, cli.get(args)
	case "status":
		return false, cli.status(args)
	case "edit":
		return false, cli.edit(args)
	case "delete", "rm":
		return false, cli.rm(args)

	default:
		return false, fmt.Errorf("unknown command: %q (try 'help')", command)
	}
}

func (cli *CLI) help() {
	fmt.Printf("Available commands:\n")
	fmt.Printf("  add <title> <deadline YYYY-MM-DD>          - create task\n")
	fmt.Printf("  list [status=<status>] [overdue] [search=<text>] [page=<n>] [limit=<n>]\n")
	fmt.Printf("                                             - list tasks (sorted by deadline, then id)\n")
	fmt.Printf("  get <id>                                   - show task by id\n")
	fmt.Printf("  status <id> <status>                       - change status\n")
	fmt.Printf("  edit <id> [title=<text>] [deadline=DATE]   - edit title/deadline\n")
	fmt.Printf("  rm <id>                                	- delete task\n")
	fmt.Printf("  help                                       - this help\n")
	fmt.Printf("  exit                                       - quit\n")
	fmt.Printf("Statuses: %s\n", domain.AllStatuses)
}

func (cli *CLI) list(args []string) error {
	filter := service.ListFilter{}
	page, limit := 1, 20

	for _, arg := range args {
		key, val, hasVal := strings.Cut(arg, "=")
		switch strings.ToLower(key) {
		case "status":
			if !hasVal {
				return fmt.Errorf("list: invalid argument")
			}
			status := domain.Status(val)
			if err := status.IsValid(); err != nil {
				return err
			}
			filter.Status = &status
		case "overdue":
			filter.OnlyPastDeadLine = true
		case "search":
			if !hasVal {
				return errors.New("list: search= requires value")
			}
			filter.Search = val
		case "page":
			n, err := strconv.Atoi(val)
			if err != nil || n < 1 {
				return fmt.Errorf("list: invalid page %q", val)
			}
			page = n
		case "limit":
			n, err := strconv.Atoi(val)
			if err != nil || n < 1 {
				return fmt.Errorf("list: invalid limit %q", val)
			}
			limit = n
		default:
			return fmt.Errorf("list: unknown argument: %q", arg)
		}
	}

	tasks, err := cli.service.List(filter)
	if err != nil {
		return err
	}

	pageTasks, total, totalPages := service.Paginate(tasks, page, limit)

	if total == 0 {
		fmt.Printf("no tasks match the filter\n")
		return nil
	}
	if len(pageTasks) == 0 {
		fmt.Printf("page %d is empty (total %d, total pages %d)\n", page, total, totalPages)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tTITLE\tDEADLINE")

	for _, t := range pageTasks {
		fmt.Fprintln(w, FormatTask(t))
	}

	w.Flush()
	fmt.Printf("showing page %d/%d, total %d task(s)\n", page, totalPages, total)
	return nil
}

func (cli *CLI) edit(args []string) error {
	if len(args) < 2 {
		return errors.New("usage: edit <id> [title=...] [deadline=YYYY-MM-DD]")
	}

	id, err := parseID(args[:1])
	if err != nil {
		return err
	}

	patch := service.EditPatch{}
	changed := false
	for _, a := range args[1:] {
		key, val, ok := strings.Cut(a, "=")
		if !ok {
			return fmt.Errorf("edit: expected key=value, got %q", a)
		}

		switch strings.ToLower(key) {
		case "title":
			patch.Title = &val
			changed = true
		case "deadline":
			deadline, err := domain.ParseDeadline(val)
			if err != nil {
				return fmt.Errorf("edit: deadline exception %w", err)
			}
			patch.Deadline = &deadline
			changed = true
		default:
			return fmt.Errorf("edit: unrecognized key %q (try 'help')", key)
		}
	}

	if !changed {
		return errors.New("edit: nothing to change")
	}

	task, err := cli.service.Edit(id, patch)
	if err != nil {
		return err
	}

	fmt.Printf("task updated: %v \n", FormatTask(task))
	return nil
}

func (cli *CLI) status(args []string) error {
	if len(args) != 2 {
		return errors.New("expected only 2 arguments")
	}

	id, err := parseID(args[:1])
	if err != nil {
		return err
	}

	newStatus := domain.Status(strings.ToLower(args[1]))

	task, err := cli.service.EditStatus(id, newStatus)
	if err != nil {
		return err
	}

	fmt.Printf("task`s status updated: %v \n", FormatTask(task))
	return nil

}

func parseID(args []string) (uuid.UUID, error) {
	if len(args) != 1 {
		return uuid.UUID{}, errors.New("expected only one id argument")
	}

	id, err := uuid.Parse(args[0])
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid id %q: %w", args[0], err)
	}

	return id, nil
}

func (cli *CLI) get(args []string) error {
	id, err := parseID(args)
	if err != nil {
		return err
	}

	task, err := cli.service.Get(id)
	if err != nil {
		return err
	}

	fmt.Println(FormatTask(task))
	return nil
}

func (cli *CLI) add(args []string) error {
	if len(args) < 2 {
		return errors.New("usage: add <title> <YYYY-MM-DD>\n}")
	}

	title := strings.TrimSpace(strings.Join(args[:len(args)-1], " "))
	deadline, err := domain.ParseDeadline(args[len(args)-1])
	if err != nil {
		return fmt.Errorf("invalid deadline %q: %w", args[len(args)-1], err)
	}

	task, err := cli.service.Add(title, deadline)
	if err != nil {
		return err
	}
	fmt.Printf("task created: %v \n", FormatTask(task))
	return nil
}

func (cli *CLI) rm(args []string) error {
	id, err := parseID(args)
	if err != nil {
		return err
	}

	err = cli.service.Delete(id)
	if err != nil {
		return err
	}

	fmt.Printf("task deleted: %v \n", id)
	return nil
}

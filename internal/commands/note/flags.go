package notecmd

import (
	"strings"
)

// SubAction defines the action to perform with notes.
type SubAction string

const (
	ActionList   SubAction = "list"
	ActionAdd    SubAction = "add"
	ActionShow   SubAction = "show"
	ActionTodo   SubAction = "todo"
	ActionToggle SubAction = "toggle"
	ActionDelete SubAction = "delete"
	ActionSearch SubAction = "search"
)

// Options holds configuration for note command.
type Options struct {
	Action   SubAction `json:"action"`
	ID       string    `json:"id,omitempty"`
	Title    string    `json:"title,omitempty"`
	Content  string    `json:"content,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
	Query    string    `json:"query,omitempty"`
	ItemIdx  int       `json:"item_idx,omitempty"`
	Dir      string    `json:"dir,omitempty"`
	Plain    bool      `json:"plain"`
	JSON     bool      `json:"json"`
}

// ParseFlags parses command line arguments for note command.
func ParseFlags(args []string) Options {
	opts := Options{
		Action: ActionList,
	}

	var positional []string
	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--plain" {
			opts.Plain = true
			i++
		} else if arg == "--json" {
			opts.JSON = true
			i++
		} else if strings.HasPrefix(arg, "--dir=") {
			opts.Dir = strings.TrimPrefix(arg, "--dir=")
			i++
		} else if (arg == "--dir" || arg == "-d") && i+1 < len(args) {
			opts.Dir = args[i+1]
			i += 2
		} else if strings.HasPrefix(arg, "--tags=") {
			opts.Tags = strings.Split(strings.TrimPrefix(arg, "--tags="), ",")
			i++
		} else if (arg == "--tags" || arg == "-t") && i+1 < len(args) {
			opts.Tags = strings.Split(args[i+1], ",")
			i += 2
		} else if strings.HasPrefix(arg, "--content=") {
			opts.Content = strings.TrimPrefix(arg, "--content=")
			i++
		} else if (arg == "--content" || arg == "-c") && i+1 < len(args) {
			opts.Content = args[i+1]
			i += 2
		} else if strings.HasPrefix(arg, "--query=") {
			opts.Query = strings.TrimPrefix(arg, "--query=")
			i++
		} else if (arg == "--query" || arg == "-q") && i+1 < len(args) {
			opts.Query = args[i+1]
			i += 2
		} else if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			i++
		} else {
			i++
		}
	}

	if len(positional) > 0 {
		cmd := strings.ToLower(positional[0])
		switch cmd {
		case "add", "new", "create":
			opts.Action = ActionAdd
			if len(positional) > 1 {
				opts.Title = strings.Join(positional[1:], " ")
			}
		case "show", "view", "get":
			opts.Action = ActionShow
			if len(positional) > 1 {
				opts.ID = positional[1]
			}
		case "todo", "todos", "tasks":
			opts.Action = ActionTodo
		case "toggle", "check", "done":
			opts.Action = ActionToggle
			if len(positional) > 1 {
				opts.ID = positional[1]
			}
			if len(positional) > 2 {
				// Parse task index 1-based
				var idx int
				for _, c := range positional[2] {
					if c >= '0' && c <= '9' {
						idx = idx*10 + int(c-'0')
					}
				}
				opts.ItemIdx = idx
			}
		case "rm", "delete", "remove":
			opts.Action = ActionDelete
			if len(positional) > 1 {
				opts.ID = positional[1]
			}
		case "search", "find":
			opts.Action = ActionSearch
			if len(positional) > 1 {
				opts.Query = strings.Join(positional[1:], " ")
			}
		case "list", "ls":
			opts.Action = ActionList
		default:
			// If not a subcommand name, treat as "show <id>" if 1 positional, or search query
			if len(positional) == 1 {
				opts.Action = ActionShow
				opts.ID = positional[0]
			} else {
				opts.Action = ActionSearch
				opts.Query = strings.Join(positional, " ")
			}
		}
	}

	// Clean up tags
	for j := range opts.Tags {
		opts.Tags[j] = strings.TrimSpace(opts.Tags[j])
	}

	return opts
}

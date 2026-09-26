package cli

// Command represents a runnable CLI subcommand.
type Command struct {
	Name        string
	Aliases     []string
	Summary     string
	Usage       string
	Description string
	Phase       int // Roadmap phase where this command is implemented
	Run         func(ctx *Context, args []string) error
}

// Matches returns true if the given name matches the command name or any alias.
func (c *Command) Matches(name string) bool {
	if c.Name == name {
		return true
	}
	for _, alias := range c.Aliases {
		if alias == name {
			return true
		}
	}
	return false
}

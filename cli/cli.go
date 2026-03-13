package cli

import (
	"errors"

	"github.com/Acketuk/Gator/internal/config"
)

type Command struct {
	Name string
	Args []string
}
type Commands struct {
	Handlers map[string]func(*config.State, Command) error
}

func (c *Commands) Register(Name string, f func(*config.State, Command) error) {
	c.Handlers[Name] = f
}

func (c *Commands) Run(state *config.State, cmd Command) error {

	handler, ok := c.Handlers[cmd.Name]
	if !ok {
		return errors.New("Unknown command\n")
	}
	return handler(state, cmd)
}

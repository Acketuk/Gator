package middleware

import (
	"context"

	"github.com/Acketuk/Gator/cli"
	"github.com/Acketuk/Gator/internal/config"
	"github.com/Acketuk/Gator/internal/database"
)

func LoggedIn(handler func(state *config.State, cmd cli.Command, user database.User) error) func(*config.State, cli.Command) error {
	return func(s *config.State, cmd cli.Command) error {

		ActiveUserName := s.Config.CurrentUserName
		user, err := s.Db.GetUser(context.Background(), ActiveUserName)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
}

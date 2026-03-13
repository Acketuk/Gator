package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Acketuk/Gator/cli"
	"github.com/Acketuk/Gator/internal/config"
	"github.com/Acketuk/Gator/internal/database"
	"github.com/Acketuk/Gator/middleware"

	_ "github.com/lib/pq"
)

func main() {

	conf, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	state := config.State{
		Config: conf,
	}

	db, err := sql.Open("postgres", state.Config.UrlDB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dbQueries := database.New(db)
	state.Db = dbQueries

	if len(os.Args) < 2 {
		fmt.Println("not enough arguments were provided")
		os.Exit(1)
	}

	commands := cli.Commands{
		Handlers: make(map[string]func(*config.State, cli.Command) error),
	}
	commands.Register("login", cli.HandlerLogin)
	commands.Register("register", cli.HandlerRegister)
	commands.Register("reset", cli.HandlerReset)
	commands.Register("users", cli.HandlerListUsers)
	commands.Register("agg", cli.HandlerAgg)
	commands.Register("addfeed", middleware.LoggedIn(cli.HandlerAddFeed))
	commands.Register("feeds", cli.HandlerGetFeeds)
	commands.Register("follow", middleware.LoggedIn(cli.HandlerFollow))
	commands.Register("following", middleware.LoggedIn(cli.HandlerFollowing))
	commands.Register("unfollow", middleware.LoggedIn(cli.HandlerUnfollow))
	commands.Register("browse", middleware.LoggedIn(cli.HandlerBrowse))

	cmd := cli.Command{Name: os.Args[1], Args: os.Args[2:]}
	if err := commands.Run(&state, cmd); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//fmt.Printf("command: %s\nparams: %#v\n\n", cmd.Name, cmd.Args)
	//fmt.Printf("current user name: %s\n", conf.CurrentUserName)

}

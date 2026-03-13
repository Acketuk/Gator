package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Acketuk/Gator/internal/config"
	"github.com/Acketuk/Gator/internal/database"
	"github.com/Acketuk/Gator/rss"
	"github.com/google/uuid"
)

func HandlerLogin(state *config.State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return errors.New("username is required\n")
	}

	name := cmd.Args[0]

	_, err := state.Db.GetUser(context.Background(), name)
	if err != nil {
		fmt.Println("user doesn't exist")
		os.Exit(1)
	}

	if err = state.Config.SetUser(name); err != nil {
		return err
	}
	return nil
}

func HandlerRegister(state *config.State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return errors.New("Username is required")
	}

	user, err := state.Db.CreateUser(context.Background(), database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	})
	if err != nil {
		os.Exit(1)
	}

	state.Config.SetUser(cmd.Args[0])
	fmt.Printf("user was cread %s\n", cmd.Args[0])
	fmt.Printf("%+v\n", user)

	return nil
}

func HandlerReset(state *config.State, cmd Command) error {
	if err := state.Db.ResetUsers(context.Background()); err != nil {
		fmt.Println("failed to reser user table")
		os.Exit(1)
	}
	return nil
}

func HandlerListUsers(state *config.State, cmd Command) error {

	users, err := state.Db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("failed to get users")
		os.Exit(1)
	}

	loggedUser := state.Config.CurrentUserName

	for _, user := range users {
		if user.Name == loggedUser {
			fmt.Printf("*  %s (current)\n", loggedUser)
			continue
		}
		fmt.Println("* ", user.Name)
	}
	return nil
}
func parsePubDate(s string) (time.Time, error) {
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported pubDate format: %s", s)
}
func scrapeFeed(state *config.State) error {
	ctx := context.Background()

	feedToFetch, err := state.Db.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("get next feed to fetch: %w", err)
	}

	if err = state.Db.MarkFeedFetched(ctx, feedToFetch.ID); err != nil {
		return fmt.Errorf("mark feed fetched: %w", err)
	}

	feed, err := rss.FetchFeed(ctx, feedToFetch.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed %q: %w", feedToFetch.Url, err)
	}

	fmt.Printf("# %s\n", feed.Channel.Title)

	for _, item := range feed.Channel.Item {
		if len(item.Title) < 1 {
			continue
		}

		fmt.Printf(" - %s\n", item.Title)

		var publishedAt sql.NullTime
		if item.PubDate != "" {
			t, err := parsePubDate(item.PubDate)
			if err == nil {
				publishedAt = sql.NullTime{
					Time:  t,
					Valid: true,
				}
			}
		}

		_, err = state.Db.CreatePost(ctx, database.CreatePostParams{
			ID: uuid.New(),
			UpdatedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: publishedAt,
			FeedID:      feedToFetch.ID,
		})
		if err != nil {
			fmt.Println("failed to create post:", err)
		}
	}

	return nil
}
func HandlerAgg(state *config.State, cmd Command) error {

	if len(cmd.Args) < 1 {
		return errors.New("please provide interval: 1m, 1h, 1d")
	}

	//Ticker
	duration, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	fmt.Printf("Collecting feeds every %s\n", duration.String())

	//fetching loop
	for ; ; <-ticker.C {
		if err = scrapeFeed(state); err != nil {
			return err
		}
		//data, _ := json.MarshalIndent(feed, "", " ")
	}

	return nil
}

func HandlerAddFeed(state *config.State, cmd Command, user database.User) error {
	if len(cmd.Args) < 2 {
		return errors.New("to add feed you must provide name and url!")
	}

	name := cmd.Args[0]
	urlFeed := cmd.Args[1]

	/*rssObj, err := rss.FetchFeed(context.Background(), urlFeed)
	if err != nil {
		return err
	}*/

	feed, err := state.Db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		UpdatedAt: time.Now().String(),
		Name:      name,
		Url:       urlFeed,
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
	})

	if err != nil {
		return err
	}

	_, err = state.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}

	prettyData, err := json.MarshalIndent(feed, "", " ")

	fmt.Println("new feed record was added, assigned to ", user.Name)
	fmt.Printf("%#+v\n", string(prettyData))

	return nil
}

func HandlerGetFeeds(state *config.State, cmd Command) error {

	feeds, err := state.Db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, v := range feeds {

		feedOwner, err := state.Db.GetUserByID(context.Background(), v.UserID.UUID)
		if err != nil {
			return err
		}

		fmt.Printf("*  %s\n", v.Name)
		fmt.Printf("*  %s\n", v.Url)
		fmt.Printf("*  Owned by: %s\n", feedOwner.Name)
		fmt.Println("")
	}

	return nil
}

func HandlerFollow(state *config.State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return errors.New("url argument not found")
	}

	url := cmd.Args[0]

	selectedFeed, err := state.Db.GetFeed(context.Background(), url)
	if err != nil {
		return err
	}

	feed, err := state.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    selectedFeed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("feed: %s\n", feed.FeedName)
	fmt.Printf("username: %s\n", feed.UserName)

	return nil
}

func HandlerFollowing(state *config.State, cmd Command, user database.User) error {

	feedsUserFollows, err := state.Db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("%s follows:\n", user.Name)
	for _, entry := range feedsUserFollows {
		fmt.Printf(" *%s\n", entry.FeedName)
	}

	return nil
}

func HandlerUnfollow(state *config.State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return errors.New("url not found")
	}

	url := cmd.Args[0]

	feed, err := state.Db.GetFeed(context.Background(), url)
	if err != nil {
		return err
	}

	if err = state.Db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}); err != nil {
		return err
	}

	fmt.Printf("  *feed -> %s - %s unfollowed successfully\n", feed.Name, feed.Url)

	return nil
}

func HandlerBrowse(state *config.State, cmd Command, user database.User) error {
	limit := 2
	if len(cmd.Args) > 0 {
		limitConv, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return err
		}
		limit = limitConv
	}

	posts, err := state.Db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		Name:  user.Name,
		Limit: int32(limit),
	})
	if err != nil {
		return err
	}

	//fmt.Printf("%#+v\n", posts)

	for _, post := range posts {
		fmt.Printf(" #%s\n  *%s\n", post.Title, post.Url)
	}

	return nil
}

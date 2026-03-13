package rss

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RssItem `xml:"item"`
	} `xml:"channel"`
}

type RssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedUrl, nil)
	if err != nil {
		return &RSSFeed{}, err
	}

	req.Header.Set("User-Agent", "gator")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, err
	}

	var rssObj RSSFeed

	if err = xml.Unmarshal(body, &rssObj); err != nil {
		return &RSSFeed{}, err
	}

	rssObj.Channel.Title = html.UnescapeString(rssObj.Channel.Title)
	rssObj.Channel.Description = html.UnescapeString(rssObj.Channel.Description)

	for i := 0; i < len(rssObj.Channel.Item); i++ {
		rssObj.Channel.Item[i].Title = html.UnescapeString(rssObj.Channel.Item[i].Title)
		rssObj.Channel.Item[i].Description = html.UnescapeString(rssObj.Channel.Item[i].Description)
	}

	return &rssObj, nil
}

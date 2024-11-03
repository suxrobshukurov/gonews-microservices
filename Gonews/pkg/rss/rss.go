package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	strip "github.com/grokify/html-strip-tags-go"

	"github.com/suxrobshukurov/gonews/pkg/storage"
)

// RSS struct for main rss tag
type RSSFeed struct {
	Channel Channel `xml:"channel"`
}

// Channel struct for channel tag
type Channel struct {
	Items []Item `xml:"item"`
}

// Item struct for item tag
type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubTime     string `xml:"pubDate"`
}

// Parse Rss feeds
func ParseRSS(url string) ([]storage.Post, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("can't get %s: %w", url, err)
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read %s: %w", url, err)
	}
	var f RSSFeed
	err = xml.Unmarshal(b, &f)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal %s: %w", url, err)
	}

	var posts []storage.Post
	for _, item := range f.Channel.Items {
		var post storage.Post
		post.Title = item.Title
		post.Link = item.Link
		post.Content = strip.StripTags(item.Description)
		item.PubTime = strings.ReplaceAll(item.PubTime, ",", "")
		t, err := time.Parse("Mon 2 Jan 2006 15:04:05 -0700", item.PubTime)
		if err != nil {
			t, err = time.Parse("Mon 2 Jan 2006 15:04:05 GMT", item.PubTime)
		}
		if err == nil {
			post.PubTime = t.Unix()
		}
		posts = append(posts, post)
	}

	return posts, nil
}

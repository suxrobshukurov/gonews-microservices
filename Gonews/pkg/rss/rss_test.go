package rss

import (
	"testing"
)

func TestParseRSS(t *testing.T) {
	posts, err := ParseRSS("https://habr.com/ru/rss/hub/go/all/?fl=ru")
	if err != nil {
		t.Fatal(err)
	}

	if len(posts) == 0 {
		t.Fatal("no posts")
	}
	t.Log("Lenposts: ", len(posts))
}

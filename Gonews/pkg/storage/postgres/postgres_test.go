package postgres

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"math/rand"

	"github.com/stretchr/testify/assert"
	"github.com/suxrobshukurov/gonews/pkg/storage"
)

var testDB *DB

func TestMain(m *testing.M) {
	os.Setenv("connstr", "postgres://test_user:password@localhost:5432/for_testing?sslmode=disable")
	var err error
	testDB, err = New()
	if err != nil {
		panic(err)
	}
	defer testDB.pool.Close()

	testDB.pool.Exec(context.Background(), "DROP TABLE IF EXISTS posts")
	testDB.pool.Exec(context.Background(), `CREATE TABLE posts (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		pub_time INTEGER DEFAULT 0,
		link TEXT NOT NULL UNIQUE
	)`)

	m.Run()

}

func TestAddPosts(t *testing.T) {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	posts := []storage.Post{
		{
			Title:   "Test Post 1",
			Content: "Content for test post 1",
			PubTime: time.Now().Unix(),
			Link:    strconv.Itoa(r.Intn(1_000_000)),
		},
		{
			Title:   "Test Post 2",
			Content: "Content for test post 2",
			PubTime: time.Now().Unix(),
			Link:    strconv.Itoa(r.Intn(1_000_000)),
		},
	}
	err := testDB.AddPosts(posts)
	assert.NoError(t, err, "Should be able to add posts without errors")
}

func TestGetPosts(t *testing.T) {

	posts, err := testDB.Posts(0, 1)
	assert.NoError(t, err, "Should be able to retrieve posts without errors")
	assert.NotEmpty(t, posts, "Should retrieve at least one post")
}

func TestPostByID(t *testing.T) {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	newPost := storage.Post{
		Title:   "Unique Test Post",
		Content: "Unique content",
		PubTime: time.Now().Unix(),
		Link:    strconv.Itoa(r.Intn(1_000_000)),
	}

	err := testDB.AddPosts([]storage.Post{newPost})
	assert.NoError(t, err)

	retrievedPost, err := testDB.PostByID(3)
	assert.NoError(t, err)
	assert.Equal(t, newPost.Title, retrievedPost.Title, "Titles should match")
}

func TestFilter(t *testing.T) {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	testPosts := []storage.Post{
		{
			Title:   "Go Programming",
			Content: "Content for Go Programming",
			PubTime: time.Now().Unix(),
			Link:    strconv.Itoa(r.Intn(1_000_000)),
		},
		{
			Title:   "Golang Tips",
			Content: "Content for Golang Tips",
			PubTime: time.Now().Unix(),
			Link:    strconv.Itoa(r.Intn(1_000_000)),
		},
	}
	err := testDB.AddPosts(testPosts)
	assert.NoError(t, err)

	filteredPosts, err := testDB.Filter("Go", 0, 2)
	assert.NoError(t, err)
	assert.NotEmpty(t, filteredPosts, "Should retrieve posts matching the filter")
}

func TestCountPosts(t *testing.T) {

	count, err := testDB.Count()
	assert.NoError(t, err, "Should be able to get count without errors")
	assert.True(t, count >= 0, "Count should be non-negative")
}

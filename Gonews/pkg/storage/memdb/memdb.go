package memdb

import (
	"sync"

	"github.com/suxrobshukurov/gonews/pkg/storage"
)

type DB struct {
	m     sync.Mutex
	id    int
	store map[int]storage.Post
}

// New creates a new memdb storage
func New() (*DB, error) {
	db := DB{
		id:    1,
		store: make(map[int]storage.Post),
	}
	return &db, nil
}

// Posts returns a list of posts from the database
// n is the number of posts to return
func (db *DB) Posts(offset int, limit int) ([]storage.Post, error) {
	db.m.Lock()
	defer db.m.Unlock()
	var posts []storage.Post
	for _, p := range db.store {
		posts = append(posts, p)
	}
	return posts, nil
}

// PostByID returns a post by its ID
func (db *DB) PostByID(id int) (storage.Post, error) {
	db.m.Lock()
	defer db.m.Unlock()
	return db.store[id], nil
}

// AddPosts adds a list of posts to the database
func (db *DB) AddPosts(posts []storage.Post) error {
	db.m.Lock()
	defer db.m.Unlock()
	for _, p := range posts {
		p.ID = db.id
		db.store[p.ID] = p
		db.id++
	}
	return nil
}

// Filter returns a filtered list of posts
func (db *DB) Filter(searchStr string, offset int, limit int) ([]storage.Post, error) {
	db.m.Lock()
	defer db.m.Unlock()
	var posts []storage.Post
	for _, p := range db.store {
		if p.Title == searchStr {
			posts = append(posts, p)
		}
	}
	return posts, nil
}

// Count returns the total number of posts
func (db *DB) Count() (int, error) {
	db.m.Lock()
	defer db.m.Unlock()
	return len(db.store), nil
}

// CountOfFilter returns the count of posts matching a search pattern
func (db *DB) CountOfFilter(str string) (int, error) {
	db.m.Lock()
	defer db.m.Unlock()
	var count int
	for _, p := range db.store {
		if p.Title == str {
			count++
		}
	}
	return count, nil
}

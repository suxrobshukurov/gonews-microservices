package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/subosito/gotenv"
	"github.com/suxrobshukurov/gonews/pkg/storage"
)

func init() {
	gotenv.Load()
}

type DB struct {
	pool *pgxpool.Pool
}

func New() (*DB, error) {
	connstr := os.Getenv("connstr")
	if connstr == "" {
		return nil, errors.New("empty connstr")
	}

	p, err := pgxpool.Connect(context.Background(), connstr)
	if err != nil {
		return nil, err
	}
	return &DB{pool: p}, nil
}


// Posts retrieves posts from the database with context support
// offset is the number of records to skip
// limit is the number of records to return
// returns a slice of Post and an error if any
func (db *DB) Posts(offset int, limit int) ([]storage.Post, error) {
	rows, err := db.pool.Query(context.Background(), `
		SELECT id, title, content, pub_time, link
		FROM posts
		ORDER BY pub_time DESC
		OFFSET $1 LIMIT $2
	`, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("can't get posts from db: %w", err)
	}
	defer rows.Close()

	var posts []storage.Post
	for rows.Next() {
		var post storage.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.PubTime, &post.Link); err != nil {
			return nil, fmt.Errorf("can't scan post: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

// PostByID retrieves a post by its id with context support
// returns a Post and an error if any
func (db *DB) PostByID(id int) (storage.Post, error) {
	var post storage.Post
	err := db.pool.QueryRow(context.Background(), `
		SELECT id, title, content, pub_time, link
		FROM posts
		WHERE id = $1
	`, id).Scan(&post.ID, &post.Title, &post.Content, &post.PubTime, &post.Link)
	if err != nil {
		return storage.Post{}, fmt.Errorf("can't get post by id from db: %w", err)
	}
	return post, nil
}


// AddPosts adds a list of posts to the database. It will upsert the posts if they
// already exist in the database, updating the title, content, and pub_time fields.
//
func (db *DB) AddPosts(posts []storage.Post) error {
	for _, p := range posts {
		_, err := db.pool.Exec(context.Background(), `
			INSERT INTO posts (title, content, pub_time, link)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (link) DO UPDATE
			SET title = EXCLUDED.title, content = EXCLUDED.content, pub_time = EXCLUDED.pub_time
		`, p.Title, p.Content, p.PubTime, p.Link)
		if err != nil {
			return fmt.Errorf("can't insert post in db: %w", err)
		}
	}
	return nil
}

// Filter retrieves posts matching a search pattern with context support
func (db *DB) Filter(searchStr string, offset int, limit int) ([]storage.Post, error) {
	searchPattern := "%" + searchStr + "%"
	rows, err := db.pool.Query(context.Background(), `
		SELECT id, title, content, pub_time, link
		FROM posts
		WHERE title ILIKE $1
		ORDER BY pub_time DESC
		OFFSET $2 LIMIT $3
	`, searchPattern, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve filtered posts from db: %w", err)
	}
	defer rows.Close()

	var posts []storage.Post
	for rows.Next() {
		var post storage.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.PubTime, &post.Link); err != nil {
			return nil, fmt.Errorf("unable to scan post row: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

// Count returns the total number of posts with context support
func (db *DB) Count() (int, error) {
	var count int
	err := db.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) AS total_rows FROM posts
	`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("can't get count from db: %w", err)
	}
	return count, nil
}

// CountOfFilter returns the count of posts matching a search pattern
// with context support.
func (db *DB) CountOfFilter(str string) (int, error) {
	str = "%" + str + "%"
	var count int
	err := db.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) AS total_rows FROM posts
		WHERE title ILIKE $1
	`, str).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("can't get count from db: %w", err)
	}
	return count, nil
}

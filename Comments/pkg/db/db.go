package db

import (
	"Comments/pkg/models"
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/subosito/gotenv"
)

type DB struct {
	pool *pgxpool.Pool
}

// Load .env
func init() {
	gotenv.Load()
}


// New creates a new DB object, or returns an error if the connstr is empty.
// New calls pgxpool.Connect to connect to the database.
func New() (*DB, error) {
	connstr := os.Getenv("connstr")
	if connstr == "" {
		return nil, errors.New("empty connstr")
	}
	p, err := pgxpool.Connect(context.Background(), connstr)
	if err != nil {
		return nil, err
	}
	db := &DB{
		pool: p,
	}
	return db, nil
}


// AddComment adds a comment to the database.
//
// AddComment takes a Comment object as argument and will
// return the generated id of the comment and an error if any.
func (db *DB) AddComment(c models.Comment) (int, error) {

	var id int
	err := db.pool.QueryRow(context.Background(),
		"INSERT INTO comments (post_id, parent_id, content, add_time) VALUES ($1, $2, $3, $4) RETURNING id",
		c.PostID, c.ParentID, c.Content, c.AddTime).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}


// Comments retrieves all comments for a post.
//
// Comments takes an int as argument and will return a slice of Comment
// objects and an error if any.
func (db *DB) Comments(id int) ([]models.Comment, error) {

	rows, err := db.pool.Query(context.Background(), "SELECT id, post_id, parent_id, content, add_time FROM comments WHERE post_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.ParentID, &c.Content, &c.AddTime); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return buildCommentTree(comments), nil
}


// UpdateComment updates a comment in the database.
//
// UpdateComment takes a Comment object as argument and will
// return an error if any.
func (db *DB) UpdateComment(c models.Comment) error {
	_, err := db.pool.Exec(context.Background(),
		"UPDATE comments SET post_id = $1, parent_id = $2, content = $3, add_time = $4 WHERE id = $5",
		c.PostID, c.ParentID, c.Content, c.AddTime, c.ID)
	return err
}


// DeleteComment deletes a comment from the database.
//
// DeleteComment takes a comment ID as argument and will
// return an error if any.
func (db *DB) DeleteComment(id int) error {
	_, err := db.pool.Exec(context.Background(), "DELETE FROM comments WHERE id = $1", id)
	return err
}
func buildCommentTree(comments []models.Comment) []models.Comment {
	childrenMap := make(map[int][]models.Comment)
	var roots []models.Comment

	for _, comment := range comments {
		if comment.ParentID == 0 {
			roots = append(roots, comment)
		} else {
			childrenMap[comment.ParentID] = append(childrenMap[comment.ParentID], comment)
		}
	}

	// iterative method
	stack := make([]*models.Comment, len(roots))
	for i := range roots {
		stack[i] = &roots[i]
	}

	for i := range roots {
		stack[i] = &roots[i]
	}

	for len(stack) > 0 {
		parent := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		parent.Replies = childrenMap[parent.ID]
		for i := range parent.Replies {
			stack = append(stack, &parent.Replies[i])
		}
	}

	// recursive method
	// var attachReplies func(*models.Comment)
	// attachReplies = func(parent *models.Comment) {
	// 	for _, child := range childrenMap[parent.ID] {
	// 		parent.Replies = append(parent.Replies, child)
	// 		attachReplies(&parent.Replies[len(parent.Replies)-1])
	// 	}
	// }

	// for i := range roots {
	// 	attachReplies(&roots[i])
	// }

	return roots
}

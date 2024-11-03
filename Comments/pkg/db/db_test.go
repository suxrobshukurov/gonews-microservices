package db

import (
	"Comments/pkg/models"
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	testDB.pool.Exec(context.Background(), "DROP TABLE IF EXISTS comments")
	testDB.pool.Exec(context.Background(), `CREATE TABLE comments (
		id SERIAL PRIMARY KEY, 
		post_id BIGINT NOT NULL, 
		parent_id BIGINT DEFAULT 0, 
		content TEXT NOT NULL, 
		add_time BIGINT NOT NULL
		)`)

	m.Run()
}

func TestAddComment(t *testing.T) {
	comment := models.Comment{
		PostID:   1,
		ParentID: 0,
		Content:  "test comment",
		AddTime:  time.Now().Unix(),
	}
	id, err := testDB.AddComment(comment)

	assert.NoError(t, err)
	assert.NotZero(t, id, "ID should not be zero")
}

func TestGetComments(t *testing.T) {
	comment := models.Comment{
		PostID:   1,
		ParentID: 0,
		Content:  "test comment",
		AddTime:  time.Now().Unix(),
	}
	_, err := testDB.AddComment(comment)
	assert.NoError(t, err)

	comments, err := testDB.Comments(1)
	assert.NoError(t, err)
	assert.NotEmpty(t, comments, "Should have at least one comment")
	assert.Equal(t, comment.PostID, comments[0].PostID)
	assert.Equal(t, comment.Content, comments[0].Content)
}

func TestUpdateComment(t *testing.T) {
	comment := models.Comment{
		PostID:   2,
		ParentID: 0,
		Content:  "test comment",
		AddTime:  time.Now().Unix(),
	}
	id, err := testDB.AddComment(comment)
	assert.NoError(t, err)

	comment.Content = "updated comment"
	comment.ID = id
	err = testDB.UpdateComment(comment)
	assert.NoError(t, err)

	comments, err := testDB.Comments(2)
	assert.NoError(t, err)
	assert.NotEmpty(t, comments, "Should have at least one comment")
	assert.Equal(t, comment.Content, comments[0].Content)
}

func TestDeleteComment(t *testing.T) {
	comment := models.Comment{
		PostID:   3,
		ParentID: 0,
		Content:  "test comment",
		AddTime:  time.Now().Unix(),
	}
	id, err := testDB.AddComment(comment)
	assert.NoError(t, err)

	err = testDB.DeleteComment(id)
	assert.NoError(t, err)

	comments, err := testDB.Comments(3)
	assert.NoError(t, err)
	assert.Empty(t, comments, "Comments should be empty after deletion")
}

func Test_buildCommentTree(t *testing.T) {
	type args struct {
		comments []models.Comment
	}
	tests := []struct {
		name string
		args args
		want []models.Comment
	}{
		{
			name: "Test with one root comments",
			args: args{
				comments: []models.Comment{
					{
						ID:       1,
						PostID:   1,
						ParentID: 0,
						Content:  "test comment",
						AddTime:  time.Now().Unix(),
					},
					{
						ID:       2,
						PostID:   1,
						ParentID: 1,
						Content:  "test comment",
						AddTime:  time.Now().Unix(),
					},
					{
						ID:       3,
						PostID:   1,
						ParentID: 2,
						Content:  "test comment",
						AddTime:  time.Now().Unix(),
					},
				},
			},
			want: []models.Comment{
				{
					ID:       1,
					PostID:   1,
					ParentID: 0,
					Content:  "test comment",
					AddTime:  time.Now().Unix(),
					Replies: []models.Comment{
						{
							ID:       2,
							PostID:   1,
							ParentID: 1,
							Content:  "test comment",
							AddTime:  time.Now().Unix(),
							Replies: []models.Comment{
								{
									ID:       3,
									PostID:   1,
									ParentID: 2,
									Content:  "test comment",
									AddTime:  time.Now().Unix(),
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildCommentTree(tt.args.comments); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accumulateComments() = \n%v\n, want \n%v\n", got, tt.want)
			}
		})
	}
}

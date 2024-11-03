package models

type Comment struct {
	ID       int       `json:"ID"`
	PostID   int       `json:"PostID"`
	ParentID int       `json:"ParentID"`
	Content  string    `json:"Content"`
	AddTime  int64     `json:"AddTime"`
	Replies  []Comment `json:"Replies"`
}

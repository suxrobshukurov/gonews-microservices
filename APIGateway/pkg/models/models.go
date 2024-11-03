package models

type PostFullDetailed struct {
	ID       int       `json:"ID"`
	Title    string    `json:"Title"`
	Content  string    `json:"Content"`
	PubTime  int64     `json:"PubTime"`
	Link     string    `json:"Link"`
	Comments []Comment `json:"Comments"`
}

type NewsShortDetailed struct {
	ID      int    `json:"ID"`
	Title   string `json:"Title"`
	PubTime int64  `json:"PubTime"`
	Link    string `json:"Link"`
}

type Comment struct {
	ID       int       `json:"ID"`
	PostID   int       `json:"PostID"`
	ParentID int       `json:"ParentID"`
	Content  string    `json:"Content"`
	AddTime  int64     `json:"AddTime"`
	Visible  bool      `json:"Visible"`
	Replies  []Comment `json:"Replies"`
}

package storage

// Post represents a single post
type Post struct {
	ID      int
	Title   string
	Content string
	PubTime int64
	Link    string
}

// Interface represents a storage
type Interface interface {
	Posts(int, int) ([]Post, error)
	PostByID(int) (Post, error)
	AddPosts([]Post) error
	Filter(string, int, int) ([]Post, error)
	Count() (int, error)
	CountOfFilter(string) (int, error)
}

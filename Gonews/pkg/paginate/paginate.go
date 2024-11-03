package paginate

import "github.com/suxrobshukurov/gonews/pkg/storage"


const PostsPerPage = 10

type Paginate struct {
	Posts []storage.Post
	Pagination Pagination
}

type Pagination struct {
	CurrentPage   int
	TotalPages    int
	NumberOfPosts int
}

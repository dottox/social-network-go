package model

type Post struct {
	Id            uint32   `json:"id"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	UserId        uint32   `json:"user_id"`
	Tags          []string `json:"tags"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	CommentsCount uint16   `json:"comments_count"`
}

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type UpdatePostPayload struct {
	Title   string   `json:"title" validate:"max=100"`
	Content string   `json:"content" validate:"max=1000"`
	Tags    []string `json:"tags"`
}

package model

/*
CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGSERIAL NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id BIGSERIAL NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
*/

type Comment struct {
	Id        uint32 `json:"id"`
	UserId    uint32 `json:"user_id"`
	PostId    uint32 `json:"post_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type CreateCommentPayload struct {
	Content string `json:"content" validate:"required,max=500"`
}

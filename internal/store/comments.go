package store

import (
	"context"
	"database/sql"

	"github.com/dottox/social/internal/model"
)

type CommentStore struct {
	db *sql.DB
}

func (s *CommentStore) Create(ctx context.Context, comment *model.Comment) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Create the query to insert the comment
	query := `
		INSERT INTO comments (user_id, post_id, content)
		VALUES ($1, $2, $3) RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	// Send the query with the context and arguments
	err = s.db.QueryRowContext(
		ctx,
		query,
		comment.UserId,
		comment.PostId,
		comment.Content,
	).Scan( // Scan the post to insert the generated values
		&comment.Id,
		&comment.CreatedAt,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Update the comments_count in the posts table
	updateQuery := `
		UPDATE posts
		SET comments_count = comments_count + 1
		WHERE id = $1
	`

	_, err = s.db.ExecContext(ctx, updateQuery, comment.PostId)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// Get comments by their postId
func (s *CommentStore) GetAllByPostId(ctx context.Context, postId uint32) ([]*model.Comment, error) {

	exists, err := postExists(ctx, s.db, postId)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, ErrResourceNotFound
	}

	// Create the query to get the comment by the id
	query := `
		SELECT id, user_id, post_id, content, created_at
		FROM comments
		WHERE post_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	// Perform the query with the ctx and id
	// Scan all the data to the blank comment
	rows, err := s.db.QueryContext(ctx, query, postId)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}
	defer rows.Close()

	// Create a new list of comments
	comments := []*model.Comment{}

	// Iterate over the rows
	for rows.Next() {
		// Create a new blank comment
		var comment model.Comment

		// Scan the row to the blank comment
		err := rows.Scan(
			&comment.Id,
			&comment.UserId,
			&comment.PostId,
			&comment.Content,
			&comment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Append the comment to the list
		comments = append(comments, &comment)
	}

	// Return the comments
	return comments, nil
}

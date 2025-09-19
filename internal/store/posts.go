package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/dottox/social/internal/model"
	"github.com/lib/pq"
)

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *model.Post) error {
	// Create the query to insert the post
	query := `
		INSERT INTO posts (title, content, user_id, tags)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`

	// Send the query with the context and arguments
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.UserId,
		pq.Array(post.Tags),
	).Scan( // Scan the post to insert the generated values
		&post.Id,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetById(ctx context.Context, id uint32) (*model.Post, error) {
	// Create the query to get the post by the id
	query := `
		SELECT id, title, content, user_id, tags, created_at, updated_at, comments_count
		FROM posts
		WHERE id = $1
	`

	// Create a new blank post
	var post model.Post

	// Perform the query with the ctx and id
	// Scan all the data to the blank post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.Id,
		&post.Title,
		&post.Content,
		&post.UserId,
		pq.Array(&post.Tags), // note: tags is a slice, so use pq.Array()
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.CommentsCount,
	)
	if err != nil {
		switch {
		// If no rows found, return a ErrResourceNotFound
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}

	// Return the post without erors
	return &post, nil
}

func (s *PostStore) Update(ctx context.Context, newPost *model.Post) (*model.Post, error) {

	// Build the set string for the update query
	setString := _createPostUpdateSetString(newPost)
	if setString == "" {
		return nil, errors.New("no fields to update")
	}

	// Create the query to get the post by the id
	updateQuery := `
		UPDATE posts
		SET ` + setString + `, updated_at = NOW()
		WHERE id = $1
	`

	// Perform the query with the ctx and id
	_, err := s.db.ExecContext(ctx, updateQuery, newPost.Id)
	if err != nil {
		switch {
		// If no rows found, return a ErrResourceNotFound
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}

	// Get the updated post
	updatedPost, err := s.GetById(ctx, newPost.Id)
	if err != nil {
		return nil, err
	}

	// Return no errors
	return updatedPost, err
}

func (s *PostStore) DeleteById(ctx context.Context, id uint32) (*model.Post, error) {

	// Get the post to return it after deletion
	deletedPost, err := s.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create the query to delete the post by the id
	query := `
		DELETE FROM posts
		WHERE id = $1
	`

	// Perform the query with the ctx and id
	_, err = s.db.ExecContext(ctx, query, id)
	if err != nil {
		switch {
		// If no rows found, return a ErrResourceNotFound
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrResourceNotFound
		default:
			return nil, err
		}
	}

	// Return no errors
	return deletedPost, nil
}

func postExists(ctx context.Context, db *sql.DB, postId uint32) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS (
			SELECT 1 FROM posts WHERE id = $1
		)
	`

	err := db.QueryRowContext(ctx, query, postId).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func _createPostUpdateSetString(post *model.Post) string {
	updatedParts := []string{}
	if post.Title != "" {
		updatedParts = append(updatedParts, fmt.Sprintf("title = '%s'", post.Title))
	}

	if post.Content != "" {
		updatedParts = append(updatedParts, fmt.Sprintf("content = '%s'", post.Content))
	}

	if post.Tags != nil {
		if len(post.Tags) == 0 {
			updatedParts = append(updatedParts, "tags = ARRAY[]::VARCHAR[]")
		} else {
			tagsArr := strings.Join(post.Tags, "','")
			updatedParts = append(updatedParts, fmt.Sprintf("tags = ARRAY['%s']", tagsArr))
		}
	}

	return strings.Join(updatedParts, ", ")
}

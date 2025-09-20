package db

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/dottox/social/internal/model"
	"github.com/dottox/social/internal/store"
)

const (
	USER_COUNT    = 100
	POST_COUNT    = 200
	COMMENT_COUNT = 500
)

func Seed(store store.Storage) error {
	ctx := context.Background()

	log.Printf("seeding %d users", USER_COUNT)
	users := generateUsers(USER_COUNT)
	for _, user := range users {
		if err := store.Users.Create(ctx, user); err != nil {
			log.Printf("failed to create user: %v", err)
			return err
		}

		fmt.Printf("Seeded user: %+v\n", user)
	}
	log.Print("seeded users correctly")

	log.Printf("seeding %d posts into users", POST_COUNT)
	posts := generatePosts(POST_COUNT, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Printf("failed to create post: %v", err)
			return err
		}

		fmt.Printf("Seeded post: %+v\n", post)
	}
	log.Print("seeded posts correctly")

	log.Printf("seeding %d comments into posts", COMMENT_COUNT)
	comments := generateComments(COMMENT_COUNT, users, posts)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			log.Printf("failed to create comment: %v", err)
			return err
		}

		fmt.Printf("Seeded comment: %+v\n", comment)

	}
	log.Print("seeded comments correctly")

	log.Print("seeded completed!")
	return nil
}

func generateUsers(count int) []*model.User {
	users := make([]*model.User, count)
	for i := 0; i < count; i++ {
		users[i] = &model.User{
			Username: generateRandomString(8),
			Email:    generateRandomEmail(),
			Password: "password123",
		}
	}
	return users
}

func generatePosts(count int, users []*model.User) []*model.Post {
	posts := make([]*model.Post, count)
	for i := 0; i < count; i++ {
		tags := []string{}
		for t := 0; t < rand.Intn(4); t++ {
			tags = append(tags, generateRandomString(4))
		}

		user := users[rand.Intn(len(users))]

		posts[i] = &model.Post{
			UserId:  user.Id,
			Title:   generateRandomString(8) + " " + generateRandomString(8),
			Content: generateRandomString(50),
			Tags:    tags,
		}
	}
	return posts
}

func generateComments(count int, users []*model.User, posts []*model.Post) []*model.Comment {
	comments := make([]*model.Comment, count)
	for i := 0; i < count; i++ {
		comments[i] = &model.Comment{
			UserId:  users[rand.Intn(len(users))].Id,
			PostId:  posts[rand.Intn(len(posts))].Id,
			Content: generateRandomString(30),
		}
		time.Sleep(time.Nanosecond) // ensure different seeds

	}

	return comments
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func generateRandomEmail() string {
	return generateRandomString(30) + "@" + generateRandomString(10) + "." + generateRandomString(3)
}

package posts

import (
	"blog/db/tags"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Post struct {
	Id       int
	AuthorId int
	Likes    int
	Comments int
	Author   string
	Title    string
	Text     string
	Tags     []tags.Tag
	Created  time.Time
}

type Posts struct {
	Posts  []Post
	Nposts int
}

func AddPost(db *sql.DB, ctx context.Context, post Post) (int, error) {
	if post.Title == "" || post.Text == "" {
		return 0, fmt.Errorf("invalid argument")
	}
	var postId int
	err := db.QueryRowContext(
		ctx,
		`INSERT INTO posts (title, text, user_id)
			VALUES ($1, $2, $3) RETURNING id`,
		post.Title, post.Text, post.AuthorId,
	).Scan(&postId)
	if err != nil {
		return 0, err
	}
	return postId, nil
}

func GetPosts(db *sql.DB, ctx context.Context) ([]Post, error) {
	rows, err := db.QueryContext(ctx, "SELECT * FROM post_view;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.Id, &post.Title, &post.Text, &post.Created,
			&post.Author, &post.Likes, &post.Comments)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func GetPost(db *sql.DB, ctx context.Context, id int) (Post, error) {
	var post Post
	err := db.QueryRowContext(ctx, "SELECT * FROM posts WHERE id = $1", id).Scan(
		&post.Id, &post.Title, &post.Text, &post.AuthorId, &post.Created,
	)
	if err != nil {
		return Post{}, err
	}
	return post, nil
}

func UpdatePost(db *sql.DB, ctx context.Context, id int, post Post) error {
	if post.Title == "" || post.Text == "" {
		return fmt.Errorf("invalid argument")
	}
	_, err := db.ExecContext(ctx,
		"UPDATE posts SET title = $1, text = $2 WHERE id = $3",
		post.Title, post.Text, id)
	if err != nil {
		return err
	}
	return nil
}

func DeletePost(db *sql.DB, ctx context.Context, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func FilterTag(db *sql.DB, ctx context.Context, tag tags.Tag) ([]Post, error) {
	if tag.Name == "" {
		return nil, fmt.Errorf("invalid argument")
	}
	rows, err := db.QueryContext(ctx,
		`SELECT post_view.* FROM post_view 
			JOIN post_tags ON post_view.id = post_tags.post_id
			JOIN tags ON post_tags.tag_id = tags.id WHERE tags.name = $1`, tag.Name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.Id, &post.Title, &post.Text, &post.Created,
			&post.Author, &post.Likes, &post.Comments)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func FilterQuery(db *sql.DB, ctx context.Context, query string) ([]Post, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT * FROM post_view WHERE title LIKE $1", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.Id, &post.Title, &post.Text, &post.Created,
			&post.Author, &post.Likes, &post.Comments)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

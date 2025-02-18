package comments

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Comment struct {
	Id       int
	AuthorId int
	PostId   int
	Text     string
	Author   string
	Created  time.Time
}

func AddComment(db *sql.DB, ctx context.Context, comment Comment) error {
	if comment.AuthorId == 0 || comment.PostId == 0 || comment.Text == "" {
		return fmt.Errorf("invalid argument")
	}
	_, err := db.ExecContext(ctx,
		`INSERT INTO comments (user_id, post_id, text)
			VALUES ($1, $2, $3)`,
		comment.AuthorId, comment.PostId,
		comment.Text)
	if err != nil {
		return err
	}
	return nil
}

func GetComments(db *sql.DB, ctx context.Context, postId int) ([]Comment, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT comments.*, users.username FROM comments
			JOIN users ON comments.user_id = users.id
			WHERE comments.post_id = $1
			ORDER BY comments.created DESC`,
		postId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err = rows.Scan(
			&comment.Id, &comment.PostId, &comment.AuthorId,
			&comment.Text, &comment.Created, &comment.Author,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func GetComment(db *sql.DB, ctx context.Context, id int) (Comment, error) {
	var comment Comment
	err := db.QueryRowContext(
		ctx,
		`SELECT comments.*, users.username FROM comments
			JOIN users ON comments.user_id = users.id
			WHERE comments.id = $1`,
		id,
	).Scan(
		&comment.Id, &comment.PostId, &comment.AuthorId,
		&comment.Text, &comment.Created, &comment.Author,
	)
	if err != nil {
		return Comment{}, err
	}
	return comment, nil
}

func UpdateComment(db *sql.DB, ctx context.Context, id int, comment Comment) error {
	if comment.Text == "" {
		return fmt.Errorf("invalid argument")
	}
	_, err := db.ExecContext(ctx, "UPDATE comments SET text = $1 WHERE id = $2",
		comment.Text, id)
	if err != nil {
		return err
	}
	return nil
}

func DeleteComment(db *sql.DB, ctx context.Context, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM comments WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

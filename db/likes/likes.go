package likes

import (
	"context"
	"database/sql"
)

func AddLike(db *sql.DB, ctx context.Context, userId, postId int, likeType string) error {
	var dbType string
	err := db.QueryRowContext(
		ctx,
		"SELECT type FROM likes WHERE user_id = $1 AND post_id = $2",
		userId, postId,
	).Scan(&dbType)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	var query string
	if err == sql.ErrNoRows {
		query = "INSERT INTO likes (user_id, post_id, type) VALUES ($1, $2, $3)"
		_, err = db.ExecContext(ctx, query, userId, postId, likeType)
	} else if likeType == dbType {
		query = "DELETE FROM likes WHERE user_id = $1 AND post_id = $2 AND type = $3"
		_, err = db.ExecContext(ctx, query, userId, postId, likeType)
	} else {
		query = "UPDATE likes SET type = $1 WHERE user_id = $2 AND post_id = $3"
		_, err = db.ExecContext(ctx, query, likeType, userId, postId)
	}
	if err != nil {
		return err
	}
	return nil
}

func GetLikes(db *sql.DB, ctx context.Context, id int) (int, error) {
	var likes int
	err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FILTER (WHERE type = 'like') -
			COUNT(*) FILTER (WHERE type = 'dislike')
			FROM likes WHERE post_id = $1`,
		id,
	).Scan(&likes)
	if err != nil {
		return 0, err
	}
	return likes, nil
}

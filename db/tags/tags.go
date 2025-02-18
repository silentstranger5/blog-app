package tags

import (
	"context"
	"database/sql"
	"fmt"
)

type Tag struct {
	Name string
}

func AddTags(db *sql.DB, ctx context.Context, postId int, tags []Tag) error {
	if tags == nil {
		return fmt.Errorf("invalid argument")
	}
	for _, tag := range tags {
		if tag.Name == "" {
			return fmt.Errorf("invalid argument")
		}
	}
	insertStmt, err := db.PrepareContext(ctx,
		"INSERT INTO tags (name) VALUES ($1) ON CONFLICT(name) DO NOTHING")
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	for _, tag := range tags {
		_, err = insertStmt.ExecContext(ctx, tag.Name)
		if err != nil {
			return err
		}
	}

	selectStmt, err := db.PrepareContext(ctx, "SELECT id FROM tags WHERE name = $1")
	if err != nil {
		return err
	}
	defer selectStmt.Close()

	tagIds := make([]int, 0)
	for _, tag := range tags {
		var tagId int
		err = selectStmt.QueryRow(tag.Name).Scan(&tagId)
		if err != nil {
			return err
		}
		tagIds = append(tagIds, tagId)
	}

	for _, tagId := range tagIds {
		_, err := db.Exec("INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2)", postId, tagId)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetTags(db *sql.DB, ctx context.Context, id int) ([]Tag, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT tags.name FROM tags
			LEFT JOIN post_tags ON tags.id = post_tags.tag_id
			LEFT JOIN posts ON post_tags.post_id = posts.id
			WHERE posts.id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query on database: %v", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func UpdateTags(db *sql.DB, ctx context.Context, id int, tags []Tag) error {
	if tags == nil {
		return fmt.Errorf("invalid argument")
	}
	for _, tag := range tags {
		if tag.Name == "" {
			return fmt.Errorf("invalid argument")
		}
	}
	err := DeleteTags(db, ctx, id)
	if err != nil {
		return err
	}
	err = AddTags(db, ctx, id, tags)
	if err != nil {
		return err
	}
	return nil
}

func DeleteTags(db *sql.DB, ctx context.Context, id int) error {
	_, err := db.ExecContext(
		ctx,
		"DELETE FROM post_tags WHERE post_id = $1", id,
	)
	if err != nil {
		return err
	}
	return nil
}

package images

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Image struct {
	Id       int
	AuthorId int
	Name     string
	Created  time.Time
}

func AddImage(db *sql.DB, ctx context.Context, image Image) error {
	if image.AuthorId == 0 || image.Name == "" {
		return fmt.Errorf("invalid argument")
	}
	_, err := db.ExecContext(ctx,
		"INSERT OR IGNORE INTO images (name, user_id) VALUES ($1, $2)",
		image.Name, image.AuthorId)
	if err != nil {
		return err
	}
	return nil
}

func GetImages(db *sql.DB, ctx context.Context) ([]Image, error) {
	rows, err := db.QueryContext(ctx, "SELECT * FROM images")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	images := make([]Image, 0)
	for rows.Next() {
		var image Image
		err = rows.Scan(&image.Id, &image.AuthorId, &image.Name, &image.Created)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, nil
}

func GetImage(db *sql.DB, ctx context.Context, id int) (Image, error) {
	var image Image
	err := db.QueryRowContext(ctx,
		"SELECT * FROM images WHERE id = $1", id,
	).Scan(&image.Id, &image.AuthorId, &image.Name, &image.Created)
	if err != nil {
		return Image{}, err
	}
	return image, nil
}

func DeleteImage(db *sql.DB, ctx context.Context, id int) error {
	_, err := db.ExecContext(ctx,
		"DELETE FROM images WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

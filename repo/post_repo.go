package repo

import (
	"database/sql"
	"weblog-app/model"
)

type PostRepository interface {
	Create(post *model.Post, sharedUserIDs []int) error
	GetByID(id int) (*model.Post, error)
	GetAccessiblePosts(userID int) ([]model.Post, error)
	HasAccess(postID int, userID int) (bool, error)
	Delete(id int, authorID int) error
}

type postRepo struct {
	db *sql.DB
}

func NewPostRepo(db *sql.DB) PostRepository {
	return &postRepo{db: db}
}

func (r *postRepo) Create(post *model.Post, sharedUserIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO posts (author_id, title, content, image_url, is_private)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	
	err = tx.QueryRow(query, post.AuthorID, post.Title, post.Content, post.ImageURL, post.IsPrivate).
		Scan(&post.ID, &post.CreatedAt)
	if err != nil {
		return err
	}

	if post.IsPrivate && len(sharedUserIDs) > 0 {
		shareQuery := `INSERT INTO post_shares (post_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
		for _, uID := range sharedUserIDs {
			if _, err := tx.Exec(shareQuery, post.ID, uID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *postRepo) GetByID(id int) (*model.Post, error) {
	query := `
		SELECT p.id, p.author_id, u.username, p.title, p.content, COALESCE(p.image_url, ''), p.is_private, p.created_at
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.id = $1`
	var p model.Post
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.AuthorID, &p.AuthorName, &p.Title, &p.Content, &p.ImageURL, &p.IsPrivate, &p.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *postRepo) GetAccessiblePosts(userID int) ([]model.Post, error) {
	query := `
		SELECT DISTINCT p.id, p.author_id, u.username, p.title, p.content, COALESCE(p.image_url, ''), p.is_private, p.created_at
		FROM posts p
		JOIN users u ON p.author_id = u.id
		LEFT JOIN post_shares ps ON p.id = ps.post_id
		WHERE p.is_private = FALSE 
		   OR p.author_id = $1 
		   OR ps.user_id = $1
		ORDER BY p.created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.AuthorName, &p.Title, &p.Content, &p.ImageURL, &p.IsPrivate, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (r *postRepo) HasAccess(postID int, userID int) (bool, error) {
	query := `
		SELECT COUNT(1)
		FROM posts p
		LEFT JOIN post_shares ps ON p.id = ps.post_id
		WHERE p.id = $1 AND (p.is_private = FALSE OR p.author_id = $2 OR ps.user_id = $2)`

	var count int
	err := r.db.QueryRow(query, postID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *postRepo) Delete(id int, authorID int) error {
	query := `DELETE FROM posts WHERE id = $1 AND author_id = $2`
	res, err := r.db.Exec(query, id, authorID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
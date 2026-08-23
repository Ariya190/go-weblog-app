package repo

import (
	"database/sql"
	"weblog-app/model"
)

type CommentRepository interface {
	Create(comment *model.Comment) error
	GetByPostID(postID int) ([]model.Comment, error)
}

type commentRepo struct {
	db *sql.DB
}

func NewCommentRepo(db *sql.DB) CommentRepository {
	return &commentRepo{db: db}
}

func (r *commentRepo) Create(comment *model.Comment) error {
	query := `
		INSERT INTO comments (post_id, author_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRow(query, comment.PostID, comment.AuthorID, comment.Content).
		Scan(&comment.ID, &comment.CreatedAt)
}

func (r *commentRepo) GetByPostID(postID int) ([]model.Comment, error) {
	query := `
		SELECT c.id, c.post_id, c.author_id, u.username, c.content, c.created_at
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.post_id = $1
		ORDER BY c.created_at ASC`

	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.AuthorName, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}
package model

import "time"

type Comment struct {
	ID         int
	PostID     int
	AuthorID   int
	AuthorName string
	Content    string
	CreatedAt  time.Time
}
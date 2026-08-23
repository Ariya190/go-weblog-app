package model

import "time"

type Post struct {
	ID         int
	AuthorID   int
	AuthorName string
	Title      string
	Content    string
	ImageURL   string
	IsPrivate  bool
	CreatedAt  time.Time
}
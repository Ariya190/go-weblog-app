package service

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"weblog-app/model"
	"weblog-app/repo"

	"github.com/google/uuid"
)

var (
	ErrPostNotFound = errors.New("post not found")
	ErrForbidden    = errors.New("access forbidden")
	ErrInvalidInput = errors.New("title and content are required")
)

type PostService interface {
	CreatePost(authorID int, title, content string, isPrivate bool, sharedUsers []string, file *multipart.FileHeader) (*model.Post, error)
	GetPostDetails(postID int, userID int) (*model.Post, []model.Comment, error)
	GetFeed(userID int) ([]model.Post, error)
	DeletePost(postID int, authorID int) error
	AddComment(postID int, authorID int, content string) (*model.Comment, error)
}

type postService struct {
	postRepo    repo.PostRepository
	commentRepo repo.CommentRepository
	userRepo    repo.UserRepository
}

func NewPostService(pr repo.PostRepository, cr repo.CommentRepository, ur repo.UserRepository) PostService {
	return &postService{
		postRepo:    pr,
		commentRepo: cr,
		userRepo:    ur,
	}
}

func (s *postService) CreatePost(authorID int, title, content string, isPrivate bool, sharedUsers []string, file *multipart.FileHeader) (*model.Post, error) {
	if strings.TrimSpace(title) == "" || strings.TrimSpace(content) == "" {
		return nil, ErrInvalidInput
	}

	var imageURL string
	if file != nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		filename := uuid.New().String() + ext
		dstPath := filepath.Join("uploads", filename)

		src, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer src.Close()

		dst, err := os.Create(dstPath)
		if err != nil {
			return nil, err
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			return nil, err
		}
		imageURL = "/uploads/" + filename
	}

	var sharedUserIDs []int
	if isPrivate && len(sharedUsers) > 0 {
		users, err := s.userRepo.GetByNames(sharedUsers)
		if err == nil {
			for _, u := range users {
				if u.ID != authorID {
					sharedUserIDs = append(sharedUserIDs, u.ID)
				}
			}
		}
	}

	post := &model.Post{
		AuthorID:  authorID,
		Title:     title,
		Content:   content,
		ImageURL:  imageURL,
		IsPrivate: isPrivate,
	}

	if err := s.postRepo.Create(post, sharedUserIDs); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *postService) GetPostDetails(postID int, userID int) (*model.Post, []model.Comment, error) {
	hasAccess, err := s.postRepo.HasAccess(postID, userID)
	if err != nil {
		return nil, nil, err
	}
	if !hasAccess {
		return nil, nil, ErrForbidden
	}

	post, err := s.postRepo.GetByID(postID)
	if err != nil || post == nil {
		return nil, nil, ErrPostNotFound
	}

	comments, err := s.commentRepo.GetByPostID(postID)
	if err != nil {
		return nil, nil, err
	}

	return post, comments, nil
}

func (s *postService) GetFeed(userID int) ([]model.Post, error) {
	return s.postRepo.GetAccessiblePosts(userID)
}

func (s *postService) DeletePost(postID int, authorID int) error {
	return s.postRepo.Delete(postID, authorID)
}

func (s *postService) AddComment(postID int, authorID int, content string) (*model.Comment, error) {
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("comment cannot be empty")
	}

	hasAccess, err := s.postRepo.HasAccess(postID, authorID)
	if err != nil || !hasAccess {
		return nil, ErrForbidden
	}

	comment := &model.Comment{
		PostID:   postID,
		AuthorID: authorID,
		Content:  content,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}
	return comment, nil
}
package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"weblog-app/repo"
	"weblog-app/service"

	"github.com/labstack/echo/v4"
)

type PostHandler struct {
	postService service.PostService
	userRepo    repo.UserRepository
}

func NewPostHandler(ps service.PostService, ur repo.UserRepository) *PostHandler {
	return &PostHandler{postService: ps, userRepo: ur}
}

func (h *PostHandler) GetFeed(c echo.Context) error {
	userID := c.Get("user_id").(int)
	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	posts, err := h.postService.GetFeed(userID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error loading posts")
	}

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Username": user.Username,
		"UserID":   userID,
		"Posts":    posts,
	})
}

func (h *PostHandler) GetNewPostPage(c echo.Context) error {
	return c.Render(http.StatusOK, "new_post.html", map[string]interface{}{})
}

func (h *PostHandler) PostCreate(c echo.Context) error {
	userID := c.Get("user_id").(int)
	title := c.FormValue("title")
	content := c.FormValue("content")
	isPrivate := c.FormValue("is_private") == "true"
	rawShares := c.FormValue("shared_users")

	var sharedUsers []string
	if rawShares != "" {
		for _, s := range strings.Split(rawShares, ",") {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				sharedUsers = append(sharedUsers, trimmed)
			}
		}
	}

	file, _ := c.FormFile("image")

	_, err := h.postService.CreatePost(userID, title, content, isPrivate, sharedUsers, file)
	if err != nil {
		return c.Render(http.StatusBadRequest, "new_post.html", map[string]interface{}{
			"Error": err.Error(),
		})
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *PostHandler) GetPostDetail(c echo.Context) error {
	userID := c.Get("user_id").(int)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID")
	}

	post, comments, err := h.postService.GetPostDetails(id, userID)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			return c.String(http.StatusNotFound, "Post Not Found")
		}
		if errors.Is(err, service.ErrForbidden) {
			return c.String(http.StatusForbidden, "Forbidden")
		}
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}

	return c.Render(http.StatusOK, "post_detail.html", map[string]interface{}{
		"Post":     post,
		"Comments": comments,
		"UserID":   userID,
	})
}

func (h *PostHandler) PostDelete(c echo.Context) error {
	userID := c.Get("user_id").(int)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID")
	}

	err = h.postService.DeletePost(id, userID)
	if err != nil {
		return c.String(http.StatusForbidden, "Cannot delete post")
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *PostHandler) PostComment(c echo.Context) error {
	userID := c.Get("user_id").(int)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID")
	}

	content := c.FormValue("content")
	_, err = h.postService.AddComment(id, userID, content)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/weblog/"+strconv.Itoa(id))
	}

	return c.Redirect(http.StatusSeeOther, "/weblog/"+strconv.Itoa(id))
}
package handlers

import (
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
)

const maxCommentLen = 2000

// CreateComment adds a comment to a published post.
func (h *Handler) CreateComment(c fiber.Ctx) error {
	user := CurrentUser(c)
	post, err := h.Store.PostBySlug(c, c.Params("slug"))
	if err != nil {
		return orNotFound(err)
	}
	if !post.IsPublished() && !canEdit(user, post) {
		return notFound()
	}
	back := "/posts/" + url.PathEscape(post.Slug)

	body := strings.TrimSpace(strings.ReplaceAll(c.FormValue("body"), "\r\n", "\n"))
	switch n := utf8.RuneCountInString(body); {
	case n == 0:
		return c.Redirect().With("error", "Your comment is empty.").To(back + "#comments")
	case n > maxCommentLen:
		return c.Redirect().With("error", "Comments are limited to 2,000 characters.").To(back + "#comments")
	}

	comment, err := h.Store.CreateComment(c, post.ID, user.ID, body)
	if err != nil {
		return err
	}
	return c.Redirect().With("success", "Comment posted.").To(back + "#comment-" + strconv.FormatInt(comment.ID, 10))
}

// DeleteComment removes a comment. Allowed for the comment's author, the
// post's author, and admins.
func (h *Handler) DeleteComment(c fiber.Ctx) error {
	user := CurrentUser(c)
	id := fiber.Params[int64](c, "id")
	comment, err := h.Store.CommentByID(c, id)
	if err != nil {
		return orNotFound(err)
	}
	if user.ID != comment.AuthorID && !canEdit(user, comment.Post) {
		return fiber.ErrForbidden
	}
	if err := h.Store.DeleteComment(c, comment.ID); err != nil {
		return err
	}
	fallback := "/posts/" + url.PathEscape(comment.Post.Slug) + "#comments"
	return c.Redirect().With("success", "Comment deleted.").To(formNext(c, fallback))
}

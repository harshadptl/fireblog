package handlers

import (
	"context"
	"net/url"

	"github.com/gofiber/fiber/v3"
)

type toggleFunc func(ctx context.Context, userID, postID int64) (bool, int, error)

// ToggleLike likes/unlikes a post.
func (h *Handler) ToggleLike(c fiber.Ctx) error {
	return h.toggleReaction(c, h.Store.ToggleLike, "Liked!", "Like removed.")
}

// ToggleBookmark saves/unsaves a post to the reading list.
func (h *Handler) ToggleBookmark(c fiber.Ctx) error {
	return h.toggleReaction(c, h.Store.ToggleBookmark, "Saved to your reading list.", "Removed from your reading list.")
}

// toggleReaction answers fetch() calls with JSON and plain form posts with a
// redirect, so the buttons work with and without JavaScript.
func (h *Handler) toggleReaction(c fiber.Ctx, toggle toggleFunc, onMsg, offMsg string) error {
	user := CurrentUser(c)
	post, err := h.Store.PostBySlug(c, c.Params("slug"))
	if err != nil {
		return orNotFound(err)
	}
	if !post.IsPublished() {
		return notFound()
	}
	active, count, err := toggle(c, user.ID, post.ID)
	if err != nil {
		return err
	}
	if isFetch(c) {
		return c.JSON(fiber.Map{"active": active, "count": count})
	}
	msg := offMsg
	if active {
		msg = onMsg
	}
	return c.Redirect().With("success", msg).To(formNext(c, "/posts/"+url.PathEscape(post.Slug)))
}

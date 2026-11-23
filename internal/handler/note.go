package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"redbook-go/internal/service"
)

func (h *Handler) SaveNote(c *gin.Context) {
	var req service.NoteSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	data, err := h.Service.SaveNote(h.userID(c), req)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) PublishNote(c *gin.Context) {
	id, _ := parseInt64(c.Param("id"))
	data, err := h.Service.PublishNote(h.userID(c), id)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) DeleteNote(c *gin.Context) {
	id, _ := parseInt64(c.Param("id"))
	if err := h.Service.DeleteNote(h.userID(c), id); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) NoteDetail(c *gin.Context) {
	id, _ := parseInt64(c.Param("id"))
	data, err := h.Service.NoteDetail(id, h.viewerID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) NoteMy(c *gin.Context) {
	current, size := pagination(c)
	list, total, err := h.Service.MyNotes(h.userID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) NoteFeed(c *gin.Context) {
	var maxTime *int64
	if v := c.Query("maxTime"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			maxTime = &n
		}
	}
	var offset *int
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = &n
		}
	}
	size, _ := strconv.ParseInt(c.DefaultQuery("size", "10"), 10, 64)
	data, err := h.Service.Feed(h.userID(c), maxTime, offset, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) NoteRecommend(c *gin.Context) {
	current, size := pagination(c)
	list, total, err := h.Service.Recommend(current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) NoteRelated(c *gin.Context) {
	id, _ := parseInt64(c.Param("id"))
	list, err := h.Service.Related(id)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, list)
}

func (h *Handler) NoteUser(c *gin.Context) {
	target, _ := parseInt64(c.Param("userId"))
	current, size := pagination(c)
	list, total, err := h.Service.UserNotes(&target, h.viewerID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) CommentAdd(c *gin.Context) {
	noteID, _ := parseInt64(c.Param("noteId"))
	var req service.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	data, err := h.Service.AddComment(noteID, h.userID(c), req)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) CommentTree(c *gin.Context) {
	noteID, _ := parseInt64(c.Param("noteId"))
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	data, err := h.Service.CommentTree(noteID)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) CommentDelete(c *gin.Context) {
	commentID, _ := parseInt64(c.Param("commentId"))
	if err := h.Service.DeleteComment(h.userID(c), commentID); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

package handler

import "github.com/gin-gonic/gin"

func (h *Handler) LikeView(c *gin.Context) {
	id, _ := parseInt64(c.Param("receiverId"))
	data, err := h.Service.LikeView(h.userID(c), id)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) Like(c *gin.Context) {
	id, _ := parseInt64(c.Param("receiverId"))
	data, err := h.Service.Like(h.userID(c), id)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) Unlike(c *gin.Context) {
	id, _ := parseInt64(c.Param("receiverId"))
	data, err := h.Service.Unlike(h.userID(c), id)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) LikeMy(c *gin.Context) {
	current, size := pagination(c)
	list, total, err := h.Service.LikeMy(h.userID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) LikeByUser(c *gin.Context) {
	id, _ := parseInt64(c.Param("userId"))
	current, size := pagination(c)
	list, total, err := h.Service.LikeByUser(id, h.viewerID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) Collect(c *gin.Context) {
	id, _ := parseInt64(c.Param("noteId"))
	if err := h.Service.Collect(h.userID(c), id); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) Uncollect(c *gin.Context) {
	id, _ := parseInt64(c.Param("noteId"))
	if err := h.Service.Uncollect(h.userID(c), id); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) CollectMy(c *gin.Context) {
	current, size := pagination(c)
	list, total, err := h.Service.CollectMy(h.userID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) CollectByUser(c *gin.Context) {
	id, _ := parseInt64(c.Param("userId"))
	current, size := pagination(c)
	list, total, err := h.Service.CollectByUser(id, h.viewerID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) Follow(c *gin.Context) {
	id, _ := parseInt64(c.Param("userId"))
	if err := h.Service.Follow(h.userID(c), id); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) Unfollow(c *gin.Context) {
	id, _ := parseInt64(c.Param("userId"))
	if err := h.Service.Unfollow(h.userID(c), id); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) Followers(c *gin.Context) {
	current, size := pagination(c)
	list, total, err := h.Service.Followers(h.userID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) Following(c *gin.Context) {
	current, size := pagination(c)
	list, total, err := h.Service.Following(h.userID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) FollowersOfUser(c *gin.Context) {
	id, _ := parseInt64(c.Param("userId"))
	current, size := pagination(c)
	list, total, err := h.Service.Followers(id, current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) FollowingOfUser(c *gin.Context) {
	id, _ := parseInt64(c.Param("userId"))
	current, size := pagination(c)
	list, total, err := h.Service.Following(id, current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

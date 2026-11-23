package handler

import (
	"github.com/gin-gonic/gin"

	"redbook-go/internal/service"
)

func (h *Handler) ChatUnreadTotal(c *gin.Context) {
	data, err := h.Service.ChatUnreadTotal(h.userID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) ChatThreads(c *gin.Context) {
	current, size := pagination(c)
	list, total, err := h.Service.ChatThreads(h.userID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) ChatMarkThreadRead(c *gin.Context) {
	id, _ := parseInt64(c.Param("threadId"))
	if err := h.Service.ChatMarkThreadRead(h.userID(c), id); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) ChatPeerThread(c *gin.Context) {
	peerID, _ := parseInt64(c.Query("peerUserId"))
	id, err := h.Service.ChatEnsureThread(h.userID(c), peerID)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, id)
}

func (h *Handler) ChatMessages(c *gin.Context) {
	id, _ := parseInt64(c.Param("threadId"))
	current, size := pagination(c)
	list, total, err := h.Service.ChatMessages(h.userID(c), id, current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) ChatSend(c *gin.Context) {
	toUserID, _ := parseInt64(c.Query("toUserId"))
	content := c.Query("content")
	msgID, err := h.Service.ChatSend(h.userID(c), toUserID, content)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, msgID)
}

func (h *Handler) GroupCreate(c *gin.Context) {
	var req service.GroupCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	group, err := h.Service.GroupCreate(h.userID(c), req)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, group.ID)
}

func (h *Handler) GroupByOwner(c *gin.Context) {
	id, _ := parseInt64(c.Param("ownerUserId"))
	groups, err := h.Service.GroupByOwner(id, h.viewerID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, groups)
}

func (h *Handler) GroupMy(c *gin.Context) {
	groups, err := h.Service.GroupMy(h.userID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, groups)
}

func (h *Handler) GroupJoin(c *gin.Context) {
	id, _ := parseInt64(c.Param("groupId"))
	state, err := h.Service.GroupJoin(id, h.userID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, gin.H{"status": state})
}

func (h *Handler) GroupHandleJoinRequest(c *gin.Context) {
	id, _ := parseInt64(c.Param("requestId"))
	approve := c.Query("approve") == "true"
	if err := h.Service.GroupHandleJoinRequest(h.userID(c), id, approve); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) GroupMeta(c *gin.Context) {
	id, _ := parseInt64(c.Param("groupId"))
	data, err := h.Service.GroupMeta(id, h.userID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) GroupMessages(c *gin.Context) {
	id, _ := parseInt64(c.Param("groupId"))
	current, size := pagination(c)
	list, total, err := h.Service.GroupMessages(id, h.userID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) GroupSend(c *gin.Context) {
	id, _ := parseInt64(c.Param("groupId"))
	content := c.Query("content")
	msgID, err := h.Service.GroupSend(id, h.userID(c), content)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, msgID)
}

func (h *Handler) GroupMarkRead(c *gin.Context) {
	id, _ := parseInt64(c.Param("groupId"))
	if err := h.Service.GroupMarkRead(id, h.userID(c)); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) GroupDetail(c *gin.Context) {
	id, _ := parseInt64(c.Param("groupId"))
	data, err := h.Service.GroupDetail(id, h.userID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) GroupKick(c *gin.Context) {
	id, _ := parseInt64(c.Param("groupId"))
	target, _ := parseInt64(c.Query("targetUserId"))
	if err := h.Service.GroupKick(h.userID(c), id, target); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) GroupUpdateProfile(c *gin.Context) {
	id, _ := parseInt64(c.Param("groupId"))
	var req service.GroupUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if err := h.Service.GroupUpdateProfile(h.userID(c), id, req); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) NoticeInteractions(c *gin.Context) {
	current, size := pagination(c)
	category := c.DefaultQuery("category", "all")
	list, total, err := h.Service.NoticeInteractionPage(h.userID(c), current, size, category)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) NoticeRead(c *gin.Context) {
	category := c.Query("category")
	if err := h.Service.NoticeMarkRead(h.userID(c), category); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

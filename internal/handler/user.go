package handler

import (
	"time"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Phone    string `json:"phone"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

type profileRequest struct {
	NickName string `json:"nickName"`
	Icon     string `json:"icon"`
}

type userInfoRequest struct {
	City          *string `json:"city"`
	Introduce     *string `json:"introduce"`
	Gender        *bool   `json:"gender"`
	Birthday      *string `json:"birthday"`
	CollectPublic *bool   `json:"collectPublic"`
	LikePublic    *bool   `json:"likePublic"`
}

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func (h *Handler) SendCode(c *gin.Context) {
	phone := c.Query("phone")
	code, err := h.Service.SendCode(phone)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, code)
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	token, err := h.Service.LoginByCode(req.Phone, req.Code)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, token)
}

func (h *Handler) LoginPassword(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	token, err := h.Service.LoginByPassword(req.Phone, req.Password)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, token)
}

func (h *Handler) Logout(c *gin.Context) {
	if err := h.Service.Logout(c.GetHeader("authorization")); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) Sign(c *gin.Context) {
	data, err := h.Service.Sign(h.userID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) SignCount(c *gin.Context) {
	count, err := h.Service.SignCount(h.userID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, count)
}

func (h *Handler) Me(c *gin.Context) {
	data, err := h.Service.Me(h.userID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) PublicProfile(c *gin.Context) {
	id, _ := parseInt64(c.Param("userId"))
	data, err := h.Service.PublicProfile(id, h.viewerID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var req profileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	data, err := h.Service.UpdateProfile(h.userID(c), req.NickName, req.Icon)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) MyInfo(c *gin.Context) {
	data, err := h.Service.MyInfo(h.userID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) UpdateInfo(c *gin.Context) {
	var req userInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	updates := map[string]any{}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.Introduce != nil {
		updates["introduce"] = *req.Introduce
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.CollectPublic != nil {
		updates["collect_public"] = *req.CollectPublic
	}
	if req.LikePublic != nil {
		updates["like_public"] = *req.LikePublic
	}
	if req.Birthday != nil {
		if t, err := time.Parse("2006-01-02", *req.Birthday); err == nil {
			updates["birthday"] = t
		}
	}
	data, err := h.Service.UpdateInfo(h.userID(c), updates)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if err := h.Service.ChangePassword(h.userID(c), req.OldPassword, req.NewPassword); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

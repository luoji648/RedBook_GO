package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"redbook-go/internal/middleware"
	"redbook-go/internal/response"
	"redbook-go/internal/service"
)

type Handler struct {
	Service *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{Service: svc}
}

func (h *Handler) userID(c *gin.Context) int64 {
	id, _ := middleware.CurrentUserID(c)
	return id
}

func (h *Handler) viewerID(c *gin.Context) *int64 {
	id, ok := middleware.CurrentUserID(c)
	if !ok {
		return nil
	}
	return &id
}

func pagination(c *gin.Context) (int64, int64) {
	current, _ := strconv.ParseInt(c.DefaultQuery("current", "1"), 10, 64)
	size, _ := strconv.ParseInt(c.DefaultQuery("size", "20"), 10, 64)
	if current < 1 {
		current = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	return current, size
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func ok(c *gin.Context, data any) {
	response.OKJSON(c, data)
}

func okList(c *gin.Context, data any, total int64) {
	response.OKList(c, data, total)
}

func fail(c *gin.Context, status int, msg string) {
	response.FailJSON(c, status, msg)
}

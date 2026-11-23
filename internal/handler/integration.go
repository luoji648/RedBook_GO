package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ShareNote(c *gin.Context) {
	id, _ := parseInt64(c.Param("noteId"))
	data, err := h.Service.GenerateShortLink(id)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) GenerateShortLink(c *gin.Context) {
	id, _ := parseInt64(c.Query("noteId"))
	data, err := h.Service.GenerateShortLink(id)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) ShortLinkRedirect(c *gin.Context) {
	code := c.Param("code")
	longURL, err := h.Service.ResolveShortLink(code)
	if err != nil {
		c.String(http.StatusNotFound, "short link not found")
		return
	}
	c.Redirect(http.StatusFound, longURL)
}

func (h *Handler) OssPresign(c *gin.Context) {
	data, err := h.Service.OSS.PresignedPut(c.Query("ext"), c.Query("contentType"))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) OssUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		fail(c, 400, "请选择文件")
		return
	}
	defer file.Close()
	data, err := h.Service.OSS.UploadViaAPI(file, header)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) AiStream(c *gin.Context) {
	id, _ := parseInt64(c.Param("noteId"))
	mode := c.Query("mode")
	question := c.Query("question")
	resp, err := h.Service.DifyStream(c.Request.Context(), h.userID(c), id, mode, question)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	defer resp.Body.Close()
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
	c.Writer.Flush()
}

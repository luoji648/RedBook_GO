package response

import "github.com/gin-gonic/gin"

type Result struct {
	Success  bool   `json:"success"`
	ErrorMsg *string `json:"errorMsg"`
	Data     any    `json:"data"`
	Total    *int64 `json:"total"`
}

func OK(data any) Result {
	return Result{Success: true, Data: data}
}

func OKWithTotal(data any, total int64) Result {
	return Result{Success: true, Data: data, Total: &total}
}

func Fail(msg string) Result {
	return Result{Success: false, ErrorMsg: &msg}
}

func JSON(c *gin.Context, status int, result Result) {
	c.JSON(status, result)
}

func OKJSON(c *gin.Context, data any) {
	c.JSON(200, OK(data))
}

func OKList(c *gin.Context, data any, total int64) {
	c.JSON(200, OKWithTotal(data, total))
}

func FailJSON(c *gin.Context, status int, msg string) {
	c.JSON(status, Fail(msg))
}

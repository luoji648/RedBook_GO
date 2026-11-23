package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"redbook-go/internal/response"
)

func TestResponseEnvelopeShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ok", func(c *gin.Context) {
		response.OKList(c, []string{"a"}, 1)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["success"] != true {
		t.Fatalf("unexpected body: %v", body)
	}
	if body["data"] == nil {
		t.Fatalf("missing data: %v", body)
	}
	if body["total"] != float64(1) {
		t.Fatalf("missing total: %v", body)
	}
}

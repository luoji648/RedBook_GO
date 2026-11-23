package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func SwaggerDoc(routes gin.RoutesInfo) gin.HandlerFunc {
	return func(c *gin.Context) {
		paths := map[string]any{}
		for _, route := range routes {
			path := ginPathToOpenAPI(route.Path)
			method := strings.ToLower(route.Method)
			if paths[path] == nil {
				paths[path] = map[string]any{}
			}
			pathMap := paths[path].(map[string]any)
			pathMap[method] = map[string]any{
				"operationId": strings.ToLower(route.Method) + "_" + sanitizeID(path),
				"summary":     route.Handler,
			}
		}
		c.JSON(200, gin.H{
			"openapi": "3.0.0",
			"info": gin.H{
				"title":       "Redbook Go Backend",
				"version":     "1.0.0",
				"description": "API compatibility implementation for the existing Vue frontend.",
			},
			"paths": paths,
		})
	}
}

func SwaggerHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Redbook Go API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({ url: '/swagger/doc.json', dom_id: '#swagger-ui' });
  </script>
</body>
</html>`)
	}
}

func ginPathToOpenAPI(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		} else if strings.HasPrefix(part, "*") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

func sanitizeID(path string) string {
	replacer := strings.NewReplacer("/", "_", "{", "", "}", "", "-", "_")
	return replacer.Replace(path)
}

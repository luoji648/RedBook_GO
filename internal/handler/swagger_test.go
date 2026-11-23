package handler

import "testing"

func TestGinPathToOpenAPI(t *testing.T) {
	got := ginPathToOpenAPI("/note/:id")
	if got != "/note/{id}" {
		t.Fatalf("got %s", got)
	}
}

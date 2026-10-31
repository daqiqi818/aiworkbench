package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"infinite-canvas/backend/internal/dramaedit"

	"github.com/gin-gonic/gin"
)

func TestFailDramaEditingProjectsPublicErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, item := range []struct {
		name string
		err  error
		want int
	}{
		{name: "not found", err: dramaedit.ErrNotFound, want: http.StatusNotFound},
		{name: "validation", err: &dramaedit.ValidationError{Message: "无效请求"}, want: http.StatusBadRequest},
		{name: "transition", err: dramaedit.ErrInvalidTransition, want: http.StatusBadRequest},
	} {
		t.Run(item.name, func(t *testing.T) {
			writer := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(writer)
			failDramaEditing(context, item.err)
			if writer.Code != item.want {
				t.Fatalf("status = %d, want %d", writer.Code, item.want)
			}
		})
	}
}

func TestRegisterDramaEditingRoutesIncludesEpisodeFolders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterDramaEditingRoutes(router.Group("/api"), nil)
	for _, route := range router.Routes() {
		if route.Method == http.MethodGet && route.Path == "/api/drama-editing/episode-folders" {
			return
		}
	}
	t.Fatal("episode folder route was not registered")
}

func TestRegisterDramaEditingRoutesIncludesDeleteTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterDramaEditingRoutes(router.Group("/api"), nil)
	for _, route := range router.Routes() {
		if route.Method == http.MethodDelete && route.Path == "/api/drama-editing/tasks" {
			return
		}
	}
	t.Fatal("task deletion route was not registered")
}

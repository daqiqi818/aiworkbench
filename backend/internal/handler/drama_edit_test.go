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

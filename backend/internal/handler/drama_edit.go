package handler

import (
	"errors"
	"net/http"
	"strings"

	"infinite-canvas/backend/internal/app"
	"infinite-canvas/backend/internal/dramaedit"

	"github.com/gin-gonic/gin"
)

func RegisterDramaEditingRoutes(r *gin.RouterGroup, svc *app.Service) {
	r.GET("/drama-editing/tasks", func(c *gin.Context) {
		user, err := currentUser(c, svc)
		if err != nil {
			failService(c, err)
			return
		}
		tasks, err := requestDramaEditingPort(c, svc).List(user.ID, strings.TrimSpace(c.Query("status")))
		if err != nil {
			failDramaEditing(c, err)
			return
		}
		ok(c, gin.H{"tasks": tasks})
	})
	r.POST("/drama-editing/tasks", func(c *gin.Context) {
		if !requireTrustedDesktopWritePrincipal(c, "剪辑任务只能由当前桌面界面创建") {
			return
		}
		user, err := currentUser(c, svc)
		if err != nil {
			failService(c, err)
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256<<10)
		var request dramaedit.CreateRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			fail(c, http.StatusBadRequest, err)
			return
		}
		task, err := requestDramaEditingPort(c, svc).Create(user.ID, request)
		if err != nil {
			failDramaEditing(c, err)
			return
		}
		ok(c, task)
	})
	r.GET("/drama-editing/tasks/:id", func(c *gin.Context) {
		user, err := currentUser(c, svc)
		if err != nil {
			failService(c, err)
			return
		}
		task, err := requestDramaEditingPort(c, svc).Get(user.ID, c.Param("id"))
		if err != nil {
			failDramaEditing(c, err)
			return
		}
		ok(c, task)
	})
	r.POST("/drama-editing/tasks/:id/approve", func(c *gin.Context) {
		if !requireTrustedDesktopWritePrincipal(c, "剪辑方案只能由当前桌面界面批准") {
			return
		}
		user, err := currentUser(c, svc)
		if err != nil {
			failService(c, err)
			return
		}
		task, err := requestDramaEditingPort(c, svc).Approve(user.ID, c.Param("id"))
		if err != nil {
			failDramaEditing(c, err)
			return
		}
		ok(c, task)
	})
	r.POST("/drama-editing/tasks/:id/revision", func(c *gin.Context) {
		if !requireTrustedDesktopWritePrincipal(c, "剪辑任务只能由当前桌面界面修改") {
			return
		}
		user, err := currentUser(c, svc)
		if err != nil {
			failService(c, err)
			return
		}
		var request struct {
			Note string `json:"note"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			fail(c, http.StatusBadRequest, err)
			return
		}
		task, err := requestDramaEditingPort(c, svc).RequestRevision(user.ID, c.Param("id"), request.Note)
		if err != nil {
			failDramaEditing(c, err)
			return
		}
		ok(c, task)
	})
	r.POST("/drama-editing/tasks/:id/cancel", func(c *gin.Context) {
		if !requireTrustedDesktopWritePrincipal(c, "剪辑任务只能由当前桌面界面修改") {
			return
		}
		user, err := currentUser(c, svc)
		if err != nil {
			failService(c, err)
			return
		}
		var request struct {
			Note string `json:"note"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			fail(c, http.StatusBadRequest, err)
			return
		}
		task, err := requestDramaEditingPort(c, svc).Cancel(user.ID, c.Param("id"), request.Note)
		if err != nil {
			failDramaEditing(c, err)
			return
		}
		ok(c, task)
	})
}

func failDramaEditing(c *gin.Context, err error) {
	var validation *dramaedit.ValidationError
	switch {
	case errors.Is(err, dramaedit.ErrNotFound):
		fail(c, http.StatusNotFound, err)
	case errors.As(err, &validation), errors.Is(err, dramaedit.ErrInvalidTransition):
		fail(c, http.StatusBadRequest, err)
	default:
		failService(c, err)
	}
}

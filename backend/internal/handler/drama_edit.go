package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"infinite-canvas/backend/internal/app"
	"infinite-canvas/backend/internal/dramaedit"

	"github.com/gin-gonic/gin"
)

func RegisterDramaEditingRoutes(r *gin.RouterGroup, svc *app.Service) {
	r.GET("/internal/openmontage/productions/:productionID/sources/:resourceID", func(c *gin.Context) {
		productionID, resourceID := c.Param("productionID"), c.Param("resourceID")
		// The opaque token is generated per BeefTV task and never reaches the browser.
		port := requestDramaEditingPort(c, svc)
		userID, err := port.ValidateOpenMontageSourceByProduction(productionID, resourceID, c.GetHeader("X-BeefTV-Production-Token"))
		if err != nil {
			failDramaEditing(c, err)
			return
		}
		_, body, err := svc.OpenResource(userID, resourceID)
		if err != nil {
			failService(c, err)
			return
		}
		defer body.Close()
		c.Header("Content-Type", "video/mp4")
		_, _ = io.Copy(c.Writer, body)
	})
	r.GET("/drama-editing/episode-folders", func(c *gin.Context) {
		user, err := currentUser(c, svc)
		if err != nil {
			failService(c, err)
			return
		}
		folders, err := requestDramaEditingPort(c, svc).EpisodeFolders(user.ID)
		if err != nil {
			failDramaEditing(c, err)
			return
		}
		ok(c, gin.H{"folders": folders})
	})
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
	r.DELETE("/drama-editing/tasks", func(c *gin.Context) {
		if !requireTrustedDesktopWritePrincipal(c, "剪辑任务只能由当前桌面界面删除") {
			return
		}
		user, err := currentUser(c, svc)
		if err != nil {
			failService(c, err)
			return
		}
		var request dramaedit.DeleteRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			fail(c, http.StatusBadRequest, err)
			return
		}
		result, err := requestDramaEditingPort(c, svc).Delete(user.ID, request)
		if err != nil {
			failDramaEditing(c, err)
			return
		}
		ok(c, gin.H{"results": result})
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
	r.POST("/drama-editing/tasks/:id/start", func(c *gin.Context) {
		if !requireTrustedDesktopWritePrincipal(c, "剪辑任务只能由当前桌面界面开始制作") {
			return
		}
		user, err := currentUser(c, svc)
		if err != nil {
			failService(c, err)
			return
		}
		task, err := requestDramaEditingPort(c, svc).StartProduction(user.ID, c.Param("id"))
		if err != nil {
			failDramaEditing(c, err)
			return
		}
		ok(c, task)
	})
	r.POST("/drama-editing/tasks/:id/openmontage/stages/:stage/approve", func(c *gin.Context) {
		if !requireTrustedDesktopWritePrincipal(c, "OpenMontage 阶段只能由当前桌面界面批准") {
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
		task, err := requestDramaEditingPort(c, svc).ApproveNativeStage(user.ID, c.Param("id"), c.Param("stage"), request.Note)
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

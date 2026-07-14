package category

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mininglamp-OSS/octo-marketplace/internal/api/errcode"
	"github.com/Mininglamp-OSS/octo-marketplace/internal/middleware"
	categorysvc "github.com/Mininglamp-OSS/octo-marketplace/internal/service/category"
	"github.com/gin-gonic/gin"
)

// RegisterAdmin registers admin category routes on the given router group.
func (h *Handler) RegisterAdmin(rg *gin.RouterGroup, idGen func() string) {
	rg.POST("/skill/admin/categories", h.adminCreate(idGen))
	rg.PUT("/skill/admin/categories/:id", h.adminUpdate())
	rg.DELETE("/skill/admin/categories/:id", h.adminDelete())
}

type adminCreateRequest struct {
	Name      string `json:"name" binding:"required"`
	IconKey   string `json:"icon_key"`
	SortOrder int    `json:"sort_order"`
}

func (h *Handler) adminCreate(idGen func() string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := middleware.Identity(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": errcode.Unauthorized, "message": "unauthorized"})
			return
		}

		var req adminCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": errcode.BadRequest, "message": "name is required"})
			return
		}

		id := idGen()
		item, err := h.svc.Create(c.Request.Context(), id, req.Name, req.IconKey, req.SortOrder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": errcode.InternalError, "message": "internal error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"code": 0, "data": item})
	}
}

type adminUpdateRequest struct {
	Name      string `json:"name" binding:"required"`
	IconKey   string `json:"icon_key"`
	SortOrder int    `json:"sort_order"`
}

func (h *Handler) adminUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := middleware.Identity(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": errcode.Unauthorized, "message": "unauthorized"})
			return
		}

		id := c.Param("id")
		var req adminUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": errcode.BadRequest, "message": "name is required"})
			return
		}

		item, err := h.svc.Update(c.Request.Context(), id, req.Name, req.IconKey, req.SortOrder)
		if err != nil {
			if errors.Is(err, categorysvc.ErrCategoryNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"code": errcode.NotFound, "message": "category not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"code": errcode.InternalError, "message": "internal error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
	}
}

func (h *Handler) adminDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := middleware.Identity(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": errcode.Unauthorized, "message": "unauthorized"})
			return
		}

		id := c.Param("id")
		count, err := h.svc.Delete(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, categorysvc.ErrCategoryInUse) {
				c.JSON(http.StatusConflict, gin.H{
					"code":    errcode.CategoryInUse,
					"message": fmt.Sprintf("该分类下存在 %d 个 Skill，请先迁移", count),
				})
				return
			}
			if errors.Is(err, categorysvc.ErrCategoryNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"code": errcode.NotFound, "message": "category not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"code": errcode.InternalError, "message": "internal error"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

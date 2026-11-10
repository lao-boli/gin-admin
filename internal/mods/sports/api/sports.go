package api

import (
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/biz"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/gin-gonic/gin"
)

// Category 分类树管理
type Category struct {
	CategoryBIZ *biz.Category
}

// @Tags SportsCategoryAPI
// @Security ApiKeyAuth
// @Summary Query category tree
// @Router /api/v1/sports/categories/tree [get]
func (a *Category) QueryTree(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.CategoryQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.CategoryBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	if result == nil {
		result = make(schema.Categories, 0)
	}
	util.ResSuccess(c, result)
}

// @Tags SportsCategoryAPI
// @Security ApiKeyAuth
// @Summary Create category
// @Router /api/v1/sports/categories [post]
func (a *Category) Create(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.CategoryForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.CategoryBIZ.Create(ctx, item)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags SportsCategoryAPI
// @Security ApiKeyAuth
// @Summary Update category
// @Router /api/v1/sports/categories/{id} [put]
func (a *Category) Update(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.CategoryForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}
	if err := a.CategoryBIZ.Update(ctx, c.Param("id"), item); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags SportsCategoryAPI
// @Security ApiKeyAuth
// @Summary Delete category
// @Router /api/v1/sports/categories/{id} [delete]
func (a *Category) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	if err := a.CategoryBIZ.Delete(ctx, c.Param("id")); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// ---------------- Item ----------------

// Item 测试项目管理
type Item struct {
	ItemBIZ    *biz.Item
	SubjectBIZ *biz.Subject
}

// @Tags SportsItemAPI
// @Security ApiKeyAuth
// @Summary Query item list
// @Router /api/v1/sports/items [get]
func (a *Item) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.ItemQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.ItemBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	if result == nil {
		result = make(schema.Items, 0)
	}
	util.ResSuccess(c, result)
}

// @Tags SportsItemAPI
// @Security ApiKeyAuth
// @Summary Get item (with fields/metrics/rule)
// @Router /api/v1/sports/items/{id} [get]
func (a *Item) Get(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := a.ItemBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags SportsItemAPI
// @Security ApiKeyAuth
// @Summary Create item
// @Router /api/v1/sports/items [post]
func (a *Item) Create(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.ItemForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.ItemBIZ.Create(ctx, item)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags SportsItemAPI
// @Security ApiKeyAuth
// @Summary Update item
// @Router /api/v1/sports/items/{id} [put]
func (a *Item) Update(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.ItemForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}
	if err := a.ItemBIZ.Update(ctx, c.Param("id"), item); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags SportsItemAPI
// @Security ApiKeyAuth
// @Summary Delete item
// @Router /api/v1/sports/items/{id} [delete]
func (a *Item) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	if err := a.ItemBIZ.Delete(ctx, c.Param("id")); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags SportsItemAPI
// @Security ApiKeyAuth
// @Summary Copy item to target category
// @Router /api/v1/sports/items/{id}/copy [post]
func (a *Item) Copy(c *gin.Context) {
	ctx := c.Request.Context()
	var params struct {
		TargetCategoryID string `json:"target_category_id" binding:"required"`
	}
	if err := util.ParseJSON(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.ItemBIZ.Copy(ctx, c.Param("id"), params.TargetCategoryID)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags SportsItemAPI
// @Security ApiKeyAuth
// @Summary Validate metric formula
// @Router /api/v1/sports/items/metrics/validate [post]
func (a *Item) ValidateMetric(c *gin.Context) {
	ctx := c.Request.Context()
	var params struct {
		Formula string `json:"formula" binding:"required"`
	}
	if err := util.ParseJSON(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	if err := a.ItemBIZ.ValidateMetric(ctx, params.Formula); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags SportsItemAPI
// @Security ApiKeyAuth
// @Summary Eval fields preview (fields → metrics → score)
// @Router /api/v1/sports/items/{id}/eval [post]
func (a *Item) Eval(c *gin.Context) {
	ctx := c.Request.Context()
	var params struct {
		SubjectID string                 `json:"subject_id" binding:"required"`
		Fields    map[string]interface{} `json:"fields"`
	}
	if err := util.ParseJSON(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	subject, err := a.SubjectBIZ.Get(ctx, params.SubjectID)
	if err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.ItemBIZ.Eval(ctx, c.Param("id"), subject, params.Fields)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// ---------------- Standard ----------------

// Standard 评分标准管理
type Standard struct {
	StandardBIZ *biz.Standard
}

// @Tags SportsStandardAPI
// @Security ApiKeyAuth
// @Summary Query standards by item
// @Router /api/v1/sports/standards [get]
func (a *Standard) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.StandardQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.StandardBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	if result == nil {
		result = make(schema.Standards, 0)
	}
	util.ResSuccess(c, result)
}

// @Tags SportsStandardAPI
// @Security ApiKeyAuth
// @Summary Create standard
// @Router /api/v1/sports/standards [post]
func (a *Standard) Create(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.StandardForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.StandardBIZ.Create(ctx, item)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags SportsStandardAPI
// @Security ApiKeyAuth
// @Summary Batch create standards (replace item's standards)
// @Router /api/v1/sports/standards/batch [post]
func (a *Standard) BatchCreate(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.StandardBatchForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}
	err := a.StandardBIZ.BatchCreate(ctx, item)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags SportsStandardAPI
// @Security ApiKeyAuth
// @Summary Delete standard
// @Router /api/v1/sports/standards/{id} [delete]
func (a *Standard) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	if err := a.StandardBIZ.Delete(ctx, c.Param("id")); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

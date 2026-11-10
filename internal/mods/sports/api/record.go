package api

import (
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/biz"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/gin-gonic/gin"
)

// Subject 受试者管理
type Subject struct {
	SubjectBIZ *biz.Subject
}

// @Tags SportsSubjectAPI
// @Security ApiKeyAuth
// @Summary Query subject list
// @Router /api/v1/sports/subjects [get]
func (a *Subject) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.SubjectQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.SubjectBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags SportsSubjectAPI
// @Security ApiKeyAuth
// @Summary Get subject
// @Router /api/v1/sports/subjects/{id} [get]
func (a *Subject) Get(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := a.SubjectBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags SportsSubjectAPI
// @Security ApiKeyAuth
// @Summary Create subject
// @Router /api/v1/sports/subjects [post]
func (a *Subject) Create(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.SubjectForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.SubjectBIZ.Create(ctx, item)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags SportsSubjectAPI
// @Security ApiKeyAuth
// @Summary Update subject
// @Router /api/v1/sports/subjects/{id} [put]
func (a *Subject) Update(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.SubjectForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}
	if err := a.SubjectBIZ.Update(ctx, c.Param("id"), item); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags SportsSubjectAPI
// @Security ApiKeyAuth
// @Summary Delete subject
// @Router /api/v1/sports/subjects/{id} [delete]
func (a *Subject) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	if err := a.SubjectBIZ.Delete(ctx, c.Param("id")); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// ---------------- Record ----------------

// Record 测试记录
type Record struct {
	RecordBIZ *biz.Record
}

// @Tags SportsRecordAPI
// @Security ApiKeyAuth
// @Summary Query records
// @Router /api/v1/sports/records [get]
func (a *Record) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.RecordQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.RecordBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags SportsRecordAPI
// @Security ApiKeyAuth
// @Summary Create record (single)
// @Router /api/v1/sports/records [post]
func (a *Record) Create(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.RecordForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.RecordBIZ.Create(ctx, item, util.FromUserID(ctx))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags SportsRecordAPI
// @Security ApiKeyAuth
// @Summary Update record
// @Router /api/v1/sports/records/{id} [put]
func (a *Record) Update(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.RecordForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}
	if err := a.RecordBIZ.Update(ctx, c.Param("id"), item); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags SportsRecordAPI
// @Security ApiKeyAuth
// @Summary Delete record
// @Router /api/v1/sports/records/{id} [delete]
func (a *Record) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	if err := a.RecordBIZ.Delete(ctx, c.Param("id")); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// ---------------- Report ----------------

// Report 评估报告
type Report struct {
	ReportBIZ *biz.Report
}

// @Tags SportsReportAPI
// @Security ApiKeyAuth
// @Summary Preview report data
// @Router /api/v1/sports/reports/preview [post]
func (a *Report) Preview(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.ReportPreviewParams)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.ReportBIZ.Preview(ctx, item)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags SportsReportAPI
// @Security ApiKeyAuth
// @Summary Query reports
// @Router /api/v1/sports/reports [get]
func (a *Report) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params schema.ReportQueryParam
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.ReportBIZ.Query(ctx, params)
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags SportsReportAPI
// @Security ApiKeyAuth
// @Summary Get report
// @Router /api/v1/sports/reports/{id} [get]
func (a *Report) Get(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := a.ReportBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags SportsReportAPI
// @Security ApiKeyAuth
// @Summary Create report
// @Router /api/v1/sports/reports [post]
func (a *Report) Create(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(schema.ReportForm)
	if err := util.ParseJSON(c, item); err != nil {
		util.ResError(c, err)
		return
	}
	result, err := a.ReportBIZ.Create(ctx, item, util.FromUserID(ctx))
	if err != nil {
		util.ResError(c, err)
		return
	}
	util.ResSuccess(c, result)
}

// @Tags SportsReportAPI
// @Security ApiKeyAuth
// @Summary Update report conclusion
// @Router /api/v1/sports/reports/{id}/conclusion [put]
func (a *Report) UpdateConclusion(c *gin.Context) {
	ctx := c.Request.Context()
	var params struct {
		Conclusion string `json:"conclusion"`
	}
	if err := util.ParseJSON(c, &params); err != nil {
		util.ResError(c, err)
		return
	}
	if err := a.ReportBIZ.Update(ctx, c.Param("id"), params.Conclusion); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

// @Tags SportsReportAPI
// @Security ApiKeyAuth
// @Summary Delete report
// @Router /api/v1/sports/reports/{id} [delete]
func (a *Report) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	if err := a.ReportBIZ.Delete(ctx, c.Param("id")); err != nil {
		util.ResError(c, err)
		return
	}
	util.ResOK(c)
}

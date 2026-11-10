package sports

import (
	"context"
	"log"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/api"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/biz"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Sports struct {
	DB          *gorm.DB
	CategoryAPI *api.Category
	ItemAPI     *api.Item
	SubjectAPI  *api.Subject
	RecordAPI   *api.Record
	ReportAPI   *api.Report
	StandardAPI *api.Standard
	SeedBIZ     *biz.Seed
}

func (a *Sports) AutoMigrate(ctx context.Context) error {
	return a.DB.AutoMigrate(
		new(schema.Category),
		new(schema.Item),
		new(schema.ItemField),
		new(schema.ItemMetric),
		new(schema.ItemRule),
		new(schema.ItemAdvice),
		new(schema.Standard),
		new(schema.Subject),
		new(schema.Record),
		new(schema.Report),
	)
}

func (a *Sports) Init(ctx context.Context) error {
	if err := a.AutoMigrate(ctx); err != nil {
		return err
	}
	if err := a.SeedBIZ.Init(ctx); err != nil {
		logging.Context(ctx).Error("sports seed data failed", zap.Error(err))
	}
	return nil
}

func (a *Sports) RegisterV1Routers(ctx context.Context, v1 *gin.RouterGroup) error {
	sports := v1.Group("sports")
	{
		categories := sports.Group("categories")
		{
			categories.GET("tree", a.CategoryAPI.QueryTree)
			categories.POST("", a.CategoryAPI.Create)
			categories.PUT(":id", a.CategoryAPI.Update)
			categories.DELETE(":id", a.CategoryAPI.Delete)
		}

		items := sports.Group("items")
		{
			items.GET("", a.ItemAPI.Query)
			items.GET(":id", a.ItemAPI.Get)
			items.POST("", a.ItemAPI.Create)
			items.PUT(":id", a.ItemAPI.Update)
			items.DELETE(":id", a.ItemAPI.Delete)
			items.POST(":id/copy", a.ItemAPI.Copy)
			items.POST(":id/eval", a.ItemAPI.Eval)
			items.POST("metrics/validate", a.ItemAPI.ValidateMetric)
		}

		standards := sports.Group("standards")
		{
			standards.GET("", a.StandardAPI.Query)
			standards.POST("", a.StandardAPI.Create)
			standards.POST("batch", a.StandardAPI.BatchCreate)
			standards.DELETE(":id", a.StandardAPI.Delete)
		}

		subjects := sports.Group("subjects")
		{
			subjects.GET("", a.SubjectAPI.Query)
			subjects.GET(":id", a.SubjectAPI.Get)
			subjects.POST("", a.SubjectAPI.Create)
			subjects.PUT(":id", a.SubjectAPI.Update)
			subjects.DELETE(":id", a.SubjectAPI.Delete)
		}

		records := sports.Group("records")
		{
			records.GET("", a.RecordAPI.Query)
			records.POST("", a.RecordAPI.Create)
			records.PUT(":id", a.RecordAPI.Update)
			records.DELETE(":id", a.RecordAPI.Delete)
		}

		reports := sports.Group("reports")
		{
			reports.POST("preview", a.ReportAPI.Preview)
			reports.GET("", a.ReportAPI.Query)
			reports.GET(":id", a.ReportAPI.Get)
			reports.POST("", a.ReportAPI.Create)
			reports.PUT(":id/conclusion", a.ReportAPI.UpdateConclusion)
			reports.DELETE(":id", a.ReportAPI.Delete)
		}
	}
	return nil
}

func (a *Sports) Release(ctx context.Context) error {
	return nil
}

var _ = log.Println

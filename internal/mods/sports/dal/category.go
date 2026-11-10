package dal

import (
	"context"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

func getCategoryDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Category))
}

type Category struct {
	DB *gorm.DB
}

func (a *Category) Query(ctx context.Context, params schema.CategoryQueryParam) (schema.Categories, error) {
	db := getCategoryDB(ctx, a.DB)
	if v := params.LikeName; v != "" {
		db = db.Where("name LIKE ?", "%"+v+"%")
	}
	if v := params.Status; v != "" {
		db = db.Where("status = ?", v)
	}

	var list schema.Categories
	if err := db.Order("sequence DESC").Order("created_at ASC").Find(&list).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return list, nil
}

func (a *Category) Get(ctx context.Context, id string) (*schema.Category, error) {
	item := new(schema.Category)
	ok, err := util.FindOne(ctx, getCategoryDB(ctx, a.DB).Where("id=?", id), util.QueryOptions{}, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

func (a *Category) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, getCategoryDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

func (a *Category) Create(ctx context.Context, item *schema.Category) error {
	return errors.WithStack(getCategoryDB(ctx, a.DB).Create(item).Error)
}

func (a *Category) Update(ctx context.Context, item *schema.Category) error {
	result := getCategoryDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

func (a *Category) Delete(ctx context.Context, id string) error {
	return errors.WithStack(getCategoryDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.Category)).Error)
}

func (a *Category) CountChildren(ctx context.Context, id string) (int64, error) {
	var count int64
	err := getCategoryDB(ctx, a.DB).Where("parent_id=?", id).Count(&count).Error
	return count, errors.WithStack(err)
}

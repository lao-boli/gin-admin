package biz

import (
	"context"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// Category 分类树管理（三个根节点禁止删除；规范/人群节点允许改名与停用）
type Category struct {
	Trans     *util.Trans
	CategoryDAL *dal.Category
	ItemDAL   *dal.Item
}

func (a *Category) Query(ctx context.Context, params schema.CategoryQueryParam) (schema.Categories, error) {
	list, err := a.CategoryDAL.Query(ctx, params)
	if err != nil {
		return nil, err
	}
	// 统计每个节点下挂的项目数（录入页卡片展示）
	if counts, err := a.ItemDAL.CountGroupByCategory(ctx); err == nil {
		for _, item := range list {
			item.ItemCount = int(counts[item.ID])
		}
	}
	return list.ToTree(), nil
}

func (a *Category) Create(ctx context.Context, formItem *schema.CategoryForm) (*schema.Category, error) {
	if formItem.ParentID != "" {
		exists, err := a.CategoryDAL.Exists(ctx, formItem.ParentID)
		if err != nil {
			return nil, err
		} else if !exists {
			return nil, errors.NotFound("", "父节点不存在")
		}
	}
	category := &schema.Category{
		ID:        util.NewXID(),
		ParentID:  formItem.ParentID,
		CreatedAt: time.Now(),
	}
	if err := formItem.FillTo(category); err != nil {
		return nil, err
	}
	if category.Type == "" {
		category.Type = schema.CategoryTypeGroup
	}
	if err := a.CategoryDAL.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (a *Category) Update(ctx context.Context, id string, formItem *schema.CategoryForm) error {
	category, err := a.CategoryDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if category == nil {
		return errors.NotFound("", "分类不存在")
	}

	// 根节点禁止改父级（防止挂到子树成环），允许改名/停用
	if category.IsSystem && formItem.ParentID != category.ParentID {
		return errors.BadRequest("", "系统内置节点不允许移动")
	}
	if formItem.ParentID == id {
		return errors.BadRequest("", "父节点不能是自己")
	}

	if err := formItem.FillTo(category); err != nil {
		return err
	}
	if formItem.ParentID != "" {
		category.ParentID = formItem.ParentID
	}
	category.UpdatedAt = time.Now()
	return a.CategoryDAL.Update(ctx, category)
}

func (a *Category) Delete(ctx context.Context, id string) error {
	category, err := a.CategoryDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if category == nil {
		return errors.NotFound("", "分类不存在")
	}
	if category.IsSystem {
		return errors.BadRequest("", "系统内置节点不允许删除（可停用）")
	}

	if count, err := a.CategoryDAL.CountChildren(ctx, id); err != nil {
		return err
	} else if count > 0 {
		return errors.BadRequest("", "该节点下存在子节点，请先删除子节点")
	}
	if count, err := a.ItemDAL.CountByCategory(ctx, id); err != nil {
		return err
	} else if count > 0 {
		return errors.BadRequest("", "该节点下存在测试项目，请先删除项目")
	}
	return a.CategoryDAL.Delete(ctx, id)
}

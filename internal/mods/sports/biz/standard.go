package biz

import (
	"context"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// Standard 评分标准管理
type Standard struct {
	Trans       *util.Trans
	StandardDAL *dal.Standard
}

func (a *Standard) Query(ctx context.Context, params schema.StandardQueryParam) (schema.Standards, error) {
	list, err := a.StandardDAL.Query(ctx, params)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make(schema.Standards, 0)
	}
	return list, nil
}

func (a *Standard) Create(ctx context.Context, formItem *schema.StandardForm) (*schema.Standard, error) {
	item := &schema.Standard{
		ID:        util.NewXID(),
		ItemID:    formItem.ItemID,
		Gender:    formItem.Gender,
		AgeMin:    formItem.AgeMin,
		AgeMax:    formItem.AgeMax,
		Score:     formItem.Score,
		Grade:     formItem.Grade,
		LowerBound: formItem.LowerBound,
		UpperBound: formItem.UpperBound,
		CreatedAt: time.Now(),
	}
	if err := a.StandardDAL.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// BatchCreate 批量替换某项目的标准（先清后灌）
func (a *Standard) BatchCreate(ctx context.Context, formItem *schema.StandardBatchForm) error {
	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.StandardDAL.DeleteByItemID(ctx, formItem.ItemID); err != nil {
			return err
		}
		now := time.Now()
		for _, f := range formItem.Standards {
			item := &schema.Standard{
				ID:         util.NewXID(),
				ItemID:     formItem.ItemID,
				Gender:     f.Gender,
				AgeMin:     f.AgeMin,
				AgeMax:     f.AgeMax,
				Score:      f.Score,
				Grade:      f.Grade,
				LowerBound: f.LowerBound,
				UpperBound: f.UpperBound,
				CreatedAt:  now,
			}
			if err := a.StandardDAL.Create(ctx, item); err != nil {
				return err
			}
		}
		return nil
	})
}

func (a *Standard) Delete(ctx context.Context, id string) error {
	item, err := a.StandardDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if item == nil {
		return errors.NotFound("", "标准不存在")
	}
	return a.StandardDAL.Delete(ctx, id)
}

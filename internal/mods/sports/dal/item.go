package dal

import (
	"context"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

func getItemDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Item))
}

type Item struct {
	DB *gorm.DB
}

func (a *Item) Query(ctx context.Context, params schema.ItemQueryParam) (schema.Items, error) {
	db := getItemDB(ctx, a.DB)
	if v := params.CategoryID; v != "" {
		db = db.Where("category_id = ?", v)
	}
	if v := params.LikeName; v != "" {
		db = db.Where("name LIKE ?", "%"+v+"%")
	}
	if v := params.Status; v != "" {
		db = db.Where("status = ?", v)
	}

	var list schema.Items
	if err := db.Order("sequence DESC").Order("created_at ASC").Find(&list).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return list, nil
}

func (a *Item) Get(ctx context.Context, id string) (*schema.Item, error) {
	item := new(schema.Item)
	ok, err := util.FindOne(ctx, getItemDB(ctx, a.DB).Where("id=?", id), util.QueryOptions{}, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

func (a *Item) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, getItemDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

func (a *Item) Create(ctx context.Context, item *schema.Item) error {
	return errors.WithStack(getItemDB(ctx, a.DB).Create(item).Error)
}

func (a *Item) Update(ctx context.Context, item *schema.Item) error {
	result := getItemDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

func (a *Item) Delete(ctx context.Context, id string) error {
	return errors.WithStack(getItemDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.Item)).Error)
}

func (a *Item) CountByCategory(ctx context.Context, categoryID string) (int64, error) {
	var count int64
	err := getItemDB(ctx, a.DB).Where("category_id=?", categoryID).Count(&count).Error
	return count, errors.WithStack(err)
}

// CountGroupByCategory 按分类统计项目数（分类树展示用）
func (a *Item) CountGroupByCategory(ctx context.Context) (map[string]int64, error) {
	type row struct {
		CategoryID string
		Total      int64
	}
	var rows []row
	err := getItemDB(ctx, a.DB).
		Select("category_id, count(*) AS total").
		Group("category_id").
		Scan(&rows).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result := make(map[string]int64, len(rows))
	for _, r := range rows {
		result[r.CategoryID] = r.Total
	}
	return result, nil
}

// ---------------- ItemField ----------------

func getItemFieldDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ItemField))
}

func (a *Item) QueryFields(ctx context.Context, itemID string) (schema.ItemFields, error) {
	var list schema.ItemFields
	err := getItemFieldDB(ctx, a.DB).Where("item_id=?", itemID).Order("sequence DESC").Find(&list).Error
	return list, errors.WithStack(err)
}

func (a *Item) CreateField(ctx context.Context, field *schema.ItemField) error {
	return errors.WithStack(getItemFieldDB(ctx, a.DB).Create(field).Error)
}

func (a *Item) DeleteFieldsByItemID(ctx context.Context, itemID string) error {
	return errors.WithStack(getItemFieldDB(ctx, a.DB).Where("item_id=?", itemID).Delete(new(schema.ItemField)).Error)
}

// ---------------- ItemMetric ----------------

func getItemMetricDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ItemMetric))
}

func (a *Item) QueryMetrics(ctx context.Context, itemID string) (schema.ItemMetrics, error) {
	var list schema.ItemMetrics
	err := getItemMetricDB(ctx, a.DB).Where("item_id=?", itemID).Order("sequence DESC").Find(&list).Error
	return list, errors.WithStack(err)
}

func (a *Item) CreateMetric(ctx context.Context, metric *schema.ItemMetric) error {
	return errors.WithStack(getItemMetricDB(ctx, a.DB).Create(metric).Error)
}

func (a *Item) DeleteMetricsByItemID(ctx context.Context, itemID string) error {
	return errors.WithStack(getItemMetricDB(ctx, a.DB).Where("item_id=?", itemID).Delete(new(schema.ItemMetric)).Error)
}

// ---------------- ItemRule ----------------

func getItemRuleDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ItemRule))
}

func (a *Item) GetRule(ctx context.Context, itemID string) (*schema.ItemRule, error) {
	rule := new(schema.ItemRule)
	ok, err := util.FindOne(ctx, getItemRuleDB(ctx, a.DB).Where("item_id=?", itemID), util.QueryOptions{}, rule)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return rule, nil
}

func (a *Item) CreateRule(ctx context.Context, rule *schema.ItemRule) error {
	return errors.WithStack(getItemRuleDB(ctx, a.DB).Create(rule).Error)
}

func (a *Item) UpdateRule(ctx context.Context, rule *schema.ItemRule) error {
	result := getItemRuleDB(ctx, a.DB).Where("id=?", rule.ID).Select("*").Omit("created_at").Updates(rule)
	return errors.WithStack(result.Error)
}

func (a *Item) DeleteRuleByItemID(ctx context.Context, itemID string) error {
	return errors.WithStack(getItemRuleDB(ctx, a.DB).Where("item_id=?", itemID).Delete(new(schema.ItemRule)).Error)
}

// ---------------- ItemAdvice ----------------

func getItemAdviceDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.ItemAdvice))
}

func (a *Item) QueryAdvices(ctx context.Context, itemID string) (schema.ItemAdvices, error) {
	var list schema.ItemAdvices
	err := getItemAdviceDB(ctx, a.DB).Where("item_id=?", itemID).Order("sequence DESC").Find(&list).Error
	return list, errors.WithStack(err)
}

func (a *Item) CreateAdvice(ctx context.Context, advice *schema.ItemAdvice) error {
	return errors.WithStack(getItemAdviceDB(ctx, a.DB).Create(advice).Error)
}

func (a *Item) DeleteAdvicesByItemID(ctx context.Context, itemID string) error {
	return errors.WithStack(getItemAdviceDB(ctx, a.DB).Where("item_id=?", itemID).Delete(new(schema.ItemAdvice)).Error)
}

// ---------------- Standard ----------------

func getStandardDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Standard))
}

type Standard struct {
	DB *gorm.DB
}

func (a *Standard) Query(ctx context.Context, params schema.StandardQueryParam) (schema.Standards, error) {
	var list schema.Standards
	err := getStandardDB(ctx, a.DB).Where("item_id=?", params.ItemID).Order("score DESC").Find(&list).Error
	return list, errors.WithStack(err)
}

func (a *Standard) Get(ctx context.Context, id string) (*schema.Standard, error) {
	item := new(schema.Standard)
	ok, err := util.FindOne(ctx, getStandardDB(ctx, a.DB).Where("id=?", id), util.QueryOptions{}, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

func (a *Standard) Create(ctx context.Context, item *schema.Standard) error {
	return errors.WithStack(getStandardDB(ctx, a.DB).Create(item).Error)
}

func (a *Standard) Delete(ctx context.Context, id string) error {
	return errors.WithStack(getStandardDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.Standard)).Error)
}

func (a *Standard) DeleteByItemID(ctx context.Context, itemID string) error {
	return errors.WithStack(getStandardDB(ctx, a.DB).Where("item_id=?", itemID).Delete(new(schema.Standard)).Error)
}

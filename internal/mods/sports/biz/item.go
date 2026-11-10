package biz

import (
	"context"
	"strings"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// Item 测试项目管理（含字段/指标/规则配置与跨人群复制）
type Item struct {
	Trans       *util.Trans
	ItemDAL     *dal.Item
	StandardDAL *dal.Standard
	EvalEngine  *EvalEngine
}

func (a *Item) Query(ctx context.Context, params schema.ItemQueryParam) (schema.Items, error) {
	list, err := a.ItemDAL.Query(ctx, params)
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		a.fillConfig(ctx, item, false)
	}
	return list, nil
}

func (a *Item) Get(ctx context.Context, id string) (*schema.Item, error) {
	item, err := a.ItemDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if item == nil {
		return nil, errors.NotFound("", "项目不存在")
	}
	if err := a.fillConfig(ctx, item, true); err != nil {
		return nil, err
	}
	return item, nil
}

func (a *Item) fillConfig(ctx context.Context, item *schema.Item, withRule bool) error {
	fields, err := a.ItemDAL.QueryFields(ctx, item.ID)
	if err != nil {
		return err
	}
	item.Fields = fields
	metrics, err := a.ItemDAL.QueryMetrics(ctx, item.ID)
	if err != nil {
		return err
	}
	item.Metrics = metrics
	advices, err := a.ItemDAL.QueryAdvices(ctx, item.ID)
	if err != nil {
		return err
	}
	item.Advices = advices
	if withRule {
		rule, err := a.ItemDAL.GetRule(ctx, item.ID)
		if err != nil {
			return err
		}
		item.Rule = rule
	}
	return nil
}

// validateForm 校验表单中的公式
func (a *Item) validateForm(formItem *schema.ItemForm) error {
	for _, m := range formItem.Metrics {
		if err := a.EvalEngine.ValidateFormula(m.Formula); err != nil {
			return errors.BadRequest("", "指标 %s 公式有误: %s", m.Name, err.Error())
		}
	}
	return nil
}

func (a *Item) Create(ctx context.Context, formItem *schema.ItemForm) (*schema.Item, error) {
	if err := a.validateForm(formItem); err != nil {
		return nil, err
	}
	item := &schema.Item{
		ID:         util.NewXID(),
		CategoryID: formItem.CategoryID,
		Name:       formItem.Name,
		Unit:       formItem.Unit,
		MetricType: formItem.MetricType,
		Sequence:   formItem.Sequence,
		Status:     formItem.Status,
		CreatedAt:  time.Now(),
	}
	err := a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.ItemDAL.Create(ctx, item); err != nil {
			return err
		}
		return a.saveConfig(ctx, item.ID, formItem.Fields, formItem.Metrics, formItem.Rule, formItem.Advices)
	})
	if err != nil {
		return nil, err
	}
	return a.Get(ctx, item.ID)
}

func (a *Item) saveConfig(ctx context.Context, itemID string, fields schema.ItemFields, metrics schema.ItemMetrics, rule *schema.ItemRule, advices schema.ItemAdvices) error {
	if err := a.ItemDAL.DeleteFieldsByItemID(ctx, itemID); err != nil {
		return err
	}
	for i, f := range fields {
		if f.ID == "" {
			f.ID = util.NewXID()
		}
		f.ItemID = itemID
		if f.CreatedAt.IsZero() {
			f.CreatedAt = time.Now()
		}
		f.UpdatedAt = time.Now()
		f.Sequence = len(fields) - i
		if err := a.ItemDAL.CreateField(ctx, f); err != nil {
			return err
		}
	}

	if err := a.ItemDAL.DeleteMetricsByItemID(ctx, itemID); err != nil {
		return err
	}
	for i, m := range metrics {
		if m.ID == "" {
			m.ID = util.NewXID()
		}
		m.ItemID = itemID
		if m.CreatedAt.IsZero() {
			m.CreatedAt = time.Now()
		}
		m.UpdatedAt = time.Now()
		m.Sequence = len(metrics) - i
		if err := a.ItemDAL.CreateMetric(ctx, m); err != nil {
			return err
		}
	}

	// rule: 全删全插
	if err := a.ItemDAL.DeleteRuleByItemID(ctx, itemID); err != nil {
		return err
	}
	if rule != nil && rule.RuleType != "" {
		rule.ID = util.NewXID()
		rule.ItemID = itemID
		rule.CreatedAt = time.Now()
		rule.UpdatedAt = time.Now()
		if err := a.ItemDAL.CreateRule(ctx, rule); err != nil {
			return err
		}
	}

	// advices: 全删全插
	if err := a.ItemDAL.DeleteAdvicesByItemID(ctx, itemID); err != nil {
		return err
	}
	for i, adv := range advices {
		if strings.TrimSpace(adv.Advice) == "" {
			continue
		}
		if adv.ID == "" {
			adv.ID = util.NewXID()
		}
		adv.ItemID = itemID
		if adv.CreatedAt.IsZero() {
			adv.CreatedAt = time.Now()
		}
		adv.UpdatedAt = time.Now()
		adv.Sequence = len(advices) - i
		if err := a.ItemDAL.CreateAdvice(ctx, adv); err != nil {
			return err
		}
	}
	return nil
}

func (a *Item) Update(ctx context.Context, id string, formItem *schema.ItemForm) error {
	item, err := a.ItemDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if item == nil {
		return errors.NotFound("", "项目不存在")
	}
	if err := a.validateForm(formItem); err != nil {
		return err
	}

	item.CategoryID = formItem.CategoryID
	item.Name = formItem.Name
	item.Unit = formItem.Unit
	item.MetricType = formItem.MetricType
	item.Sequence = formItem.Sequence
	item.Status = formItem.Status
	item.UpdatedAt = time.Now()

	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.ItemDAL.Update(ctx, item); err != nil {
			return err
		}
		return a.saveConfig(ctx, id, formItem.Fields, formItem.Metrics, formItem.Rule, formItem.Advices)
	})
}

func (a *Item) Delete(ctx context.Context, id string) error {
	item, err := a.ItemDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if item == nil {
		return errors.NotFound("", "项目不存在")
	}
	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.ItemDAL.Delete(ctx, id); err != nil {
			return err
		}
		if err := a.ItemDAL.DeleteFieldsByItemID(ctx, id); err != nil {
			return err
		}
		if err := a.ItemDAL.DeleteMetricsByItemID(ctx, id); err != nil {
			return err
		}
		if err := a.ItemDAL.DeleteRuleByItemID(ctx, id); err != nil {
			return err
		}
		if err := a.ItemDAL.DeleteAdvicesByItemID(ctx, id); err != nil {
			return err
		}
		return a.StandardDAL.DeleteByItemID(ctx, id)
	})
}

// Copy 复制项目（连带字段/指标/规则/标准）到目标分类
func (a *Item) Copy(ctx context.Context, id string, targetCategoryID string) (*schema.Item, error) {
	source, err := a.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	formItem := &schema.ItemForm{
		CategoryID: targetCategoryID,
		Name:       source.Name,
		Unit:       source.Unit,
		MetricType: source.MetricType,
		Sequence:   source.Sequence,
		Status:     schema.CategoryStatusEnabled,
		Fields:     source.Fields,
		Metrics:    source.Metrics,
		Rule:       source.Rule,
		Advices:    copyAdvices(source.Advices),
	}
	if rule := formItem.Rule; rule != nil {
		cp := *rule
		cp.ID = ""
		cp.ItemID = ""
		formItem.Rule = &cp
	}
	return a.Create(ctx, formItem)
}

// ValidateMetric 公式校验
func (a *Item) ValidateMetric(ctx context.Context, formula string) error {
	if err := a.EvalEngine.ValidateFormula(formula); err != nil {
		return errors.BadRequest("", "公式有误: %s", err.Error())
	}
	return nil
}

// Eval 试算：字段值 → 指标 → 得分
func (a *Item) Eval(ctx context.Context, itemID string, subject *schema.Subject, fields map[string]interface{}) (*schema.RecordEvalResult, error) {
	item, err := a.Get(ctx, itemID)
	if err != nil {
		return nil, err
	} else if item == nil {
		return nil, errors.NotFound("", "项目不存在")
	}
	standards, err := a.StandardDAL.Query(ctx, schema.StandardQueryParam{ItemID: itemID})
	if err != nil {
		return nil, err
	}
	return a.EvalEngine.Evaluate(item, standards, subject, fields, time.Now())
}

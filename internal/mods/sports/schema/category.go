package schema

import (
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
)

const (
	CategoryStatusEnabled  = "enabled"
	CategoryStatusDisabled = "disabled"

	// Node type markers (informational only, tree is free-form)
	CategoryTypeAudience = "audience" // 人群节点（幼儿/小学/中学/大学）
	CategoryTypeSpec     = "spec"     // 运动能力测试规范节点
	CategoryTypeGroup    = "group"    // 普通分组
)

// Category 测试分类树节点（三个根节点系统内置禁止删除；规范节点允许改名/停用）
type Category struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`
	ParentID  string    `json:"parent_id" gorm:"size:20;index;"`
	Name      string    `json:"name" gorm:"size:64;index;"`
	Type      string    `json:"type" gorm:"size:20;index;"` // group/audience/spec
	Sequence  int       `json:"sequence" gorm:"index;"`
	IsSystem  bool      `json:"is_system"`                  // 禁止删除（根节点及预置节点）
	Status    string    `json:"status" gorm:"size:20;index;"`
	CreatedAt time.Time `json:"created_at" gorm:"index;"`
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`
	Children  Categories `json:"children,omitempty" gorm:"-"`
	ItemCount int        `json:"item_count" gorm:"-"` // 该节点下挂的测试项目数（查询时统计）
}

func (a *Category) TableName() string {
	return config.C.FormatTableName("sports_category")
}

type Categories []*Category

func (a Categories) ToTree() Categories {
	m := make(map[string]*Category, len(a))
	for _, item := range a {
		m[item.ID] = item
	}
	var roots Categories
	for _, item := range a {
		if item.ParentID == "" || item.ParentID == "0" {
			roots = append(roots, item)
			continue
		}
		if parent, ok := m[item.ParentID]; ok {
			if parent.Children == nil {
				parent.Children = make(Categories, 0)
			}
			parent.Children = append(parent.Children, item)
		}
	}
	return roots
}

// Category with children (for tree output)
type CategoryTree struct {
	Category
	Children Categories `json:"children,omitempty" gorm:"-"`
}

type CategoryQueryParam struct {
	LikeName string `form:"name"`
	Status   string `form:"status" binding:"oneof=disabled enabled ''"`
}

type CategoryForm struct {
	ParentID string `json:"parent_id" binding:"omitempty,max=20"`
	Name     string `json:"name" binding:"required,max=64"`
	Type     string `json:"type" binding:"omitempty,oneof=audience group spec"`
	Sequence int    `json:"sequence"`
	Status   string `json:"status" binding:"required,oneof=disabled enabled"`
}

func (a *CategoryForm) FillTo(category *Category) error {
	category.Name = a.Name
	category.Type = a.Type
	category.Sequence = a.Sequence
	category.Status = a.Status
	return nil
}

// ---------------- 测试项目 ----------------

// Item 测试项目（挂在分类节点下）
type Item struct {
	ID         string    `json:"id" gorm:"size:20;primarykey;"`
	CategoryID string    `json:"category_id" gorm:"size:20;index;"`
	Name       string    `json:"name" gorm:"size:64;index;"`
	Unit       string    `json:"unit" gorm:"size:16;"`
	MetricType string    `json:"metric_type" gorm:"size:16;"` // numeric/timer/counter/level
	Sequence   int       `json:"sequence" gorm:"index;"`
	Status     string    `json:"status" gorm:"size:20;index;"`
	CreatedAt  time.Time `json:"created_at" gorm:"index;"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"index;"`

	// nested config
	Fields  ItemFields  `json:"fields" gorm:"-"`
	Metrics ItemMetrics `json:"metrics" gorm:"-"`
	Rule    *ItemRule   `json:"rule" gorm:"-"`
	Advices ItemAdvices `json:"advices" gorm:"-"`
}

func (a *Item) TableName() string {
	return config.C.FormatTableName("sports_item")
}

type Items []*Item

type ItemQueryParam struct {
	CategoryID string `form:"categoryId"`
	LikeName   string `form:"name"`
	Status     string `form:"status" binding:"oneof=disabled enabled ''"`
}

type ItemForm struct {
	CategoryID string      `json:"category_id" binding:"required,max=20"`
	Name       string      `json:"name" binding:"required,max=64"`
	Unit       string      `json:"unit" binding:"max=16"`
	MetricType string      `json:"metric_type" binding:"required,oneof=counter level numeric timer"`
	Sequence   int         `json:"sequence"`
	Status     string      `json:"status" binding:"required,oneof=disabled enabled"`
	Fields     ItemFields  `json:"fields"`
	Metrics    ItemMetrics `json:"metrics"`
	Rule       *ItemRule   `json:"rule"`
	Advices    ItemAdvices `json:"advices"`
}

// ItemField 输入字段
type ItemField struct {
	ID           string     `json:"id" gorm:"size:20;primarykey;"`
	ItemID       string     `json:"item_id" gorm:"size:20;index;"`
	Key          string     `json:"key" gorm:"size:32;index;"`
	Name         string     `json:"name" gorm:"size:64;"`
	FieldType    string     `json:"field_type" gorm:"size:16;"` // numeric/timer/counter/level
	Unit         string     `json:"unit" gorm:"size:16;"`
	Required     bool       `json:"required"`
	DefaultValue *float64   `json:"default_value" gorm:"type:numeric(12,4);"`
	Sequence     int        `json:"sequence" gorm:"index;"`
	CreatedAt    time.Time  `json:"created_at" gorm:"index;"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"index;"`
}

func (a *ItemField) TableName() string {
	return config.C.FormatTableName("sports_item_field")
}

type ItemFields []*ItemField

// ItemMetric 派生指标（公式）
type ItemMetric struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`
	ItemID    string    `json:"item_id" gorm:"size:20;index;"`
	Key       string    `json:"key" gorm:"size:32;index;"`
	Name      string    `json:"name" gorm:"size:64;"`
	Formula   string    `json:"formula" gorm:"type:text;"` // expr 表达式
	Unit      string    `json:"unit" gorm:"size:16;"`
	InReport  bool      `json:"in_report"`
	Sequence  int       `json:"sequence" gorm:"index;"`
	CreatedAt time.Time `json:"created_at" gorm:"index;"`
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`
}

func (a *ItemMetric) TableName() string {
	return config.C.FormatTableName("sports_item_metric")
}

type ItemMetrics []*ItemMetric

// ItemRule 评分规则（每项目一条）
type ItemRule struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`
	ItemID    string    `json:"item_id" gorm:"size:20;uniqueIndex;"`
	RuleType  string    `json:"rule_type" gorm:"size:16;"` // lookup/formula/fit
	RefKey    string    `json:"ref_key" gorm:"size:32;"`   // 参与评分的字段/指标 key
	Config    JSONB     `json:"config" gorm:"type:jsonb;"`
	CreatedAt time.Time `json:"created_at" gorm:"index;"`
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`
}

func (a *ItemRule) TableName() string {
	return config.C.FormatTableName("sports_item_rule")
}

// ItemAdvice 改进建议：按指标/得分的数值区间匹配，命中的建议出现在录入页与评估报告中
type ItemAdvice struct {
	ID         string    `json:"id" gorm:"size:20;primarykey;"`
	ItemID     string    `json:"item_id" gorm:"size:20;index;"`
	RefKey     string    `json:"ref_key" gorm:"size:32;"` // score=得分，或字段/指标 key
	LowerBound *float64  `json:"lower_bound" gorm:"type:numeric(12,4);"`
	UpperBound *float64  `json:"upper_bound" gorm:"type:numeric(12,4);"`
	Advice     string    `json:"advice" gorm:"type:text;"`
	Sequence   int       `json:"sequence" gorm:"index;"`
	CreatedAt  time.Time `json:"created_at" gorm:"index;"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"index;"`
}

func (a *ItemAdvice) TableName() string {
	return config.C.FormatTableName("sports_item_advice")
}

type ItemAdvices []*ItemAdvice

// ---------------- 评分标准 ----------------

type Standard struct {
	ID         string     `json:"id" gorm:"size:20;primarykey;"`
	ItemID     string     `json:"item_id" gorm:"size:20;index;"`
	Gender     int        `json:"gender" gorm:"index;"` // 0=不限 1男 2女
	AgeMin     *int       `json:"age_min"`
	AgeMax     *int       `json:"age_max"`
	Score      int        `json:"score"`
	Grade      string     `json:"grade" gorm:"size:8;"`
	LowerBound *float64   `json:"lower_bound" gorm:"type:numeric(10,2);"`
	UpperBound *float64   `json:"upper_bound" gorm:"type:numeric(10,2);"`
	CreatedAt  time.Time  `json:"created_at" gorm:"index;"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"index;"`
}

func (a *Standard) TableName() string {
	return config.C.FormatTableName("sports_standard")
}

type Standards []*Standard

// Point 拟合点
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type StandardQueryParam struct {
	ItemID string `form:"itemId" binding:"required"`
}

type StandardForm struct {
	ItemID     string   `json:"item_id" binding:"required,max=20"`
	Gender     int      `json:"gender"`
	AgeMin     *int     `json:"age_min"`
	AgeMax     *int     `json:"age_max"`
	Score      int      `json:"score" binding:"gte=0,lte=100"`
	Grade      string   `json:"grade" binding:"max=8"`
	LowerBound *float64 `json:"lower_bound"`
	UpperBound *float64 `json:"upper_bound"`
}

type StandardBatchForm struct {
	ItemID    string         `json:"item_id" binding:"required,max=20"`
	Standards []StandardForm `json:"standards"`
}

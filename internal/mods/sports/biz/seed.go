package biz

import (
	"context"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

// Seed 种子数据：三个根 + 人群节点 + 22 规范 + 国标项目 + 20米往返跑样例
type Seed struct {
	DB *gorm.DB
}

func (a *Seed) Init(ctx context.Context) error {
	categoryDAL := &dal.Category{DB: a.DB}
	list, err := categoryDAL.Query(ctx, schema.CategoryQueryParam{})
	if err != nil {
		return err
	}
	if len(list) > 0 {
		return nil // 已有数据，跳过
	}

	itemDAL := &dal.Item{DB: a.DB}
	standardDAL := &dal.Standard{DB: a.DB}

	return a.DB.Transaction(func(tx *gorm.DB) error {
		s := &seedHelper{ctx: ctx, categoryDAL: categoryDAL, itemDAL: itemDAL, standardDAL: standardDAL}

		// ===== 全民体测 =====
		nationRoot := s.category("全民体测", "", schema.CategoryTypeGroup, 100, true)
		preschool := s.category("幼儿", nationRoot, schema.CategoryTypeAudience, 100, true)
		primary := s.category("小学", nationRoot, schema.CategoryTypeAudience, 90, true)
		middle := s.category("中学", nationRoot, schema.CategoryTypeAudience, 80, true)
		college := s.category("大学", nationRoot, schema.CategoryTypeAudience, 70, true)

		// ===== 运动能力测试：22 个规范 =====
		specRoot := s.category("运动能力测试", "", schema.CategoryTypeGroup, 90, true)
		specs := []string{
			"足球课程学生运动能力测评规范", "乒乓球课程学生运动能力测评规范", "羽毛球课程学生运动能力测评规范",
			"篮球课程学生运动能力测评规范", "软式棒垒球课程学生运动能力测评规范", "滑板课程学生运动能力测评规范",
			"健美操课程学生运动能力测评规范", "滑雪课程学生运动能力测评规范", "排球课程学生运动能力测评规范",
			"滑冰课程学生运动能力测评规范", "排舞课程学生运动能力测评规范", "田径课程学生运动能力测评规范",
			"体操课程学生运动能力测评规范", "游泳课程学生运动能力测评规范", "学生基本运动能力测评规范",
			"五体球课程学生运动能力测评规范", "中国式摔跤课程学生运动能力测评规范", "跳绳课程学生运动能力测评规范",
			"轮滑课程学生运动能力测评规范", "武术课程学生运动能力测评规范", "网球课程学生运动能力测评规范",
			"蹦床课程学生运动能力测评规范",
		}
		for i, name := range specs {
			s.category(name, specRoot, schema.CategoryTypeSpec, 100-i, true)
		}

		// ===== 老年人健康评估 =====
		elderRoot := s.category("老年人健康评估", "", schema.CategoryTypeGroup, 80, true)
		elderFit := s.category("体质测试", elderRoot, schema.CategoryTypeGroup, 100, false)
		s.category("康复运动", elderRoot, schema.CategoryTypeGroup, 90, false)

		// ===== 幼儿项目（国民体质测定标准）=====
		s.simpleItem(preschool, "身高", "cm", "numeric", 100)
		s.simpleItem(preschool, "体重", "kg", "numeric", 95)
		s.bmiItem(preschool, 90)
		s.simpleItem(preschool, "坐位体前屈", "cm", "numeric", 85)
		s.leapItem(preschool, 80)
		s.simpleItem(preschool, "网球掷远", "m", "numeric", 75)
		s.simpleItem(preschool, "双脚连续跳", "s", "timer", 70)
		s.simpleItem(preschool, "15米绕障碍跑", "s", "timer", 65)
		s.simpleItem(preschool, "走平衡木", "s", "timer", 60)
		s.shuttleRunItem(preschool, 50)

		// ===== 小学项目 =====
		s.bmiItem(primary, 100)
		s.simpleItem(primary, "肺活量", "ml", "numeric", 95)
		s.simpleItem(primary, "50米跑", "s", "timer", 90)
		s.simpleItem(primary, "坐位体前屈", "cm", "numeric", 85)
		s.simpleItem(primary, "1分钟跳绳", "次", "counter", 80)
		s.simpleItem(primary, "1分钟仰卧起坐", "次", "counter", 75)
		s.simpleItem(primary, "50米×8往返跑", "s", "timer", 70)

		// ===== 中学项目 =====
		s.bmiItem(middle, 100)
		s.simpleItem(middle, "肺活量", "ml", "numeric", 95)
		s.simpleItem(middle, "50米跑", "s", "timer", 90)
		s.simpleItem(middle, "坐位体前屈", "cm", "numeric", 85)
		s.leapItem(middle, 80)
		s.simpleItem(middle, "1分钟仰卧起坐", "次", "counter", 75)
		s.simpleItem(middle, "引体向上", "次", "counter", 70)
		s.simpleItem(middle, "800米跑", "s", "timer", 65)
		s.simpleItem(middle, "1000米跑", "s", "timer", 60)

		// ===== 大学项目 =====
		s.bmiItem(college, 100)
		s.simpleItem(college, "肺活量", "ml", "numeric", 95)
		s.simpleItem(college, "50米跑", "s", "timer", 90)
		s.simpleItem(college, "坐位体前屈", "cm", "numeric", 85)
		s.leapItem(college, 80)
		s.simpleItem(college, "1分钟仰卧起坐", "次", "counter", 75)
		s.simpleItem(college, "引体向上", "次", "counter", 70)
		s.simpleItem(college, "800米跑", "s", "timer", 65)
		s.simpleItem(college, "1000米跑", "s", "timer", 60)

		// ===== 老年人体质测试（SFT）=====
		s.simpleItem(elderFit, "握力", "kg", "numeric", 100)
		s.simpleItem(elderFit, "30秒坐站", "次", "counter", 95)
		s.simpleItem(elderFit, "2分钟原地高抬腿", "次", "counter", 90)
		s.simpleItem(elderFit, "坐位体前屈", "cm", "numeric", 85)
		s.simpleItem(elderFit, "单脚站立", "s", "timer", 80)
		s.simpleItem(elderFit, "6分钟步行", "m", "numeric", 75)

		// ===== 演示评分标准（可删）=====
		s.demoLeapStandards(preschool)
		s.demoShuttleStandards(preschool)
		return nil
	})
}

type seedHelper struct {
	ctx         context.Context
	categoryDAL *dal.Category
	itemDAL     *dal.Item
	standardDAL *dal.Standard
}

func (a *seedHelper) category(name, parentID, typ string, seq int, isSystem bool) string {
	c := &schema.Category{
		ID:        util.NewXID(),
		ParentID:  parentID,
		Name:      name,
		Type:      typ,
		Sequence:  seq,
		IsSystem:  isSystem,
		Status:    schema.CategoryStatusEnabled,
		CreatedAt: time.Now(),
	}
	if err := a.categoryDAL.Create(a.ctx, c); err != nil {
		panic(err)
	}
	return c.ID
}

func (a *seedHelper) item(categoryID, name, unit, metricType string, seq int, fields schema.ItemFields, metrics schema.ItemMetrics, rule *schema.ItemRule) *schema.Item {
	item := &schema.Item{
		ID:         util.NewXID(),
		CategoryID: categoryID,
		Name:       name,
		Unit:       unit,
		MetricType: metricType,
		Sequence:   seq,
		Status:     schema.CategoryStatusEnabled,
		CreatedAt:  time.Now(),
	}
	if err := a.itemDAL.Create(a.ctx, item); err != nil {
		panic(err)
	}
	for i, f := range fields {
		f.ID = util.NewXID()
		f.ItemID = item.ID
		f.CreatedAt = time.Now()
		f.Sequence = len(fields) - i
		if err := a.itemDAL.CreateField(a.ctx, f); err != nil {
			panic(err)
		}
	}
	for i, m := range metrics {
		m.ID = util.NewXID()
		m.ItemID = item.ID
		m.CreatedAt = time.Now()
		m.Sequence = len(metrics) - i
		if err := a.itemDAL.CreateMetric(a.ctx, m); err != nil {
			panic(err)
		}
	}
	if rule != nil {
		rule.ID = util.NewXID()
		rule.ItemID = item.ID
		rule.CreatedAt = time.Now()
		if err := a.itemDAL.CreateRule(a.ctx, rule); err != nil {
			panic(err)
		}
	}
	return item
}

// advice 预置一条改进建议（区间 [lo, hi)，nil 表示不限）
func (a *seedHelper) advice(itemID, refKey string, lo, hi *float64, text string, seq int) {
	adv := &schema.ItemAdvice{
		ID:         util.NewXID(),
		ItemID:     itemID,
		RefKey:     refKey,
		LowerBound: lo,
		UpperBound: hi,
		Advice:     text,
		Sequence:   seq,
		CreatedAt:  time.Now(),
	}
	if err := a.itemDAL.CreateAdvice(a.ctx, adv); err != nil {
		panic(err)
	}
}

func (a *seedHelper) simpleItem(categoryID, name, unit, metricType string, seq int) {	key := "value"
	a.item(categoryID, name, unit, metricType, seq, schema.ItemFields{{
		Key: key, Name: name, FieldType: metricType, Unit: unit, Required: true,
	}}, nil, &schema.ItemRule{RuleType: "lookup", RefKey: key})
}

func (a *seedHelper) bmiItem(categoryID string, seq int) {
	h := util.NewXID()
	_ = h
	a.item(categoryID, "BMI", "", "numeric", seq,
		schema.ItemFields{
			{Key: "height", Name: "身高", FieldType: "numeric", Unit: "cm", Required: true},
			{Key: "weight", Name: "体重", FieldType: "numeric", Unit: "kg", Required: true},
		},
		schema.ItemMetrics{{
			Key: "bmi", Name: "BMI", Formula: "bmi(height, weight)", InReport: true,
		}},
		&schema.ItemRule{RuleType: "lookup", RefKey: "bmi"})
}

func (a *seedHelper) leapItem(categoryID string, seq int) {
	it := a.item(categoryID, "立定跳远", "m", "numeric", seq,
		schema.ItemFields{{Key: "distance", Name: "跳远成绩", FieldType: "numeric", Unit: "m", Required: true}},
		nil, &schema.ItemRule{RuleType: "lookup", RefKey: "distance"})

	b := func(v float64) *float64 { return &v }
	a.advice(it.ID, "score", b(85), nil,
		"下肢爆发力良好，可保持每周 2 次增强式训练（跳箱、单腿跳），重点提升起跳速度与落地稳定性。", 3)
	a.advice(it.ID, "score", b(70), b(85),
		"下肢爆发力中等，建议每周 2 次力量+技术组合练习：深蹲、提踵配合摆臂起跳技术，注意落地缓冲。", 2)
	a.advice(it.ID, "score", nil, b(70),
		"下肢爆发力偏弱，建议先打基础力量（半蹲、提踵、负重台阶）并纠正起跳摆臂动作，4 周后复测。", 1)
}

func (a *seedHelper) shuttleRunItem(categoryID string, seq int) {
	f := func(key, name string, def float64) *schema.ItemField {
		return &schema.ItemField{Key: key, Name: name, FieldType: "numeric", Unit: "", Required: true, DefaultValue: &def}
	}
	it := a.item(categoryID, "20米往返跑", "", "timer", seq,
		schema.ItemFields{
			f("speed2", "第2分钟速度(m/s)", 2.0),
			f("speed4", "第4分钟速度(m/s)", 2.5),
			f("speed6", "第6分钟速度(m/s)", 3.0),
			{Key: "hr2", Name: "第2分钟平均心率", FieldType: "numeric", Unit: "bpm", Required: true},
			{Key: "hr4", Name: "第4分钟平均心率", FieldType: "numeric", Unit: "bpm", Required: true},
			{Key: "hr6", Name: "第6分钟平均心率", FieldType: "numeric", Unit: "bpm", Required: true},
		},
		schema.ItemMetrics{
			{Key: "a", Name: "截距a(基础心率)", Formula: "fit([{x: speed2, y: hr2}, {x: speed4, y: hr4}, {x: speed6, y: hr6}], 1).a", Unit: "bpm", InReport: true},
			{Key: "b", Name: "斜率b(心率-速度响应)", Formula: "fit([{x: speed2, y: hr2}, {x: speed4, y: hr4}, {x: speed6, y: hr6}], 1).b", Unit: "bpm/(m/s)", InReport: true},
			{Key: "vo2max", Name: "最大摄氧量", Formula: "31.025 + 3.238*speed6 - 3.248*age + 0.1536*speed6*age", Unit: "ml/kg/min", InReport: true},
		},
		&schema.ItemRule{RuleType: "fit", RefKey: "b"})

	mb := func(v float64) *float64 { return &v }
	a.advice(it.ID, "b", nil, mb(25),
		"心率随速度上升平缓（斜率 b 较小），心肺适应良好；可保持每周 2 次有氧训练，并尝试提高折返速度。", 3)
	a.advice(it.ID, "b", mb(25), mb(35),
		"心率-速度斜率处于中等水平，建议每周 3 次、每次 20 分钟间歇跑，关注第 2 分钟心率回落情况。", 2)
	a.advice(it.ID, "b", mb(35), nil,
		"心率随速度上升偏快（斜率 b 较大），提示心肺耐力不足；建议从低强度持续跑起步，逐步延长运动时间。", 1)
}

func (a *seedHelper) demoLeapStandards(preschoolID string) {
	// 查找该分类下立定跳远
	items, _ := a.itemDAL.Query(a.ctx, schema.ItemQueryParam{CategoryID: preschoolID, LikeName: "立定跳远"})
	if len(items) == 0 {
		return
	}
	itemID := items[0].ID
	f := func(lb, ub float64, score int, grade string) *schema.Standard {
		return &schema.Standard{ID: util.NewXID(), ItemID: itemID, Gender: 0, Score: score, Grade: grade,
			LowerBound: &lb, UpperBound: &ub, CreatedAt: time.Now()}
	}
	mk := func(v float64) *float64 { return &v }
	inf := mk(99)
	_ = inf
	for _, s := range []*schema.Standard{
		f(1.4, 99, 100, "优秀"),
		f(1.2, 1.4, 85, "良好"),
		f(1.0, 1.2, 70, "及格"),
		f(0, 1.0, 50, "待提高"),
	} {
		_ = a.standardDAL.Create(a.ctx, s)
	}
}

func (a *seedHelper) demoShuttleStandards(preschoolID string) {
	items, _ := a.itemDAL.Query(a.ctx, schema.ItemQueryParam{CategoryID: preschoolID, LikeName: "20米往返跑"})
	if len(items) == 0 {
		return
	}
	itemID := items[0].ID
	f := func(lb, ub float64, score int, grade string) *schema.Standard {
		return &schema.Standard{ID: util.NewXID(), ItemID: itemID, Gender: 0, Score: score, Grade: grade,
			LowerBound: &lb, UpperBound: &ub, CreatedAt: time.Now()}
	}
	for _, s := range []*schema.Standard{
		f(0, 1.5, 100, "优秀"),
		f(1.5, 2.0, 85, "良好"),
		f(2.0, 2.5, 70, "及格"),
		f(2.5, 99, 50, "待提高"),
	} {
		_ = a.standardDAL.Create(a.ctx, s)
	}
}

package biz

import (
	"context"
	"errors"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

// Seed 种子数据：三个根 + 人群节点 + 22 规范 + 国标项目 + 大学心肺耐力测试（极量/亚极量/20米往返跑）
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
		// 已有数据的库：只做幂等增量（补齐大学心肺耐力测试、搬移 20 米往返跑）
		return a.ensureCollegeCardio(ctx)
	}

	itemDAL := &dal.Item{DB: a.DB}
	standardDAL := &dal.Standard{DB: a.DB}

	return a.DB.Transaction(func(tx *gorm.DB) error {
		s := &seedHelper{ctx: ctx, DB: a.DB, categoryDAL: categoryDAL, itemDAL: itemDAL, standardDAL: standardDAL}

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
		// 心肺耐力三件套（对齐 trackV2 runall flavor 的极量/亚极量协议）
		s.maximalItem(college, 55)
		s.submaximalItem(college, 54)
		s.shuttleRunItem(college, 53)

		// ===== 老年人体质测试（SFT）=====
		s.simpleItem(elderFit, "握力", "kg", "numeric", 100)
		s.simpleItem(elderFit, "30秒坐站", "次", "counter", 95)
		s.simpleItem(elderFit, "2分钟原地高抬腿", "次", "counter", 90)
		s.simpleItem(elderFit, "坐位体前屈", "cm", "numeric", 85)
		s.simpleItem(elderFit, "单脚站立", "s", "timer", 80)
		s.simpleItem(elderFit, "6分钟步行", "m", "numeric", 75)

		// ===== 演示评分标准（可删）=====
		s.demoLeapStandards(preschool)
		return nil
	})
}

type seedHelper struct {
	ctx         context.Context
	DB          *gorm.DB
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

func (a *seedHelper) simpleItem(categoryID, name, unit, metricType string, seq int) {
	key := "value"
	a.item(categoryID, name, unit, metricType, seq, schema.ItemFields{{
		Key: key, Name: name, FieldType: metricType, Unit: unit, Required: true,
	}}, nil, &schema.ItemRule{RuleType: "lookup", RefKey: key})
}

func (a *seedHelper) bmiItem(categoryID string, seq int) {
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

// shuttleRunItem 20 米往返跑：分段速度 + 分段心率，外推 VO2max
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
			// 场地跑 ACSM 公式：VO2 = 3.5 + 0.2 × 速度(m/min)，speed6 单位为 m/s，故 ×60
			{Key: "vo2max", Name: "最大摄氧量", Formula: "3.5 + 0.2 * speed6 * 60", Unit: "ml/kg/min", InReport: true},
		},
		&schema.ItemRule{RuleType: "lookup", RefKey: "vo2max"})

	a.vo2maxStandards(it.ID)
	a.vo2maxAdvices(it.ID)
}

// maximalItem 极量测试：跑台/场地递增负荷 10 段 × 2min（6→15 km/h，每级 +1），测至力竭
func (a *seedHelper) maximalItem(categoryID string, seq int) {
	stages := 10.0
	it := a.item(categoryID, "极量测试", "km/h", "numeric", seq,
		schema.ItemFields{
			{Key: "max_speed", Name: "最高完成速度", FieldType: "numeric", Unit: "km/h", Required: true},
			{Key: "hr_max", Name: "实测最大心率", FieldType: "numeric", Unit: "bpm", Required: true},
			{Key: "stages", Name: "完成级数(共10级)", FieldType: "counter", Unit: "级", Required: true, DefaultValue: &stages},
		},
		schema.ItemMetrics{
			// ACSM 跑台/平地跑公式：VO2 = 3.5 + 0.2 × 速度(m/min)
			{Key: "vo2max", Name: "最大摄氧量", Formula: "3.5 + 0.2 * max_speed * 1000 / 60", Unit: "ml/kg/min", InReport: true},
			// Tanaka 预测最大心率 208 - 0.7 × 年龄，用于判断是否真正达到力竭
			{Key: "hr_reach", Name: "最大心率达成率", Formula: "round(hr_max * 100 / (208 - 0.7 * age))", Unit: "%", InReport: true},
		},
		&schema.ItemRule{RuleType: "lookup", RefKey: "vo2max"})

	a.vo2maxStandards(it.ID)
	a.vo2maxAdvices(it.ID)
}

// submaximalItem 亚极量测试：4 段 × 2min（6/7/8/9 km/h），用心率-速度回归外推到最大心率
func (a *seedHelper) submaximalItem(categoryID string, seq int) {
	f := func(key, name string, def float64) *schema.ItemField {
		return &schema.ItemField{Key: key, Name: name, FieldType: "numeric", Unit: "km/h", Required: true, DefaultValue: &def}
	}
	hr := func(key, name string) *schema.ItemField {
		return &schema.ItemField{Key: key, Name: name, FieldType: "numeric", Unit: "bpm", Required: true}
	}
	points := "[{x: speed1, y: hr1}, {x: speed2, y: hr2}, {x: speed3, y: hr3}, {x: speed4, y: hr4}]"
	it := a.item(categoryID, "亚极量测试", "km/h", "numeric", seq,
		schema.ItemFields{
			f("speed1", "第1级速度(6km/h)", 6),
			f("speed2", "第2级速度(7km/h)", 7),
			f("speed3", "第3级速度(8km/h)", 8),
			f("speed4", "第4级速度(9km/h)", 9),
			hr("hr1", "第1级末心率"),
			hr("hr2", "第2级末心率"),
			hr("hr3", "第3级末心率"),
			hr("hr4", "第4级末心率"),
		},
		schema.ItemMetrics{
			{Key: "a", Name: "截距a(基础心率)", Formula: "fit(" + points + ", 1).a", Unit: "bpm", InReport: true},
			{Key: "b", Name: "斜率b(心率-速度响应)", Formula: "fit(" + points + ", 1).b", Unit: "bpm/(km/h)", InReport: true},
			{Key: "hrmax_pred", Name: "预测最大心率", Formula: "round(208 - 0.7 * age)", Unit: "bpm"},
			// 外推：心率到预测最大值时对应的速度，上限 22 km/h 防止回归斜率过小导致虚高
			{Key: "v_max", Name: "外推最大速度", Formula: "min((208 - 0.7 * age - a) / b, 22)", Unit: "km/h", InReport: true},
			{Key: "vo2max", Name: "最大摄氧量", Formula: "3.5 + 0.2 * v_max * 1000 / 60", Unit: "ml/kg/min", InReport: true},
		},
		&schema.ItemRule{RuleType: "lookup", RefKey: "vo2max"})

	a.vo2maxStandards(it.ID)
	a.vo2maxAdvices(it.ID)
}

// vo2maxStandards 按 ACSM 成年人心肺耐力分级写入评分标准（男/女各 5 档，ml/kg/min）
func (a *seedHelper) vo2maxStandards(itemID string) {
	type band struct {
		lo, hi *float64
		score  int
		grade  string
	}
	pt := func(v float64) *float64 { return &v }

	// 男（ACSM 20-29 岁 VO2max 五分位）
	male := []band{
		{pt(51.5), nil, 100, "优秀"},
		{pt(46.5), pt(51.5), 85, "良好"},
		{pt(42.0), pt(46.5), 75, "中等"},
		{pt(35.0), pt(42.0), 60, "及格"},
		{nil, pt(35.0), 40, "待提高"},
	}
	// 女（ACSM 20-29 岁 VO2max 五分位）
	female := []band{
		{pt(43.9), nil, 100, "优秀"},
		{pt(39.5), pt(43.9), 85, "良好"},
		{pt(35.5), pt(39.5), 75, "中等"},
		{pt(29.5), pt(35.5), 60, "及格"},
		{nil, pt(29.5), 40, "待提高"},
	}

	for _, g := range []struct {
		gender int
		bands  []band
	}{{1, male}, {2, female}} {
		for _, b := range g.bands {
			s := &schema.Standard{
				ID:         util.NewXID(),
				ItemID:     itemID,
				Gender:     g.gender,
				Score:      b.score,
				Grade:      b.grade,
				LowerBound: b.lo,
				UpperBound: b.hi,
				CreatedAt:  time.Now(),
			}
			if err := a.standardDAL.Create(a.ctx, s); err != nil {
				panic(err)
			}
		}
	}
}

// vo2maxAdvices 按 VO2max 区间给出改进建议
func (a *seedHelper) vo2maxAdvices(itemID string) {
	pt := func(v float64) *float64 { return &v }
	a.advice(itemID, "vo2max", pt(46.5), nil,
		"心肺耐力处于优秀水平，建议保持每周 3 次、每次 30 分钟以上的中高强度有氧（间歇跑、骑行、游泳），并每 4 周做一次递增负荷复测跟踪。", 3)
	a.advice(itemID, "vo2max", pt(38), pt(46.5),
		"心肺耐力中等，建议每周 3 次 20-30 分钟持续跑（心率控制在最大心率的 65%-80%），每周安排 1 次 4×4 分钟高强间歇提升峰值摄氧量。", 2)
	a.advice(itemID, "vo2max", nil, pt(38),
		"心肺耐力偏弱，建议从快走/慢跑交替开始（每周 3 次、每次 20 分钟），逐步延长连续跑时间；如测试中出现胸闷、头晕应停止并就医评估。", 1)
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
	for _, s := range []*schema.Standard{
		f(1.4, 99, 100, "优秀"),
		f(1.2, 1.4, 85, "良好"),
		f(1.0, 1.2, 70, "及格"),
		f(0, 1.0, 50, "待提高"),
	} {
		_ = a.standardDAL.Create(a.ctx, s)
	}
}

// ensureCollegeCardio 对已有数据库做幂等增量：
// 1) 把 20 米往返跑从原分类搬到「全民体测 / 大学」，并把评分口径修正为 VO2max；
// 2) 大学下补齐极量测试 / 亚极量测试（含字段、指标、规则、标准、建议）。
func (a *Seed) ensureCollegeCardio(ctx context.Context) error {
	var college schema.Category
	err := a.DB.WithContext(ctx).
		Where("name = ? AND type = ?", "大学", schema.CategoryTypeAudience).
		First(&college).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	return a.DB.Transaction(func(tx *gorm.DB) error {
		// 事务内重新绑定 DAL，保证增量写入与搬家同属一个事务
		s := &seedHelper{
			ctx:         ctx,
			DB:          tx,
			categoryDAL: &dal.Category{DB: tx},
			itemDAL:     &dal.Item{DB: tx},
			standardDAL: &dal.Standard{DB: tx},
		}

		// ---- 20 米往返跑：搬家 + 修正评分口径 + 重建标准 ----
		var shuttle schema.Item
		err := tx.WithContext(ctx).Where("name = ?", "20米往返跑").First(&shuttle).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			s.shuttleRunItem(college.ID, 53)
		case err != nil:
			return err
		default:
			if shuttle.CategoryID != college.ID {
				if err := tx.Model(&schema.Item{}).Where("id = ?", shuttle.ID).
					Update("category_id", college.ID).Error; err != nil {
					return err
				}
			}
			// 评分引用改为 vo2max（旧种子用 b 斜率，标准表对不上）
			if err := tx.Model(&schema.ItemRule{}).
				Where("item_id = ?", shuttle.ID).
				Updates(map[string]interface{}{"rule_type": "lookup", "ref_key": "vo2max"}).Error; err != nil {
				return err
			}
			// 指标公式修正：speed 为 m/s，需 ×60 换成 m/min
			if err := tx.Model(&schema.ItemMetric{}).
				Where("item_id = ? AND key = ?", shuttle.ID, "vo2max").
				Update("formula", "3.5 + 0.2 * speed6 * 60").Error; err != nil {
				return err
			}
			if err := s.replaceStandards(shuttle.ID, func() { s.vo2maxStandards(shuttle.ID) }); err != nil {
				return err
			}
			// 建议也从「按斜率 b」换成「按 VO2max 区间」
			var vo2AdviceCount int64
			if err := tx.Model(&schema.ItemAdvice{}).
				Where("item_id = ? AND ref_key = ?", shuttle.ID, "vo2max").
				Count(&vo2AdviceCount).Error; err != nil {
				return err
			}
			if vo2AdviceCount == 0 {
				if err := tx.Where("item_id = ?", shuttle.ID).
					Delete(&schema.ItemAdvice{}).Error; err != nil {
					return err
				}
				s.vo2maxAdvices(shuttle.ID)
			}
		}

		// ---- 极量 / 亚极量 ----
		for _, c := range []struct {
			name string
			seq  int
			seed func(string, int)
		}{
			{"极量测试", 55, s.maximalItem},
			{"亚极量测试", 54, s.submaximalItem},
		} {
			var cnt int64
			if err := tx.Model(&schema.Item{}).
				Where("category_id = ? AND name = ?", college.ID, c.name).
				Count(&cnt).Error; err != nil {
				return err
			}
			if cnt == 0 {
				c.seed(college.ID, c.seq)
			}
		}
		return nil
	})
}

// replaceStandards 若该项目还没有分性别的标准（新口径），先清空旧标准再生成
func (a *seedHelper) replaceStandards(itemID string, gen func()) error {
	var gendered int64
	if err := a.DB.WithContext(a.ctx).Model(&schema.Standard{}).
		Where("item_id = ? AND gender <> 0", itemID).Count(&gendered).Error; err != nil {
		return err
	}
	if gendered > 0 {
		return nil
	}
	if err := a.DB.WithContext(a.ctx).Where("item_id = ?", itemID).
		Delete(&schema.Standard{}).Error; err != nil {
		return err
	}
	gen()
	return nil
}

package biz

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/expr-lang/expr"
)

// mustJSON 序列化为 JSONB
func mustJSON(v interface{}) schema.JSONB {
	b, err := json.Marshal(v)
	if err != nil {
		return schema.JSONB("null")
	}
	return schema.JSONB(b)
}

// EvalEngine 公式求值 + 评分引擎（expr 沙箱，白名单函数）
type EvalEngine struct{}

// toFloat 尽力把任意值转 float64
func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case jsonNumber:
		f, err := n.Float64()
		return f, err == nil
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(n), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

type jsonNumber interface {
	Float64() (float64, error)
}

// leastSquares 多项式最小二乘拟合：y = a0 + a1*x + a2*x^2 + ...
// 返回系数（索引即次数），并计算 R²。
func leastSquares(points []schema.Point, degree int) (coefs []float64, r2 float64, err error) {
	n := len(points)
	if n < degree+1 {
		return nil, 0, fmt.Errorf("至少需要 %d 个点做 %d 次拟合", degree+1, degree)
	}
	if degree > 3 {
		return nil, 0, fmt.Errorf("拟合次数最高支持 3 次")
	}

	m := degree + 1
	// 正规方程 A^T A c = A^T y
	ata := make([][]float64, m)
	aty := make([]float64, m)
	for i := 0; i < m; i++ {
		ata[i] = make([]float64, m)
	}
	var sx, sy float64
	var sumY2 float64
	for _, p := range points {
		pow := make([]float64, 2*m)
		pow[0] = 1
		for k := 1; k < 2*m; k++ {
			pow[k] = pow[k-1] * p.X
		}
		for i := 0; i < m; i++ {
			for j := 0; j < m; j++ {
				ata[i][j] += pow[i+j]
			}
			aty[i] += pow[i] * p.Y
		}
		sy += p.Y
		sumY2 += p.Y * p.Y
		sx += p.X
	}

	// 高斯消元
	c := make([]float64, m)
	aug := make([][]float64, m)
	for i := 0; i < m; i++ {
		row := make([]float64, m+1)
		copy(row, ata[i])
		row[m] = aty[i]
		aug[i] = row
	}
	for col := 0; col < m; col++ {
		pivot := col
		for r := col + 1; r < m; r++ {
			if math.Abs(aug[r][col]) > math.Abs(aug[pivot][col]) {
				pivot = r
			}
		}
		if math.Abs(aug[pivot][col]) < 1e-12 {
			return nil, 0, fmt.Errorf("矩阵奇异，无法拟合（点可能重复）")
		}
		aug[col], aug[pivot] = aug[pivot], aug[col]
		for r := 0; r < m; r++ {
			if r == col {
				continue
			}
			f := aug[r][col] / aug[col][col]
			for k := col; k <= m; k++ {
				aug[r][k] -= f * aug[col][k]
			}
		}
	}
	for i := 0; i < m; i++ {
		c[i] = aug[i][m] / aug[i][i]
	}

	// R²
	meanY := sy / float64(n)
	ssRes, ssTot := 0.0, 0.0
	for _, p := range points {
		pred := 0.0
		pow := 1.0
		for k := 0; k < m; k++ {
			pred += c[k] * pow
			pow *= p.X
		}
		ssRes += (p.Y - pred) * (p.Y - pred)
		ssTot += (p.Y - meanY) * (p.Y - meanY)
	}
	if ssTot > 0 {
		r2 = 1 - ssRes/ssTot
	} else {
		r2 = 1
	}
	_ = sx
	return c, r2, nil
}

// newExprOptions expr 沙箱配置：白名单内置函数
func newExprOptions() []expr.Option {
	fitFn := expr.Function("fit", func(params ...interface{}) (interface{}, error) {
		if len(params) < 1 {
			return nil, fmt.Errorf("fit 需要至少 1 个参数")
		}
		degree := 1
		if len(params) >= 2 {
			if d, ok := toFloat(params[1]); ok {
				degree = int(d)
			}
		}
		var points []schema.Point
		arr, ok := params[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("fit 第一个参数须为点集数组")
		}
		for _, item := range arr {
			m, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("点集元素须为 {x:.., y:..}")
			}
			x, okx := toFloat(m["x"])
			y, oky := toFloat(m["y"])
			if !okx || !oky {
				return nil, fmt.Errorf("点集元素 x/y 须为数值")
			}
			points = append(points, schema.Point{X: x, Y: y})
		}
		coefs, r2, err := leastSquares(points, degree)
		if err != nil {
			return nil, err
		}
		result := map[string]interface{}{}
		// a=截距, b=一次项系数, c=二次项系数, d=三次项系数
		names := []string{"a", "b", "c", "d"}
		for i, v := range coefs {
			if i < len(names) {
				result[names[i]] = v
			}
		}
		result["r2"] = r2
		return result, nil
	})

	bmiFn := expr.Function("bmi", func(params ...interface{}) (interface{}, error) {
		if len(params) != 2 {
			return nil, fmt.Errorf("bmi 需要身高(cm)、体重(kg)两个参数")
		}
		h, ok1 := toFloat(params[0])
		w, ok2 := toFloat(params[1])
		if !ok1 || !ok2 || h <= 0 {
			return nil, fmt.Errorf("bmi 参数须为正数")
		}
		hm := h / 100
		return w / (hm * hm), nil
	})

	maxFn := expr.Function("max", func(params ...interface{}) (interface{}, error) {
		m := math.Inf(-1)
		for _, p := range params {
			if v, ok := toFloat(p); ok && v > m {
				m = v
			}
		}
		return m, nil
	})

	minFn := expr.Function("min", func(params ...interface{}) (interface{}, error) {
		m := math.Inf(1)
		for _, p := range params {
			if v, ok := toFloat(p); ok && v < m {
				m = v
			}
		}
		return m, nil
	})

	absFn := expr.Function("abs", func(params ...interface{}) (interface{}, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("abs 需要 1 个参数")
		}
		v, ok := toFloat(params[0])
		if !ok {
			return nil, fmt.Errorf("abs 参数须为数值")
		}
		return math.Abs(v), nil
	})

	roundFn := expr.Function("round", func(params ...interface{}) (interface{}, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("round 需要 1 个参数")
		}
		v, ok := toFloat(params[0])
		if !ok {
			return nil, fmt.Errorf("round 参数须为数值")
		}
		return math.Round(v), nil
	})

	powFn := expr.Function("pow", func(params ...interface{}) (interface{}, error) {
		if len(params) != 2 {
			return nil, fmt.Errorf("pow 需要 2 个参数")
		}
		x, ok1 := toFloat(params[0])
		y, ok2 := toFloat(params[1])
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("pow 参数须为数值")
		}
		return math.Pow(x, y), nil
	})

	return []expr.Option{fitFn, bmiFn, maxFn, minFn, absFn, roundFn, powFn}
}

var exprOptions = newExprOptions()

// Evaluate 完整试算：字段 → 指标 → 评分
func (a *EvalEngine) Evaluate(item *schema.Item, standards schema.Standards, subject *schema.Subject, fields map[string]interface{}, testedAt time.Time) (*schema.RecordEvalResult, error) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	// 应用字段默认值（协议固定值，如 20米往返跑的测试速度）
	for _, f := range item.Fields {
		if _, ok := fields[f.Key]; !ok && f.DefaultValue != nil {
			fields[f.Key] = *f.DefaultValue
		}
	}
	env := a.BuildEnv(subject, testedAt, fields, nil)
	metrics, errs := a.EvalMetrics(item.Metrics, env)
	warnings := make([]string, 0, len(errs))
	for _, e := range errs {
		warnings = append(warnings, e.Error())
	}

	score, grade, scoreErr := a.EvalScore(item.Rule, subject, fields, metrics, standards)
	if scoreErr != nil {
		warnings = append(warnings, scoreErr.Error())
		score = nil
		grade = ""
	}
	return &schema.RecordEvalResult{
		Metrics:  metrics,
		Score:    score,
		Grade:    grade,
		Advices:  MatchAdvices(item.Advices, fields, metrics, score),
		Warnings: warnings,
	}, nil
}

// BuildEnv 组装求值环境：字段值 + 受试者属性 + 已算指标
func (a *EvalEngine) BuildEnv(subject *schema.Subject, testedAt time.Time, fields map[string]interface{}, metrics map[string]interface{}) map[string]interface{} {
	env := make(map[string]interface{})
	for k, v := range fields {
		env[k] = v
	}
	for k, v := range metrics {
		env[k] = v
	}
	if subject != nil {
		env["gender"] = subject.Gender
		if subject.BirthDate != nil {
			env["age"] = int(testedAt.Sub(*subject.BirthDate).Hours() / 24 / 365.25)
		} else {
			env["age"] = nil
		}
		env["subject_code"] = subject.Code
		env["audience_type"] = subject.AudienceType
	}
	return env
}

// EvalFormula 求值单个公式（Compile+Run，确保内置函数注册生效）
func (a *EvalEngine) EvalFormula(formula string, env map[string]interface{}) (interface{}, error) {
	if strings.TrimSpace(formula) == "" {
		return nil, fmt.Errorf("公式为空")
	}
	program, err := expr.Compile(formula, exprOptions...)
	if err != nil {
		return nil, fmt.Errorf("公式编译失败: %w", err)
	}
	out, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("公式求值失败: %w", err)
	}
	return out, nil
}

// ValidateFormula 编译校验公式（不执行）
func (a *EvalEngine) ValidateFormula(formula string) error {
	if strings.TrimSpace(formula) == "" {
		return fmt.Errorf("公式为空")
	}
	_, err := expr.Compile(formula, exprOptions...)
	return err
}

// EvalMetrics 按顺序求值项目的全部指标（后续指标可引用前面的指标）
func (a *EvalEngine) EvalMetrics(itemMetrics schema.ItemMetrics, env map[string]interface{}) (map[string]interface{}, []error) {
	result := make(map[string]interface{})
	errs := make([]error, 0)
	for _, m := range itemMetrics {
		out, err := a.EvalFormula(m.Formula, env)
		if err != nil {
			errs = append(errs, fmt.Errorf("指标 %s(%s): %w", m.Name, m.Key, err))
			result[m.Key] = nil
			env[m.Key] = nil
			continue
		}
		result[m.Key] = out
		env[m.Key] = out
	}
	return result, errs
}

// EvalScore 按评分规则算分
// ruleType=lookup/fit: refKey 指向字段或指标值 → 查分段表
// ruleType=formula: config.expr 直接算分
func (a *EvalEngine) EvalScore(rule *schema.ItemRule, subject *schema.Subject, fields, metrics map[string]interface{}, standards schema.Standards) (*int, string, error) {
	if rule == nil {
		return nil, "", nil
	}
	env := a.BuildEnv(subject, time.Now(), fields, metrics)

	var value interface{}
	switch rule.RuleType {
	case "formula":
		cfg, _ := rule.Config.Map()
		exprStr, _ := cfg["expr"].(string)
		out, err := a.EvalFormula(exprStr, env)
		if err != nil {
			return nil, "", err
		}
		if f, ok := toFloat(out); ok {
			s := int(math.Round(f))
			return &s, "", nil
		}
		return nil, "", fmt.Errorf("公式结果不是数值")
	default: // lookup / fit
		if rule.RefKey != "" {
			if v, ok := metrics[rule.RefKey]; ok {
				value = v
			} else if v, ok := fields[rule.RefKey]; ok {
				value = v
			}
		} else if len(fields) == 1 {
			for _, v := range fields {
				value = v
			}
		}
	}

	if value == nil {
		// 引用值缺失（指标求值失败或字段未填）→ 不打分，不报错，由 warnings 提示
		return nil, "", nil
	}
	f, ok := toFloat(value)
	if !ok {
		// 等级型字段：相等匹配 lower_bound? 暂不支持，直接返回无分
		return nil, "", nil
	}

	var age int = -1
	if subject != nil && subject.BirthDate != nil {
		age = int(time.Since(*subject.BirthDate).Hours() / 24 / 365.25)
	}

	var best *schema.Standard
	for i := range standards {
		s := standards[i]
		if s.Gender != 0 && subject != nil && s.Gender != subject.Gender {
			continue
		}
		if age >= 0 {
			if s.AgeMin != nil && age < *s.AgeMin {
				continue
			}
			if s.AgeMax != nil && age > *s.AgeMax {
				continue
			}
		}
		if s.LowerBound != nil && f < *s.LowerBound {
			continue
		}
		if s.UpperBound != nil && f >= *s.UpperBound {
			continue
		}
		// 取满足条件中分数最高的一条（分段重叠时取高）
		if best == nil || s.Score > best.Score {
			best = s
		}
	}

	if best == nil {
		return nil, "", nil
	}
	score := best.Score
	return &score, best.Grade, nil
}

package biz

import (
	"strings"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
)

// copyAdvices 深拷贝建议配置（复制项目时用，清空主键与归属）
func copyAdvices(in schema.ItemAdvices) schema.ItemAdvices {
	if len(in) == 0 {
		return nil
	}
	out := make(schema.ItemAdvices, 0, len(in))
	for _, a := range in {
		if a == nil {
			continue
		}
		cp := *a
		cp.ID = ""
		cp.ItemID = ""
		out = append(out, &cp)
	}
	return out
}

// MatchAdvices 按建议配置的区间匹配当前成绩，返回命中的建议文本。
// ref_key 为空或 "score" 时按得分匹配，否则优先取派生指标、其次取原始输入字段。
// 区间语义为 [lower_bound, upper_bound)，任一端为空表示不限。
func MatchAdvices(advices schema.ItemAdvices, fields, metrics map[string]interface{}, score *int) []string {
	if len(advices) == 0 {
		return nil
	}
	var out []string
	for _, a := range advices {
		if a == nil || strings.TrimSpace(a.Advice) == "" {
			continue
		}
		key := a.RefKey
		if key == "" {
			key = "score"
		}

		var (
			value float64
			ok    bool
		)
		switch {
		case key == "score":
			if score != nil {
				value, ok = float64(*score), true
			}
		default:
			if raw, exists := metrics[key]; exists {
				value, ok = toFloat(raw)
			} else if raw, exists := fields[key]; exists {
				value, ok = toFloat(raw)
			}
		}
		if !ok {
			continue
		}
		if a.LowerBound != nil && value < *a.LowerBound {
			continue
		}
		if a.UpperBound != nil && value >= *a.UpperBound {
			continue
		}
		out = append(out, a.Advice)
	}
	return out
}

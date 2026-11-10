package biz

import (
	"context"
	"encoding/json"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// Record 测试记录（单人流程式录入：选受试者 → 选项目 → 单条保存）
type Record struct {
	Trans       *util.Trans
	RecordDAL   *dal.Record
	SubjectDAL  *dal.Subject
	ItemDAL     *dal.Item
	StandardDAL *dal.Standard
	EvalEngine  *EvalEngine
}

func (a *Record) Query(ctx context.Context, params schema.RecordQueryParam) (*schema.RecordQueryResult, error) {
	return a.RecordDAL.Query(ctx, params)
}

// loadItemWithConfig 加载项目并填充 fields/metrics/rule 配置
func (a *Record) loadItemWithConfig(ctx context.Context, itemID string) (*schema.Item, error) {
	item, err := a.ItemDAL.Get(ctx, itemID)
	if err != nil {
		return nil, err
	} else if item == nil {
		return nil, errors.NotFound("", "测试项目不存在")
	}
	if item.Fields, err = a.ItemDAL.QueryFields(ctx, itemID); err != nil {
		return nil, err
	}
	if item.Metrics, err = a.ItemDAL.QueryMetrics(ctx, itemID); err != nil {
		return nil, err
	}
	if item.Rule, err = a.ItemDAL.GetRule(ctx, itemID); err != nil {
		return nil, err
	}
	if item.Advices, err = a.ItemDAL.QueryAdvices(ctx, itemID); err != nil {
		return nil, err
	}
	return item, nil
}

// Create 单条录入：服务端求值 fields → metrics → score，并落快照
func (a *Record) Create(ctx context.Context, formItem *schema.RecordForm, auditorID string) (*schema.Record, error) {
	subject, err := a.SubjectDAL.Get(ctx, formItem.SubjectID)
	if err != nil {
		return nil, err
	} else if subject == nil {
		return nil, errors.NotFound("", "受试者不存在")
	}
	item, err := a.loadItemWithConfig(ctx, formItem.ItemID)
	if err != nil {
		return nil, err
	}
	fields, metrics, fieldsJSON, err := a.evalFields(item, subject, formItem.Fields, formItem.TestedAt)
	if err != nil {
		return nil, err
	}
	standards, err := a.StandardDAL.Query(ctx, schema.StandardQueryParam{ItemID: item.ID})
	if err != nil {
		return nil, err
	}
	score, grade, _ := a.EvalEngine.EvalScore(item.Rule, subject, fields, metrics, standards)
	advices := MatchAdvices(item.Advices, fields, metrics, score)

	testedAt := time.Now()
	if formItem.TestedAt != nil {
		testedAt = *formItem.TestedAt
	}
	record := &schema.Record{
		ID:        util.NewXID(),
		SubjectID: formItem.SubjectID,
		ItemID:    formItem.ItemID,
		TestedAt:  testedAt,
		Fields:    fieldsJSON,
		Metrics:   mustJSON(metrics),
		Advices:   mustJSON(advices),
		Score:     score,
		Grade:     grade,
		AuditorID: auditorID,
		Remark:    formItem.Remark,
		CreatedAt: time.Now(),
	}
	if err := a.RecordDAL.Create(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

func (a *Record) Update(ctx context.Context, id string, formItem *schema.RecordForm) error {
	record, err := a.RecordDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if record == nil {
		return errors.NotFound("", "记录不存在")
	}

	subject, err := a.SubjectDAL.Get(ctx, formItem.SubjectID)
	if err != nil {
		return err
	} else if subject == nil {
		return errors.NotFound("", "受试者不存在")
	}
	item, err := a.loadItemWithConfig(ctx, formItem.ItemID)
	if err != nil {
		return err
	}
	fields, metrics, fieldsJSON, err := a.evalFields(item, subject, formItem.Fields, formItem.TestedAt)
	if err != nil {
		return err
	}
	standards, err := a.StandardDAL.Query(ctx, schema.StandardQueryParam{ItemID: item.ID})
	if err != nil {
		return err
	}
	score, grade, _ := a.EvalEngine.EvalScore(item.Rule, subject, fields, metrics, standards)

	record.SubjectID = formItem.SubjectID
	record.ItemID = formItem.ItemID
	if formItem.TestedAt != nil {
		record.TestedAt = *formItem.TestedAt
	}
	record.Fields = fieldsJSON
	record.Metrics = mustJSON(metrics)
	record.Advices = mustJSON(MatchAdvices(item.Advices, fields, metrics, score))
	record.Score = score
	record.Grade = grade
	record.Remark = formItem.Remark
	record.UpdatedAt = time.Now()
	return a.RecordDAL.Update(ctx, record)
}

// evalFields 求值并序列化 fields
func (a *Record) evalFields(item *schema.Item, subject *schema.Subject, input map[string]interface{}, testedAt *time.Time) (map[string]interface{}, map[string]interface{}, schema.JSONB, error) {
	at := time.Now()
	if testedAt != nil {
		at = *testedAt
	}
	// 只保留项目定义过的字段
	fields := make(map[string]interface{})
	for _, f := range item.Fields {
		if v, ok := input[f.Key]; ok && v != nil {
			fields[f.Key] = v
		} else if f.DefaultValue != nil {
			fields[f.Key] = *f.DefaultValue
		}
	}
	env := a.EvalEngine.BuildEnv(subject, at, fields, nil)
	metrics, _ := a.EvalEngine.EvalMetrics(item.Metrics, env)

	fieldsJSON, err := mustJSON(fields).MarshalJSON()
	if err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}
	sb := schema.JSONB(fieldsJSON)
	return fields, metrics, sb, nil
}

func (a *Record) Delete(ctx context.Context, id string) error {
	exists, err := a.RecordDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if exists == nil {
		return errors.NotFound("", "记录不存在")
	}
	return a.RecordDAL.Delete(ctx, id)
}

// ---------------- Report ----------------

type Report struct {
	Trans      *util.Trans
	ReportDAL  *dal.Report
	RecordDAL  *dal.Record
	SubjectDAL *dal.Subject
	ItemDAL    *dal.Item
}

// Preview 聚合勾选的记录 → 雷达图数据
func (a *Report) Preview(ctx context.Context, params *schema.ReportPreviewParams) ([]map[string]interface{}, error) {
	subject, err := a.SubjectDAL.Get(ctx, params.SubjectID)
	if err != nil {
		return nil, err
	} else if subject == nil {
		return nil, errors.NotFound("", "受试者不存在")
	}
	records, err := a.RecordDAL.QueryAll(ctx, params.SubjectID, params.RecordIDs)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.BadRequest("", "未找到所选测试记录")
	}

	items := make(map[string]*schema.Item)
	for _, r := range records {
		if _, ok := items[r.ItemID]; !ok {
			item, err := a.ItemDAL.Get(ctx, r.ItemID)
			if err != nil {
				return nil, err
			} else if item != nil {
				if item.Metrics, err = a.ItemDAL.QueryMetrics(ctx, r.ItemID); err != nil {
					return nil, err
				}
				if item.Advices, err = a.ItemDAL.QueryAdvices(ctx, r.ItemID); err != nil {
					return nil, err
				}
			}
			items[r.ItemID] = item
		}
	}

	data := make([]map[string]interface{}, 0, len(records))
	for _, r := range records {
		metrics := map[string]interface{}{}
		if r.Metrics != nil {
			metrics, _ = r.Metrics.Map()
		}
		fields := map[string]interface{}{}
		if r.Fields != nil {
			fields, _ = r.Fields.Map()
		}
		advices := parseAdvices(r.Advices)
		// 历史记录可能没有建议快照，按项目当前配置补算
		if len(advices) == 0 && items[r.ItemID] != nil {
			advices = MatchAdvices(items[r.ItemID].Advices, fields, metrics, r.Score)
		}
		row := map[string]interface{}{
			"record_id": r.ID,
			"item_id":   r.ItemID,
			"item_name": "",
			"score":     r.Score,
			"grade":     r.Grade,
			"tested_at": r.TestedAt,
			"fields":    r.Fields,
			"metrics":   filterReportMetrics(items[r.ItemID], metrics),
			"advices":   advices,
		}
		if items[r.ItemID] != nil {
			row["item_name"] = items[r.ItemID].Name
		}
		data = append(data, row)
	}
	return data, nil
}

// parseAdvices 解析记录里的建议快照
func parseAdvices(j schema.JSONB) []string {
	if len(j) == 0 {
		return nil
	}
	var out []string
	if err := json.Unmarshal(j, &out); err != nil {
		return nil
	}
	return out
}

func filterReportMetrics(item *schema.Item, metrics map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	if item == nil {
		return out
	}
	for _, m := range item.Metrics {
		if m.InReport {
			out[m.Key] = metrics[m.Key]
		}
	}
	return out
}

func (a *Report) Query(ctx context.Context, params schema.ReportQueryParam) (*schema.ReportQueryResult, error) {
	return a.ReportDAL.Query(ctx, params)
}

func (a *Report) Get(ctx context.Context, id string) (*schema.Report, error) {
	report, err := a.ReportDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if report == nil {
		return nil, errors.NotFound("", "报告不存在")
	}
	return report, nil
}

func (a *Report) Create(ctx context.Context, formItem *schema.ReportForm, creatorID string) (*schema.Report, error) {
	data, err := a.Preview(ctx, &schema.ReportPreviewParams{SubjectID: formItem.SubjectID, RecordIDs: formItem.RecordIDs})
	if err != nil {
		return nil, err
	}
	recordIDs, err := mustJSON(formItem.RecordIDs).MarshalJSON()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	dataJSON, err := mustJSON(data).MarshalJSON()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	report := &schema.Report{
		ID:         util.NewXID(),
		SubjectID:  formItem.SubjectID,
		Title:      formItem.Title,
		RecordIDs:  schema.JSONB(recordIDs),
		Data:       schema.JSONB(dataJSON),
		Conclusion: formItem.Conclusion,
		CreatorID:  creatorID,
		CreatedAt:  time.Now(),
	}
	if err := a.ReportDAL.Create(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (a *Report) Update(ctx context.Context, id string, conclusion string) error {
	report, err := a.ReportDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if report == nil {
		return errors.NotFound("", "报告不存在")
	}
	report.Conclusion = conclusion
	report.UpdatedAt = time.Now()
	return a.ReportDAL.Update(ctx, report)
}

func (a *Report) Delete(ctx context.Context, id string) error {
	report, err := a.ReportDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if report == nil {
		return errors.NotFound("", "报告不存在")
	}
	return a.ReportDAL.Delete(ctx, id)
}

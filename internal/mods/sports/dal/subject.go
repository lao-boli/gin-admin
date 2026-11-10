package dal

import (
	"context"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

func getSubjectDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Subject))
}

type Subject struct {
	DB *gorm.DB
}

func (a *Subject) Query(ctx context.Context, params schema.SubjectQueryParam) (*schema.SubjectQueryResult, error) {
	params.Pagination = true

	db := getSubjectDB(ctx, a.DB)
	if v := params.Keyword; v != "" {
		kw := "%" + v + "%"
		db = db.Where("name LIKE ? OR code LIKE ?", kw, kw)
	}
	if v := params.AudienceType; v != "" {
		db = db.Where("audience_type = ?", v)
	}
	if v := params.Status; v != "" {
		db = db.Where("status = ?", v)
	}

	var list schema.Subjects
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, util.QueryOptions{}, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &schema.SubjectQueryResult{Data: list, PageResult: pageResult}, nil
}

func (a *Subject) Get(ctx context.Context, id string) (*schema.Subject, error) {
	item := new(schema.Subject)
	ok, err := util.FindOne(ctx, getSubjectDB(ctx, a.DB).Where("id=?", id), util.QueryOptions{}, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

func (a *Subject) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, getSubjectDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

func (a *Subject) ExistsCode(ctx context.Context, code string) (bool, error) {
	ok, err := util.Exists(ctx, getSubjectDB(ctx, a.DB).Where("code=?", code))
	return ok, errors.WithStack(err)
}

func (a *Subject) MaxCode(ctx context.Context) (string, error) {
	var code string
	err := getSubjectDB(ctx, a.DB).Select("code").Order("created_at DESC").Limit(1).Scan(&code).Error
	return code, errors.WithStack(err)
}

func (a *Subject) Create(ctx context.Context, item *schema.Subject) error {
	return errors.WithStack(getSubjectDB(ctx, a.DB).Create(item).Error)
}

func (a *Subject) Update(ctx context.Context, item *schema.Subject) error {
	result := getSubjectDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

func (a *Subject) Delete(ctx context.Context, id string) error {
	return errors.WithStack(getSubjectDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.Subject)).Error)
}

// ---------------- Record ----------------

func getRecordDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Record))
}

type Record struct {
	DB *gorm.DB
}

func (a *Record) Query(ctx context.Context, params schema.RecordQueryParam) (*schema.RecordQueryResult, error) {
	params.Pagination = true

	db := getRecordDB(ctx, a.DB)
	if v := params.SubjectID; v != "" {
		db = db.Where("subject_id = ?", v)
	}
	if v := params.ItemID; v != "" {
		db = db.Where("item_id = ?", v)
	}
	if v := params.From; v != nil {
		db = db.Where("tested_at >= ?", v)
	}
	if v := params.To; v != nil {
		db = db.Where("tested_at <= ?", v)
	}

	var list schema.Records
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam,
		util.QueryOptions{OrderFields: []util.OrderByParam{{Field: "tested_at", Direction: util.DESC}}}, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &schema.RecordQueryResult{Data: list, PageResult: pageResult}, nil
}

func (a *Record) QueryAll(ctx context.Context, subjectID string, recordIDs []string) (schema.Records, error) {
	db := getRecordDB(ctx, a.DB).Where("subject_id = ?", subjectID)
	if len(recordIDs) > 0 {
		db = db.Where("id IN (?)", recordIDs)
	}
	var list schema.Records
	err := db.Order("tested_at ASC").Find(&list).Error
	return list, errors.WithStack(err)
}

func (a *Record) Get(ctx context.Context, id string) (*schema.Record, error) {
	item := new(schema.Record)
	ok, err := util.FindOne(ctx, getRecordDB(ctx, a.DB).Where("id=?", id), util.QueryOptions{}, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

func (a *Record) Create(ctx context.Context, item *schema.Record) error {
	return errors.WithStack(getRecordDB(ctx, a.DB).Create(item).Error)
}

func (a *Record) Update(ctx context.Context, item *schema.Record) error {
	result := getRecordDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

func (a *Record) Delete(ctx context.Context, id string) error {
	return errors.WithStack(getRecordDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.Record)).Error)
}

// ---------------- Report ----------------

func getReportDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Report))
}

type Report struct {
	DB *gorm.DB
}

func (a *Report) Query(ctx context.Context, params schema.ReportQueryParam) (*schema.ReportQueryResult, error) {
	params.Pagination = true
	var list schema.Reports
	pageResult, err := util.WrapPageQuery(ctx, getReportDB(ctx, a.DB).Where("subject_id = ?", params.SubjectID),
		params.PaginationParam, util.QueryOptions{}, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &schema.ReportQueryResult{Data: list, PageResult: pageResult}, nil
}

func (a *Report) Get(ctx context.Context, id string) (*schema.Report, error) {
	item := new(schema.Report)
	ok, err := util.FindOne(ctx, getReportDB(ctx, a.DB).Where("id=?", id), util.QueryOptions{}, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

func (a *Report) Create(ctx context.Context, item *schema.Report) error {
	return errors.WithStack(getReportDB(ctx, a.DB).Create(item).Error)
}

func (a *Report) Update(ctx context.Context, item *schema.Report) error {
	result := getReportDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

func (a *Report) Delete(ctx context.Context, id string) error {
	return errors.WithStack(getReportDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.Report)).Error)
}

package schema

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// JSONB 通用 jsonb 值对象（map / array 均可）
type JSONB json.RawMessage

func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *JSONB) UnmarshalJSON(b []byte) error {
	*j = append((*j)[0:0], b...)
	return nil
}

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

func (j *JSONB) Scan(v interface{}) error {
	switch data := v.(type) {
	case nil:
		*j = nil
		return nil
	case []byte:
		*j = append((*j)[0:0], data...)
		return nil
	case string:
		*j = append((*j)[0:0], data...)
		return nil
	default:
		return errors.New("unsupported jsonb source type")
	}
}

func (j JSONB) Map() (map[string]interface{}, error) {
	if len(j) == 0 {
		return map[string]interface{}{}, nil
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(bytes.TrimSpace(j), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// ---------------- 受试者 ----------------

const (
	SubjectStatusEnabled  = "enabled"
	SubjectStatusDisabled = "disabled"
)

var subjectAudienceDict = map[string]string{
	"preschool": "幼儿",
	"primary":   "小学",
	"middle":    "中学",
	"college":   "大学",
	"elderly":   "老年人",
}

func SubjectAudienceLabel(key string) string {
	if v, ok := subjectAudienceDict[key]; ok {
		return v
	}
	return key
}

type Subject struct {
	ID           string     `json:"id" gorm:"size:20;primarykey;"`
	Code         string     `json:"code" gorm:"size:32;uniqueIndex;"`
	Name         string     `json:"name" gorm:"size:64;index;"`
	Gender       int        `json:"gender"` // 1男 2女
	BirthDate    *time.Time `json:"birth_date"`
	Height       float64    `json:"height" gorm:"type:numeric(6,2);default:0"` // 身高(cm)，录入时自动带入
	Weight       float64    `json:"weight" gorm:"type:numeric(6,2);default:0"` // 体重(kg)，录入时自动带入
	AudienceType string     `json:"audience_type" gorm:"size:32;index;"`
	Phone        string     `json:"phone" gorm:"size:20;"`
	Remark       string     `json:"remark" gorm:"size:255;"`
	Status       string     `json:"status" gorm:"size:20;index;"`
	CreatedAt    time.Time  `json:"created_at" gorm:"index;"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"index;"`
}

func (a *Subject) TableName() string {
	return config.C.FormatTableName("sports_subject")
}

type Subjects []*Subject

type SubjectQueryParam struct {
	util.PaginationParam
	LikeName     string `form:"name"`
	Keyword      string `form:"keyword"` // 姓名或编号
	AudienceType string `form:"audienceType"`
	Status       string `form:"status" binding:"oneof=disabled enabled ''"`
}

type SubjectQueryResult struct {
	Data       Subjects
	PageResult *util.PaginationResult
}

type SubjectForm struct {
	Name         string     `json:"name" binding:"required,max=64"`
	Gender       int        `json:"gender" binding:"required,oneof=1 2"`
	BirthDate    *time.Time `json:"birth_date"`
	Height       float64    `json:"height"`
	Weight       float64    `json:"weight"`
	AudienceType string     `json:"audience_type" binding:"omitempty,max=32"`
	Phone        string     `json:"phone" binding:"max=20"`
	Remark       string     `json:"remark" binding:"max=255"`
	Status       string     `json:"status" binding:"required,oneof=disabled enabled"`
}

func (a *SubjectForm) FillTo(subject *Subject) error {
	subject.Name = a.Name
	subject.Gender = a.Gender
	subject.BirthDate = a.BirthDate
	subject.Height = a.Height
	subject.Weight = a.Weight
	subject.AudienceType = a.AudienceType
	subject.Phone = a.Phone
	subject.Remark = a.Remark
	subject.Status = a.Status
	return nil
}

// ---------------- 测试记录 ----------------

type Record struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"`
	SubjectID string    `json:"subject_id" gorm:"size:20;index;"`
	ItemID    string    `json:"item_id" gorm:"size:20;index;"`
	TestedAt  time.Time `json:"tested_at" gorm:"index;"`
	Fields    JSONB     `json:"fields" gorm:"type:jsonb;"`
	Metrics   JSONB     `json:"metrics" gorm:"type:jsonb;"`
	Advices   JSONB     `json:"advices" gorm:"type:jsonb;"` // 命中的改进建议快照（[]string）
	Score     *int      `json:"score"`
	Grade     string    `json:"grade" gorm:"size:8;"`
	AuditorID string    `json:"auditor_id" gorm:"size:20;index;"`
	Remark    string    `json:"remark" gorm:"size:255;"`
	CreatedAt time.Time `json:"created_at" gorm:"index;"`
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`
}

func (a *Record) TableName() string {
	return config.C.FormatTableName("sports_record")
}

type Records []*Record

type RecordQueryParam struct {
	util.PaginationParam
	SubjectID string     `form:"subjectId" binding:"required"`
	ItemID    string     `form:"itemId"`
	From      *time.Time `form:"from"`
	To        *time.Time `form:"to"`
}

type RecordQueryResult struct {
	Data       Records
	PageResult *util.PaginationResult
}

type RecordForm struct {
	SubjectID string                 `json:"subject_id" binding:"required,max=20"`
	ItemID    string                 `json:"item_id" binding:"required,max=20"`
	TestedAt  *time.Time             `json:"tested_at"`
	Fields    map[string]interface{} `json:"fields" binding:"required"`
	Remark    string                 `json:"remark" binding:"max=255"`
}

// RecordEvalResult 试算结果（录入页实时预览）
type RecordEvalResult struct {
	Metrics  map[string]interface{} `json:"metrics"`
	Score    *int                   `json:"score"`
	Grade    string                 `json:"grade"`
	Advices  []string               `json:"advices,omitempty"`
	Warnings []string               `json:"warnings,omitempty"`
}

// ---------------- 评估报告 ----------------

type Report struct {
	ID         string    `json:"id" gorm:"size:20;primarykey;"`
	SubjectID  string    `json:"subject_id" gorm:"size:20;index;"`
	Title      string    `json:"title" gorm:"size:128;"`
	RecordIDs  JSONB     `json:"record_ids" gorm:"type:jsonb;"`
	Data       JSONB     `json:"data" gorm:"type:jsonb;"`
	Conclusion string    `json:"conclusion" gorm:"type:text;"`
	CreatorID  string    `json:"creator_id" gorm:"size:20;"`
	CreatedAt  time.Time `json:"created_at" gorm:"index;"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"index;"`
}

func (a *Report) TableName() string {
	return config.C.FormatTableName("sports_report")
}

type Reports []*Report

type ReportQueryParam struct {
	util.PaginationParam
	SubjectID string `form:"subjectId" binding:"required"`
}

type ReportQueryResult struct {
	Data       Reports
	PageResult *util.PaginationResult
}

type ReportForm struct {
	SubjectID  string   `json:"subject_id" binding:"required,max=20"`
	Title      string   `json:"title" binding:"required,max=128"`
	RecordIDs  []string `json:"record_ids" binding:"required,min=1"`
	Conclusion string   `json:"conclusion"`
}

type ReportPreviewParams struct {
	SubjectID string   `json:"subject_id" binding:"required,max=20"`
	RecordIDs []string `json:"record_ids" binding:"required,min=1"`
}

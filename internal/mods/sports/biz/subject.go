package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// Subject 受试者管理
type Subject struct {
	Trans       *util.Trans
	SubjectDAL  *dal.Subject
	RecordDAL   *dal.Record
}

func (a *Subject) Query(ctx context.Context, params schema.SubjectQueryParam) (*schema.SubjectQueryResult, error) {
	return a.SubjectDAL.Query(ctx, params)
}

func (a *Subject) Get(ctx context.Context, id string) (*schema.Subject, error) {
	subject, err := a.SubjectDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if subject == nil {
		return nil, errors.NotFound("", "受试者不存在")
	}
	return subject, nil
}

func (a *Subject) Create(ctx context.Context, formItem *schema.SubjectForm) (*schema.Subject, error) {
	subject := &schema.Subject{
		ID:        util.NewXID(),
		CreatedAt: time.Now(),
	}
	if err := formItem.FillTo(subject); err != nil {
		return nil, err
	}

	// 编号自动生成：S + 年月 + 4位序号
	maxCode, err := a.SubjectDAL.MaxCode(ctx)
	if err != nil {
		return nil, err
	}
	seq := 1
	prefix := fmt.Sprintf("S%s", time.Now().Format("0601"))
	if len(maxCode) == len(prefix)+4 && maxCode[:len(prefix)] == prefix {
		fmt.Sscanf(maxCode[len(prefix):], "%d", &seq)
		seq++
	}
	subject.Code = fmt.Sprintf("%s%04d", prefix, seq)
	if subject.AudienceType == "" {
		subject.AudienceType = "preschool"
	}

	if err := a.SubjectDAL.Create(ctx, subject); err != nil {
		return nil, err
	}
	return subject, nil
}

func (a *Subject) Update(ctx context.Context, id string, formItem *schema.SubjectForm) error {
	subject, err := a.SubjectDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if subject == nil {
		return errors.NotFound("", "受试者不存在")
	}
	if err := formItem.FillTo(subject); err != nil {
		return err
	}
	subject.UpdatedAt = time.Now()
	return a.SubjectDAL.Update(ctx, subject)
}

func (a *Subject) Delete(ctx context.Context, id string) error {
	exists, err := a.SubjectDAL.Exists(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "受试者不存在")
	}
	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.SubjectDAL.Delete(ctx, id); err != nil {
			return err
		}
		return nil
	})
}

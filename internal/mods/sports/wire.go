package sports

import (
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/api"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/biz"
	"github.com/LyricTian/gin-admin/v10/internal/mods/sports/dal"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/google/wire"
)

// Collection of wire providers
var Set = wire.NewSet(
	wire.Struct(new(Sports), "*"),
	wire.Struct(new(dal.Category), "*"),
	wire.Struct(new(dal.Item), "*"),
	wire.Struct(new(dal.Standard), "*"),
	wire.Struct(new(dal.Subject), "*"),
	wire.Struct(new(dal.Record), "*"),
	wire.Struct(new(dal.Report), "*"),
	wire.Struct(new(biz.Category), "*"),
	wire.Struct(new(biz.Item), "*"),
	wire.Struct(new(biz.Standard), "*"),
	wire.Struct(new(biz.Subject), "*"),
	wire.Struct(new(biz.Record), "*"),
	wire.Struct(new(biz.Report), "*"),
	wire.Struct(new(biz.Seed), "*"),
	wire.Struct(new(biz.EvalEngine), "*"),
	wire.Struct(new(api.Category), "*"),
	wire.Struct(new(api.Item), "*"),
	wire.Struct(new(api.Standard), "*"),
	wire.Struct(new(api.Subject), "*"),
	wire.Struct(new(api.Record), "*"),
	wire.Struct(new(api.Report), "*"),
)

var _ = util.NewXID

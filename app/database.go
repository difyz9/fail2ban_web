package app

import (
	"context"
	"fail2ban-web/internal/model"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseParams 数据库依赖参数
type DatabaseParams struct {
	fx.In
	Logger *zap.Logger
}

// NewDatabase 创建数据库连接
func NewDatabase(lc fx.Lifecycle, params DatabaseParams) (*gorm.DB, error) {
	// 配置 GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)

	// 打开数据库连接
	db, err := gorm.Open(sqlite.Open("fail2ban_web.db"), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	params.Logger.Info("Database connection established")

	// 添加生命周期钩子
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Logger.Info("Running database migrations...")
			// 自动迁移
			if err := db.AutoMigrate(
				&model.BannedIP{},
				&model.Fail2banJail{},
			); err != nil {
				return err
			}
			params.Logger.Info("Database migrations completed successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("Closing database connection...")
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})

	return db, nil
}

// DatabaseModule 数据库模块
var DatabaseModule = fx.Module("database",
	fx.Provide(NewDatabase),
)

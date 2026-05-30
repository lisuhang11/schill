package db

import (
	"os"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func AutoMigrateIfEnabled(db *gorm.DB, enabled bool, models ...interface{}) error {
	if db == nil {
		return nil
	}

	if !enabled {
		logx.Infof("skip automigrate because AutoMigrate is disabled")
		return nil
	}

	if isProductionEnv() {
		logx.Infof("skip automigrate in production environment")
		return nil
	}

	return db.AutoMigrate(models...)
}

func isProductionEnv() bool {
	for _, key := range []string{"APP_ENV", "GO_ENV", "ENV"} {
		value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
		if value == "prod" || value == "production" {
			return true
		}
	}
	return false
}

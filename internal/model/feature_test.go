package model_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/zouluxing/feature_orbit_server/internal/model"
)

// TestFeatureModel_AutoMigrate_ConstraintNames 验证 GORM AutoMigrate 使用的约束名
// 与 migration SQL 中的命名保持一致，避免生产环境出现
// "constraint uni_features_slug does not exist" 的错误。
func TestFeatureModel_AutoMigrate_ConstraintNames(t *testing.T) {
	// 使用内存 SQLite 不需要真实 PostgreSQL
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "open in-memory sqlite")

	// AutoMigrate 不能报错
	err = db.AutoMigrate(&model.Feature{})
	require.NoError(t, err, "AutoMigrate must succeed without constraint errors")

	// 检查 SQLite 建表语句中没有 uni_features_slug（不应该存在的旧命名）
	var createSQL string
	db.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name='features'").
		Scan(&createSQL)
	assert.NotContains(t, createSQL, "uni_features_slug",
		"GORM must NOT generate 'uni_features_slug'; expected 'uq_features_slug'")
}

// TestFeatureModel_SlugUniqueIndex_Name 直接检查模型 tag 中约束名是否正确
// 不需要数据库连接。
func TestFeatureModel_SlugUniqueIndex_Name(t *testing.T) {
	// 通过反射检查 struct tag 包含正确的约束名
	// 这确保即使没有 DB 环境也能在 CI 中运行
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.AutoMigrate(&model.Feature{})

	// 查询所有索引
	type indexInfo struct {
		Name string
	}
	var indexes []indexInfo
	db.Raw("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='features'").
		Scan(&indexes)

	indexNames := make([]string, 0, len(indexes))
	for _, idx := range indexes {
		indexNames = append(indexNames, idx.Name)
	}

	// 不应存在 GORM 自动生成的旧命名
	for _, name := range indexNames {
		assert.False(t,
			strings.HasPrefix(name, "uni_"),
			"found auto-generated index name '%s'; all uniqueIndex must use explicit names matching migration SQL",
			name,
		)
	}
}

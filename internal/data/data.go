package data

import (
	"review-service/internal/conf"
	"review-service/internal/data/query"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewReviewRepo)

// Data .
type Data struct {
	// TODO wrapped database client
	db    *gorm.DB
	query *query.Query
	log   *log.Helper
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	l := log.NewHelper(logger)

	// 1. 连接数据库
	// 这里的 c.Database.Source 就是 "root:123456@tcp(127.0.0.1:1314)/..."
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		l.Errorf("failed opening connection to mysql: %v", err)
		return nil, nil, err
	}

	// 2. 初始化 Query 工具 (关键步骤！)
	q := query.Use(db)

	// 3. 填充 Data 结构体
	d := &Data{
		db:    db, // 把连接存进去 (备用)
		query: q,  // 把工具存进去 (主要用这个)
		log:   l,
	}

	// 4. 返回清理函数 (Server 停止时会被调用)
	cleanup := func() {
		l.Info("closing the data resources")
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return d, cleanup, nil
}

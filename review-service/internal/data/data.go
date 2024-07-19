package data

import (
	"errors"
	"fmt"
	"review-service/internal/conf"
	"review-service/internal/data/query"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewReviewRepo, NewDB, NewES, NewRdbClient)

// Data .
type Data struct {
	// TODO wrapped database client
	query *query.Query
	es    *elasticsearch.TypedClient
	rdb   *redis.Client
}

// NewData .
func NewData(db *gorm.DB, es *elasticsearch.TypedClient, rdb *redis.Client, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	// 设置全局数据连接
	query.SetDefault(db)
	return &Data{
		query: query.Q,
		es:    es,
		rdb:   rdb,
	}, cleanup, nil
}

func NewRdbClient(cfg *conf.Data) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		WriteTimeout: cfg.Redis.ReadTimeout.AsDuration(),
		ReadTimeout:  cfg.Redis.WriteTimeout.AsDuration(),
	})
}

func NewES(cfg *conf.ES) (*elasticsearch.TypedClient, error) {
	c, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: cfg.GetAddress(),
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}
func NewDB(c *conf.Data) (*gorm.DB, error) {
	// 从配置文件中获取数据库连接
	driver := c.Database.Driver
	dsn := c.Database.Source
	switch strings.ToLower(driver) {
	case "mysql":
		db, err := gorm.Open(mysql.Open(dsn))
		if err != nil {
			panic(fmt.Errorf("connect db failed,%v", err))
		}
		return db, nil
	case "oracal":
		// 略
		return nil, errors.New("暂时不支持的数据库类型:oracal")
	case "sqlite":
		// 略
		return nil, errors.New("暂时不支持的数据库类型:sqlite")
	default:
		panic("不支持的数据库类型")
	}
}

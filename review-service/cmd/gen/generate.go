package main

import (
	"flag"
	"fmt"
	"review-service/internal/conf"
	"strings"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// GORM-GEN生成代码配置
var flagconf string

/* init函数，这个函数在文件中所有函数之前执行 */
func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path,eg:-conf config.yaml 配置文件路径")
}

func main() {
	flag.Parse()
	// 指定配置文件的位置
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()
	//从指定位置加载配置
	if err := c.Load(); err != nil {
		panic(err)
	}
	// 将配置赋值到结构体中
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	g := gen.NewGenerator(gen.Config{
		OutPath:       "../../internal/data/query",                   //指定的MODEL文件输出位置
		Mode:          gen.WithDefaultQuery | gen.WithQueryInterface, //生成MODEL文件的方式
		FieldNullable: true,                                          //数据库表中允许为空的字段，在MODEL中将会是指针类型
		// ModelPkgPath: "dal/model",                                   //默认与OUTPATH相同
	})

	connectDB := func() *gorm.DB {
		driver := bc.Data.Database.Driver
		dsn := bc.Data.Database.Source
		switch strings.ToLower(driver) {
		case "mysql":
			db, err := gorm.Open(mysql.Open(dsn))
			if err != nil {
				panic(fmt.Errorf("connect db failed,%v", err))
			}
			return db
		case "oracal":
			// 略
			return nil
		case "sqlite":
			// 略
			return nil
		default:
			panic("不支持的数据库类型")
		}
	}()

	g.UseDB(connectDB)
	g.ApplyBasic(g.GenerateAllTable()...) //该数据库中所有表格都要求生成MODEL文件
	g.Execute()                           //开始生成
}

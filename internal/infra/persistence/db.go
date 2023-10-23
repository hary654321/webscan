package persistence

import (
	"context"
	"github.com/elastic/go-elasticsearch/v7"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/driver/mysql"
	//"gorm.io/driver/sqlite"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"webscan/config"
)

func InitElastic(configureInterface config.ConfigureInterface) (*elasticsearch.Client, error) {
	// 创建Elasticsearch客户端
	dbc := configureInterface.GetDatabaseConfig()
	cfg := elasticsearch.Config{
		Addresses: []string{dbc.Elasticsearch}, // Elasticsearch服务器地址
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	// 发送Ping请求
	res, err := client.Ping(client.Ping.WithContext(context.Background()))
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	return client, err
}

// InitDatabase 初始化数据库连接
func InitDatabase(configureInterface config.ConfigureInterface) (*gorm.DB, *gorm.DB) {
	app := configureInterface.GetAppConfig()
	conf := configureInterface.GetDatabaseConfig()

	// init sqlite connection
	mySqlite, err := gorm.Open(sqlite.Open("webscan.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	mySqlite.Logger.LogMode(logger.Info)

	// init mysql connection
	var myMySQL *gorm.DB
	if app.Role == "master" {
		myMySQL, err = gorm.Open(mysql.Open(conf.Mysql), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		myMySQL.Logger.LogMode(logger.Info)
	}
	return myMySQL, mySqlite
}

//func (d *DatabaseImpl) AutoMigrate() {
//	if d.sqlite != nil {
//		d.sqlite.AutoMigrate(&entity.Task{})
//	}
//}

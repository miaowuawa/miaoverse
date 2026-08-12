package ConfigService

import (
	"context"
	"fmt"
	validator2 "github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	fiberstoreredis "github.com/gofiber/storage/redis/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"miaoverse/dao"
	"miaoverse/dao/article"
	"miaoverse/dao/content"
	"miaoverse/dao/interacts"
	"miaoverse/dao/user"
	"miaoverse/model/server"
	"miaoverse/model/server/conf"
	"miaoverse/service/DBMigration"
	"miaoverse/service/UserBlock"
	storagemongo "miaoverse/service/mongo"
	storages3 "miaoverse/service/s3"
	"miaoverse/service/security/sms/codemanager"
	"miaoverse/service/security/sms/smsbao"
	"miaoverse/util"
	"strconv"
	"time"
)

func ConfToServants(conf *conf.AppConfig) (*server.Servants, error) {
	//step1 init db
	//mysql
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Sql.Username,
		conf.Sql.Password,
		conf.Sql.Host,
		conf.Sql.Port,
		conf.Sql.DbName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error), // 设置日志级别
	})
	if err != nil {
		return nil, fmt.Errorf("GORM连接数据库失败：%v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("GORM连接数据库失败：%v", err)
	}
	// 设置连接池参数
	sqlDB.SetMaxOpenConns(conf.Sql.MaxOpenConns) // 最大打开连接数
	sqlDB.SetMaxIdleConns(conf.Sql.MaxIdleConns) // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Duration(conf.Sql.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(conf.Sql.ConnMaxIdletime) * time.Minute) // 连接最大空闲时间
	// 自动迁移：把数据库结构迁移到最新版本，迁移失败直接阻止启动
	// 迁移使用独立连接池（DSN 额外开启 multiStatements），主业务连接池不开启该选项，避免扩大注入面
	migrationDSN := dsn + "&multiStatements=true"
	migrationDB, err := gorm.Open(mysql.Open(migrationDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return nil, fmt.Errorf("打开数据库迁移连接失败：%v", err)
	}
	migrationSQLDB, err := migrationDB.DB()
	if err != nil {
		return nil, fmt.Errorf("打开数据库迁移连接失败：%v", err)
	}
	defer migrationSQLDB.Close()
	if err := DBMigration.Migrate(migrationDB); err != nil {
		return nil, err
	}
	//redis
	SessionStorage := fiberstoreredis.New(fiberstoreredis.Config{
		Host:     conf.Redis.Host,
		Port:     conf.Redis.Port,
		Username: conf.Redis.Username,
		Password: conf.Redis.Password,
		Database: conf.Session.DB,
	})
	smsRedisClient := redis.NewClient(&redis.Options{
		Password: conf.Redis.Password,
		Addr:     conf.Redis.Host + ":" + strconv.Itoa(conf.Redis.Port),
		DB:       conf.SmsBao.DB,
	})
	smsRedisCtx := context.Background()
	sessionCtx := context.Background()
	SessionStorage.Conn().Ping(sessionCtx)
	status := smsRedisClient.Ping(smsRedisCtx)
	if status.Err() != nil {
		return nil, status.Err()
	}
	//step2 init services
	//init smsbao
	smsServant := &smsbao.SmsBaoServant{
		Password: conf.SmsBao.Passwd,
		Username: conf.SmsBao.Username,
		Head:     conf.SmsBao.Head,
		Gateway:  conf.SmsBao.Gateway,
	}

	//init code manager
	codeManager := &codemanager.CodeManager{
		Redis:          smsRedisClient,
		CodeExpireTime: time.Duration(conf.SMS.RateLimit) * time.Minute,
		Context:        smsRedisCtx,
	}

	//init user DAO

	userdao := &user.UserDAO{}
	userdao.DB = db

	//init content DAO
	contentdao := &content.ContentDAO{}
	contentdao.DB = db

	//init interacts DAO
	interactsdao := &interacts.InteractsDAO{}
	interactsdao.DB = db

	//init user block bitmap servant
	blockServant := UserBlock.NewServant(smsRedisClient, conf.Cache.DB)
	//step3 validator
	validator := validator2.New()
	err = util.Validate.InitialValidator(validator)
	if err != nil {
		return nil, err
	}

	var s3Servant *storages3.Servant
	if conf.S3.Enabled {
		s3Servant, err = storages3.NewServant(context.Background(), storages3.Config{
			Endpoint:              conf.S3.Endpoint,
			Region:                conf.S3.Region,
			AccessKeyID:           conf.S3.AccessKeyID,
			SecretAccessKey:       conf.S3.SecretAccessKey,
			SessionToken:          conf.S3.SessionToken,
			Bucket:                conf.S3.Bucket,
			PublicBaseURL:         conf.S3.PublicBaseURL,
			UsePathStyle:          conf.S3.UsePathStyle,
			TempSignatureSecret:   conf.S3.TempSignatureSecret,
			TempSignatureDuration: time.Duration(conf.S3.TempSignatureExpireSeconds) * time.Second,
			TempLinkDuration:      time.Duration(conf.S3.TempLinkExpireSeconds) * time.Second,
		})
		if err != nil {
			return nil, err
		}
	}

	var mongoServant *storagemongo.Servant
	var articleServant *article.ArticleDAO
	if conf.Mongo.Enabled {
		mongoServant, err = storagemongo.NewServant(context.Background(), storagemongo.Config{
			Host:                   conf.Mongo.Host,
			Port:                   conf.Mongo.Port,
			Username:               conf.Mongo.Username,
			Password:               conf.Mongo.Password,
			AuthSource:             conf.Mongo.AuthSource,
			Database:               conf.Mongo.DbName,
			MaxPoolSize:            conf.Mongo.MaxPoolSize,
			MinPoolSize:            conf.Mongo.MinPoolSize,
			ConnectTimeout:         time.Duration(conf.Mongo.ConnectTimeoutSeconds) * time.Second,
			ServerSelectionTimeout: time.Duration(conf.Mongo.ServerSelectionTimeoutSeconds) * time.Second,
		})
		if err != nil {
			return nil, err
		}
		// 文章域跨库：元数据在 MySQL，正文在 MongoDB
		articleServant = dao.NewArticleDao(db, mongoServant.Database())
	}

	//final:return
	return &server.Servants{
		FiberSessionStorage: SessionStorage,
		SmsServant:          smsServant,
		CodeManager:         codeManager,
		UserServant:         userdao,
		ContentServant:      contentdao,
		InteractsServant:    interactsdao,
		ArticleServant:      articleServant,
		BlockServant:        blockServant,
		Validator:           validator,
		S3Servant:           s3Servant,
		MongoServant:        mongoServant,
		MaxUploadFileSize:   conf.UploadMaxFileSizeBytes(),
	}, nil

}

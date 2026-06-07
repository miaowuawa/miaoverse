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
	"miaoverse/dao/user"
	"miaoverse/model/server"
	"miaoverse/model/server/conf"
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

	//final:return
	return &server.Servants{
		FiberSessionStorage: SessionStorage,
		SmsServant:          smsServant,
		CodeManager:         codeManager,
		UserServant:         userdao,
		Validator:           validator,
		S3Servant:           s3Servant,
		MaxUploadFileSize:   conf.UploadMaxFileSizeBytes(),
	}, nil

}

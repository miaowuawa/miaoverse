package ConfigService

import (
	"context"
	"miaoverse/model/server"
	"miaoverse/model/server/conf"
	"miaoverse/service/security/sms/codemanager"
	"miaoverse/service/security/sms/smsbao"
	"strconv"
	"time"

	storage "github.com/gofiber/storage/redis/v3"

	"github.com/redis/go-redis/v9"
)

func ConfToServants(conf *conf.AppConfig) (error, *server.Servants) {
	//step1 init db
	smsRedisClient := redis.NewClient(&redis.Options{
		Password: conf.Redis.Password,
		Addr:     conf.Redis.Host + ":" + strconv.Itoa(conf.Redis.Port),
		DB:       conf.SmsBao.DB,
	})

	// 1. 创建 UniversalClient（仅改这一行，参数和 NewClient 几乎一样）
	cacheRedisClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{conf.Redis.Host + ":" + strconv.Itoa(conf.Redis.Port)}, // 单机填一个地址即可
		Password: conf.Redis.Password,
		DB:       conf.SmsBao.DB,
	})

	smsRedisCtx := context.Background()
	sessionCacheRedisCtx := context.Background()

	sessionCache := storage.NewFromConnection(cacheRedisClient.(redis.UniversalClient))

	status := smsRedisClient.Ping(smsRedisCtx)
	if status.Err() != nil {
		panic(status.Err())
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

	//step3 return
	return nil, &server.Servants{
		SmsServant:  smsServant,
		CodeManager: codeManager,
	}
}

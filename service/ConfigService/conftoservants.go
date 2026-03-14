package ConfigService

import (
	"context"
	"github.com/go-redis/redis/v8"
	"miaoverse/model/server"
	"miaoverse/model/server/conf"
	"miaoverse/service/security/sms/codemanager"
	"miaoverse/service/security/sms/smsbao"
	"strconv"
	"time"
)

func ConfToServants(conf *conf.AppConfig) (error, *server.Servants) {
	//step1 init db
	smsRedisClient := redis.NewClient(&redis.Options{
		Password: conf.Redis.Password,
		Addr:     conf.Redis.Host + ":" + strconv.Itoa(conf.Redis.Port),
		DB:       conf.SmsBao.DB,
	})
	smsRedisCtx := context.Background()

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

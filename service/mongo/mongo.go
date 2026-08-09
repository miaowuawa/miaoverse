package mongo

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"miaoverse/consts"
)

// Config MongoDB 连接配置。
type Config struct {
	Host                   string
	Port                   int
	Username               string
	Password               string
	AuthSource             string
	Database               string
	MaxPoolSize            uint64
	MinPoolSize            uint64
	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration
}

// Servant MongoDB 连接器，持有驱动 client 与目标数据库句柄。
type Servant struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewServant 建立 MongoDB 连接并做一次 Ping 校验。
// 使用 options 直接构造，不拼接带凭据的 URI，避免凭据出现在错误信息或日志中。
func NewServant(ctx context.Context, conf Config) (*Servant, error) {
	opts, err := buildClientOptions(conf)
	if err != nil {
		return nil, err
	}

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("mongodb connect: %w", err)
	}

	connectTimeout := normalizeDuration(conf.ConnectTimeout, consts.MongoDefaultConnectTimeout)
	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("mongodb ping: %w", err)
	}

	return &Servant{
		client:   client,
		database: client.Database(conf.Database),
	}, nil
}

// Client 返回底层驱动 client。
func (s *Servant) Client() *mongo.Client {
	return s.client
}

// Database 返回配置指定数据库句柄。
func (s *Servant) Database() *mongo.Database {
	return s.database
}

// Close 断开 MongoDB 连接。
func (s *Servant) Close(ctx context.Context) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Disconnect(ctx)
}

// buildClientOptions 校验配置并构造驱动选项，不发起网络请求。
func buildClientOptions(conf Config) (*options.ClientOptions, error) {
	host := strings.TrimSpace(conf.Host)
	if host == "" {
		return nil, errors.New("mongodb host is required")
	}
	if strings.TrimSpace(conf.Database) == "" {
		return nil, errors.New("mongodb db_name is required")
	}
	port := conf.Port
	if port == 0 {
		port = consts.MongoDefaultPort
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("mongodb port must be in range 1-65535, got %d", port)
	}

	opts := options.Client().
		SetHosts([]string{net.JoinHostPort(host, strconv.Itoa(port))}).
		SetConnectTimeout(normalizeDuration(conf.ConnectTimeout, consts.MongoDefaultConnectTimeout)).
		SetServerSelectionTimeout(normalizeDuration(conf.ServerSelectionTimeout, consts.MongoDefaultServerSelectionTimeout)).
		SetMaxPoolSize(conf.MaxPoolSize)
	if conf.MinPoolSize > 0 {
		opts.SetMinPoolSize(conf.MinPoolSize)
	}
	if username := strings.TrimSpace(conf.Username); username != "" {
		cred := options.Credential{
			Username: username,
			Password: conf.Password,
		}
		if authSource := strings.TrimSpace(conf.AuthSource); authSource != "" {
			cred.AuthSource = authSource
		}
		opts.SetAuth(cred)
	}

	return opts, nil
}

func normalizeDuration(value time.Duration, def time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return def
}

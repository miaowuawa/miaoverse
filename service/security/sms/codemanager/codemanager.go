package codemanager

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type CodeManager struct {
	Redis          *redis.Client
	CodeExpireTime time.Duration
	Context        context.Context
}

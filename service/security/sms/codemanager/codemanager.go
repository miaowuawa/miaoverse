package codemanager

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type CodeManager struct {
	Redis          *redis.Client
	CodeExpireTime time.Duration
	Context        context.Context
}

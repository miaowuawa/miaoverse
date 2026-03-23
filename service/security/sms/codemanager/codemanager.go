package codemanager

import (
	"context"
	"github.com/redis/go-redis/v9"

	"time"
)

type CodeManager struct {
	Redis          *redis.Client
	CodeExpireTime time.Duration
	Context        context.Context
}

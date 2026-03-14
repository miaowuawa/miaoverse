package codemanager

import (
	"github.com/go-redis/redis/v8"
)

type CodeManager struct {
	Redis *redis.Client
}

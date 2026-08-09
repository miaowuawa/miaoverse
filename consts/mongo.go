package consts

import "time"

// Mongo 域常量：默认端口与连接超时。

const MongoDefaultPort = 27017

const (
	MongoDefaultConnectTimeout         = 10 * time.Second
	MongoDefaultServerSelectionTimeout = 10 * time.Second
)

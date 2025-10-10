package config

import "time"

type Config struct {
	WebServerPort   string
	DatabaseDSN     string
	RedisAddr       string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	GRPCServerPort  string
	KafkaAddr       string
}

package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// DatabaseClient 是通用的数据库接口
type DatabaseClient interface {
	// SetKey 设置键值
	SetKey(key string, value string) error

	// GetKey 获取键值
	GetKey(key string) (string, error)

	// Ping 测试数据库连接
	Ping() error
}

// DatabaseType 定义了支持的数据库类型
type DatabaseType string

const (
	RedisType DatabaseType = "redis"
	// 可以在这里添加更多数据库类型，比如 MySQL、MongoDB 等
)

// NewDatabaseClient 根据数据库类型返回相应的 DatabaseClient 实现
func NewDatabaseClient(dbType DatabaseType) (DatabaseClient, error) {
	switch dbType {
	case RedisType:
		return NewRedisClient()
	// 添加更多数据库的实现
	default:
		return nil, errors.New("unsupported database type")
	}
}

// RedisClient 是 Redis 数据库的客户端实现
type RedisClient struct {
	Client *redis.Client
	Ctx    context.Context
}

// RedisParam 存储 Redis 连接的配置
type RedisParam struct {
	Addr     string        // Redis 服务器地址
	Password string        // Redis 密码
	DB       int           // Redis 数据库编号
	Timeout  time.Duration // 连接和操作的超时
}

// 初始化 Redis 配置
func initRedisParam() *RedisParam {
	return &RedisParam{
		Addr:     "localhost:6379",
		Password: "",              // Redis 密码
		DB:       0,               // 默认数据库
		Timeout:  5 * time.Second, // 默认超时 5 秒
	}
}

// NewRedisClient 创建一个新的 Redis 客户端
func NewRedisClient() (*RedisClient, error) {
	redisParam := initRedisParam()
	client := redis.NewClient(&redis.Options{
		Addr:         redisParam.Addr,
		Password:     redisParam.Password,
		DB:           redisParam.DB,
		DialTimeout:  redisParam.Timeout, // 设置连接超时
		ReadTimeout:  redisParam.Timeout, // 设置读取超时
		WriteTimeout: redisParam.Timeout, // 设置写入超时
	})

	// 测试连接
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %v", err)
	}

	log.Println("连接 Redis 成功!")

	return &RedisClient{
		Client: client,
		Ctx:    context.Background(),
	}, nil
}

// 实现 DatabaseClient 接口的 SetKey 方法
func (r *RedisClient) SetKey(key string, value string) error {
	err := r.Client.Set(r.Ctx, key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("设置键值失败: %v", err)
	}
	return nil
}

// 实现 DatabaseClient 接口的 GetKey 方法
func (r *RedisClient) GetKey(key string) (string, error) {
	val, err := r.Client.Get(r.Ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("获取键值失败: %v", err)
	}
	return val, nil
}

// 实现 Ping 方法，测试数据库连接
func (r *RedisClient) Ping() error {
	_, err := r.Client.Ping(r.Ctx).Result()
	if err != nil {
		return fmt.Errorf("redis Ping 失败: %v", err)
	}
	return nil
}

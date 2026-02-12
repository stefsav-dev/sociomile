package utils

import (
	"backend/config"
	"encoding/json"
	"fmt"
	"time"
)

func IsRateLimited(key string, limit int, duration time.Duration) (bool, int) {
	redisKey := fmt.Sprintf("ratelimit:%s", key)

	count, err := config.RedisClient.Incr(config.Ctx, redisKey).Result()
	if err != nil {
		return false, 0
	}

	if count == 1 {
		config.RedisClient.Expire(config.Ctx, redisKey, duration)
	}

	return int(count) > limit, int(count)
}

func IncrementFailedLogin(email string) error {
	key := fmt.Sprintf("failed_login:%s", email)
	_, err := config.RedisClient.Incr(config.Ctx, key).Result()
	if err != nil {
		return err
	}
	config.RedisClient.Expire(config.Ctx, key, 15*time.Minute)
	return nil
}

func ResetFailedLogin(email string) error {
	key := fmt.Sprintf("failed_login:%s", email)
	return config.RedisClient.Del(config.Ctx, key).Err()
}

func StoreRefreshToken(userID uint, token string, expiration time.Duration) error {
	key := fmt.Sprintf("refresh:%d:%s", userID, token)
	return SetCache(key, true, expiration)
}

func ValidateRefreshToken(userID uint, token string) bool {
	key := fmt.Sprintf("refresh:%d:%s", userID, token)
	_, err := config.RedisClient.Get(config.Ctx, key).Result()
	return err == nil
}

func BlacklistToken(token string, expiration time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", token)
	return SetCache(key, true, expiration)
}

func IsTokenBlacklisted(token string) bool {
	key := fmt.Sprintf("blacklist:%s", token)
	_, err := config.RedisClient.Get(config.Ctx, key).Result()
	return err == nil
}

func SetCache(key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return config.RedisClient.Set(config.Ctx, key, jsonData, expiration).Err()
}

func GetCache(key string, dest interface{}) error {
	val, err := config.RedisClient.Get(config.Ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func DeleteCache(key string) error {
	return config.RedisClient.Del(config.Ctx, key).Err()
}

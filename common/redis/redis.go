package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type Client struct {
	*redis.Client
}

type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func NewClient(cfg Config) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	logx.Infof("Redis connected: %s:%s", cfg.Host, cfg.Port)

	return &Client{Client: client}, nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, string(data), expiration)
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.Client.Del(ctx, keys...).Err()
}

func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, value, expiration).Result()
}

func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.Client.Expire(ctx, key, expiration).Result()
}

func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.Client.Exists(ctx, keys...).Result()
}

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}

func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.Client.Decr(ctx, key).Result()
}

func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.Client.Eval(ctx, script, keys, args...).Result()
}

func (c *Client) EvalSha(ctx context.Context, sha string, keys []string, args ...interface{}) (interface{}, error) {
	return c.Client.EvalSha(ctx, sha, keys, args...).Result()
}

func (c *Client) ScriptLoad(ctx context.Context, script string) (string, error) {
	return c.Client.ScriptLoad(ctx, script).Result()
}

func (c *Client) HExists(ctx context.Context, key string, field string) (bool, error) {
	return c.Client.HExists(ctx, key, field).Result()
}

func (c *Client) HGet(ctx context.Context, key string, field string) (string, error) {
	return c.Client.HGet(ctx, key, field).Result()
}

func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.Client.HGetAll(ctx, key).Result()
}

func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.Client.SIsMember(ctx, key, member).Result()
}

func (c *Client) SMIsMember(ctx context.Context, key string, members ...interface{}) ([]bool, error) {
	return c.Client.SMIsMember(ctx, key, members...).Result()
}

func (c *Client) BatchSIsMember(ctx context.Context, key string, members []string) ([]bool, error) {
	if len(members) == 0 {
		return []bool{}, nil
	}

	args := make([]interface{}, 0, len(members))
	for _, member := range members {
		args = append(args, member)
	}

	result, err := c.SMIsMember(ctx, key, args...)
	if err == nil {
		return result, nil
	}

	errText := strings.ToLower(err.Error())
	if !strings.Contains(errText, "unknown command") && !strings.Contains(errText, "wrong number of arguments") {
		return nil, err
	}

	pipe := c.Client.Pipeline()
	cmds := make([]*redis.BoolCmd, 0, len(members))
	for _, member := range members {
		cmds = append(cmds, pipe.SIsMember(ctx, key, member))
	}

	if _, pipeErr := pipe.Exec(ctx); pipeErr != nil && pipeErr != redis.Nil {
		return nil, pipeErr
	}

	result = make([]bool, len(cmds))
	for i, cmd := range cmds {
		member, cmdErr := cmd.Result()
		if cmdErr != nil && cmdErr != redis.Nil {
			return nil, cmdErr
		}
		result[i] = member
	}

	return result, nil
}

func (c *Client) PipelineSIsMember(ctx context.Context, checks map[string]string) (map[string]bool, error) {
	if len(checks) == 0 {
		return map[string]bool{}, nil
	}

	pipe := c.Client.Pipeline()
	cmds := make(map[string]*redis.BoolCmd, len(checks))
	for key, member := range checks {
		cmds[key] = pipe.SIsMember(ctx, key, member)
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}

	result := make(map[string]bool, len(cmds))
	for key, cmd := range cmds {
		isMember, err := cmd.Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}
		result[key] = isMember
	}

	return result, nil
}

// HMSet 设置哈希表的多个字段
func (c *Client) HMSet(ctx context.Context, key string, fields map[string]interface{}) error {
	return c.Client.HMSet(ctx, key, fields).Err()
}

// HIncrBy 为哈希表中的字段值加上指定增量
func (c *Client) HIncrBy(ctx context.Context, key string, field string, value int64) error {
	_, err := c.Client.HIncrBy(ctx, key, field, value).Result()
	return err
}

// ZAdd 向有序集合添加一个或多个成员
func (c *Client) ZAdd(ctx context.Context, key string, members ...Z) error {
	// 转换为redis.Z类型
	redisMembers := make([]redis.Z, len(members))
	for i, member := range members {
		redisMembers[i] = redis.Z{
			Score:  member.Score,
			Member: member.Member,
		}
	}
	_, err := c.Client.ZAdd(ctx, key, redisMembers...).Result()
	return err
}

// ZRevRangeByScore 从有序集合中按分数范围倒序获取成员
func (c *Client) ZRevRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error) {
	return c.Client.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
}

// ZRange 获取有序集合中指定索引范围内的成员
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.Client.ZRange(ctx, key, start, stop).Result()
}

// ZCard 获取有序集合的成员数量
func (c *Client) ZCard(ctx context.Context, key string) (int64, error) {
	return c.Client.ZCard(ctx, key).Result()
}

// ZRem 从有序集合中移除一个或多个成员
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) error {
	_, err := c.Client.ZRem(ctx, key, members...).Result()
	return err
}

// IncrBy 原子性地为 key 的值增加增量
func (c *Client) IncrBy(ctx context.Context, key string, increment int64) (int64, error) {
	return c.Client.IncrBy(ctx, key, increment).Result()
}

// Z 有序集合成员

type Z struct {
	Score  float64
	Member interface{}
}

// SlideWindowLimit 滑动窗口限流
// key: 限流键
// windowSec: 窗口大小（秒）
// limit: 窗口内最大请求数
func (c *Client) SlideWindowLimit(key string, windowSec int, limit int) error {
	ctx := context.Background()
	now := time.Now().Unix()
	windowStart := now - int64(windowSec)

	// 移除窗口外的记录
	_, err := c.Client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart)).Result()
	if err != nil {
		return err
	}

	// 统计当前窗口内的请求数
	count, err := c.Client.ZCard(ctx, key).Result()
	if err != nil {
		return err
	}

	if count >= int64(limit) {
		return fmt.Errorf("rate limit exceeded")
	}

	// 添加当前请求到窗口
	_, err = c.Client.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d", now),
	}).Result()
	if err != nil {
		return err
	}

	// 设置过期时间，避免内存泄漏
	_, err = c.Client.Expire(ctx, key, time.Duration(windowSec)*time.Second).Result()
	if err != nil {
		return err
	}

	return nil
}

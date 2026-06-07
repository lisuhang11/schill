package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	commonredis "SChill/common/redis"
	"SChill/service/relation/rpc/internal/svc"

	gzredis "github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	relationCacheFollowing = "following"
	relationCacheFollowers = "followers"
)

const followRelationLua = `
local following_key = KEYS[1]
local followers_key = KEYS[2]
local following_empty_key = KEYS[3]
local followers_empty_key = KEYS[4]

local followee_id = ARGV[1]
local follower_id = ARGV[2]
local score = ARGV[3]
local ttl = tonumber(ARGV[4])

local added = redis.call("ZADD", following_key, "NX", score, followee_id)
if added == 1 then
    redis.call("ZADD", followers_key, "NX", score, follower_id)
end

redis.call("DEL", following_empty_key, followers_empty_key)
redis.call("EXPIRE", following_key, ttl)
redis.call("EXPIRE", followers_key, ttl)

return added
`

const unfollowRelationLua = `
local following_key = KEYS[1]
local followers_key = KEYS[2]
local following_empty_key = KEYS[3]
local followers_empty_key = KEYS[4]

local followee_id = ARGV[1]
local follower_id = ARGV[2]
local ttl = tonumber(ARGV[3])

local removed = redis.call("ZREM", following_key, followee_id)
if removed == 1 then
    redis.call("ZREM", followers_key, follower_id)
end

redis.call("DEL", following_empty_key, followers_empty_key)

if redis.call("ZCARD", following_key) == 0 then
    redis.call("DEL", following_key)
else
    redis.call("EXPIRE", following_key, ttl)
end

if redis.call("ZCARD", followers_key) == 0 then
    redis.call("DEL", followers_key)
else
    redis.call("EXPIRE", followers_key, ttl)
end

return removed
`

type relationCacheEntry struct {
	UserID     uint64
	TargetID   uint64
	FollowTime int64
}

func relationFollowingKey(userID uint64) string {
	return commonredis.RelationFollowingZSetKey + strconv.FormatUint(userID, 10)
}

func relationFollowersKey(userID uint64) string {
	return commonredis.RelationFollowersZSetKey + strconv.FormatUint(userID, 10)
}

func relationFollowingEmptyKey(userID uint64) string {
	return commonredis.RelationFollowingEmptyKey + strconv.FormatUint(userID, 10)
}

func relationFollowersEmptyKey(userID uint64) string {
	return commonredis.RelationFollowersEmptyKey + strconv.FormatUint(userID, 10)
}

func relationListKey(kind string, userID uint64) string {
	if kind == relationCacheFollowers {
		return relationFollowersKey(userID)
	}
	return relationFollowingKey(userID)
}

func relationEmptyKey(kind string, userID uint64) string {
	if kind == relationCacheFollowers {
		return relationFollowersEmptyKey(userID)
	}
	return relationFollowingEmptyKey(userID)
}

func relationRebuildKey(kind string, userID uint64) string {
	return fmt.Sprintf("relation:rebuild:%s:%d", kind, userID)
}

func ensureRelationCache(ctx context.Context, svcCtx *svc.ServiceContext, kind string, userID uint64) error {
	cacheKey := relationListKey(kind, userID)
	exists, err := svcCtx.Redis.ExistsCtx(ctx, cacheKey)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	var marker bool
	if err := svcCtx.Cache.GetCtx(ctx, relationEmptyKey(kind, userID), &marker); err == nil && marker {
		return nil
	}

	_, _, err = svcCtx.SingleFlight.DoEx(relationRebuildKey(kind, userID), func() (any, error) {
		exists, err := svcCtx.Redis.ExistsCtx(ctx, cacheKey)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, nil
		}

		var marker bool
		if err := svcCtx.Cache.GetCtx(ctx, relationEmptyKey(kind, userID), &marker); err == nil && marker {
			return nil, nil
		}

		rows, err := loadRelationCacheRows(ctx, svcCtx, kind, userID)
		if err != nil {
			return nil, err
		}

		if _, err := svcCtx.Redis.DelCtx(ctx, cacheKey); err != nil {
			return nil, err
		}

		if len(rows) == 0 {
			return nil, svcCtx.Cache.SetWithExpireCtx(
				ctx,
				relationEmptyKey(kind, userID),
				true,
				time.Duration(commonredis.RelationEmptyExpire)*time.Second,
			)
		}

		pairs := make([]gzredis.Pair, 0, len(rows))
		for _, row := range rows {
			member := strconv.FormatUint(row.TargetID, 10)
			if kind == relationCacheFollowers {
				member = strconv.FormatUint(row.UserID, 10)
			}
			pairs = append(pairs, gzredis.Pair{
				Key:   member,
				Score: row.FollowTime,
			})
		}

		if _, err := svcCtx.Redis.ZaddsCtx(ctx, cacheKey, pairs...); err != nil {
			return nil, err
		}
		if err := svcCtx.Redis.ExpireCtx(ctx, cacheKey, commonredis.RelationCacheExpire); err != nil {
			return nil, err
		}
		if err := svcCtx.Cache.DelCtx(ctx, relationEmptyKey(kind, userID)); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

func loadRelationCacheRows(ctx context.Context, svcCtx *svc.ServiceContext, kind string, userID uint64) ([]relationCacheEntry, error) {
	var rows []relationCacheEntry
	query := svcCtx.DB.WithContext(ctx).
		Table("following f").
		Select("f.user_id, f.follow_id as target_id, CAST(FLOOR(UNIX_TIMESTAMP(f.created_at)) AS SIGNED) as follow_time")

	if kind == relationCacheFollowers {
		query = query.
			Joins("JOIN user u ON u.user_id = f.user_id AND u.deleted_at IS NULL").
			Where("f.follow_id = ?", userID)
	} else {
		query = query.
			Joins("JOIN user u ON u.user_id = f.follow_id AND u.deleted_at IS NULL").
			Where("f.user_id = ?", userID)
	}

	err := query.Order("f.created_at DESC").Find(&rows).Error
	return rows, err
}

func relationCount(ctx context.Context, svcCtx *svc.ServiceContext, kind string, userID uint64) (int64, error) {
	if err := ensureRelationCache(ctx, svcCtx, kind, userID); err != nil {
		return 0, err
	}

	count, err := svcCtx.Redis.ZcardCtx(ctx, relationListKey(kind, userID))
	return int64(count), err
}

func relationMembersWithScores(ctx context.Context, svcCtx *svc.ServiceContext, kind string, userID uint64, page, pageSize int64) ([]gzredis.Pair, error) {
	if err := ensureRelationCache(ctx, svcCtx, kind, userID); err != nil {
		return nil, err
	}

	start := (page - 1) * pageSize
	stop := start + pageSize - 1
	return svcCtx.Redis.ZrevrangeWithScoresCtx(ctx, relationListKey(kind, userID), start, stop)
}

func relationExists(ctx context.Context, svcCtx *svc.ServiceContext, ownerID, targetID uint64) (bool, error) {
	if err := ensureRelationCache(ctx, svcCtx, relationCacheFollowing, ownerID); err != nil {
		return false, err
	}

	_, err := svcCtx.Redis.ZscoreCtx(ctx, relationFollowingKey(ownerID), strconv.FormatUint(targetID, 10))
	if err == nil {
		return true, nil
	}
	if err == gzredis.Nil {
		return false, nil
	}

	return false, err
}

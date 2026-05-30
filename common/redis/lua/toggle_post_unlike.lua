--[[
    功能：取消点赞操作（原子性）- 支持 Hash 关系存储与 ZSet 用户历史清理
    设计对应：
        - 与点赞脚本对称，取消时需同步清理点赞关系 Hash、计数器减 1、从用户历史 ZSet 中移除该内容。
        - 三个 Key 的设计保持了数据维度的一致性，便于后续维护与扩展。

    参数说明：
        KEYS[1] - 点赞关系 Hash 的 Key，格式: schill:like:entity:post:{post_id}
        KEYS[2] - 点赞计数器 String 的 Key，格式: schill:like_count:entity:post:{post_id}
        KEYS[3] - 用户点赞历史 ZSet 的 Key，格式: schill:user_likes:{user_id}
        ARGV[1] - 当前操作用户的 ID
        ARGV[2] - 内容 ID（post_id）

    返回值：
        包含两个元素的 Lua 数组：{ status, count }
        - status: 1 表示用户之前未点赞（本次未执行操作），0 表示本次成功取消点赞
        - count:  当前最新的点赞总数（整数，最小值为 0）
--]]

-- 1. 检查用户是否存在于该内容的点赞关系 Hash 中
local relation_key = KEYS[1]
local counter_key = KEYS[2]
local user_likes_key = KEYS[3]
local user_id = ARGV[1]
local post_id = ARGV[2]

local exists = redis.call('HEXISTS', relation_key, user_id)

-- 2. 如果用户不在集合中（即从未点赞），则无需执行任何取消操作
if exists == 0 then
    -- 获取当前计数并返回，保证幂等性
    local current_count = redis.call('GET', counter_key)
    return {1, tonumber(current_count or 0)}
end

-- 3. 用户已点赞，执行取消点赞的写操作
-- 3.1 从关系 Hash 中删除该用户字段
redis.call('HDEL', relation_key, user_id)

-- 3.2 将计数器原子性地减 1，并获取最新的计数值
local new_count = redis.call('DECR', counter_key)

-- 3.3 如果传入了用户点赞历史的 ZSet Key（非空），则移除该内容 ID
--     使用 ZREM 命令删除 Member，ZSet 中不再保留该点赞记录
if user_likes_key and user_likes_key ~= '' then
    redis.call('ZREM', user_likes_key, post_id)
end

-- 4. 防御性编程：确保计数不会因异常情况（如数据不一致）变成负数
if new_count < 0 then
    -- 若出现负数，强制将计数器修正为 0
    new_count = 0
    redis.call('SET', counter_key, 0)
    -- 生产环境中建议在此处触发告警或记录日志，以便排查数据不一致的根本原因
end

-- 5. 返回状态码 0（表示本次成功取消点赞）和最新的点赞总数
return {0, new_count}
--[[
    功能：点赞操作（原子性）- 支持 Hash 存储关系、ZSet 记录用户历史
    设计思路：
        - 点赞关系使用 Hash 存储，Key 为内容维度，Field 为用户ID，Value 为点赞时间戳。
          相较于 Set，Hash 可以额外存储点赞时间，便于后续分析或排序。
        - 点赞计数器仍使用 String，保持高性能读取。
        - 用户点赞历史使用 ZSet 存储，Score 为时间戳，Member 为内容ID，
          方便查询用户最近点赞了哪些内容（如个人主页展示）。

    参数说明：
        KEYS[1] - 点赞关系 Hash 的 Key，格式: schill:like:entity:post:{post_id}
        KEYS[2] - 点赞计数器 String 的 Key，格式: schill:like_count:entity:post:{post_id}
        KEYS[3] - 用户点赞历史 ZSet 的 Key，格式: schill:user_likes:{user_id}
        ARGV[1] - 当前操作用户的 ID
        ARGV[2] - 当前操作的时间戳（秒级）
        ARGV[3] - 内容 ID（post_id）

    返回值：
        一个包含两个元素的 Lua 数组：{ status, count }
        - status: 1 表示用户之前已点赞（本次未执行操作），0 表示本次成功点赞
        - count:  当前最新的点赞总数（整数）
--]]

-- 1. 检查用户是否已经存在于该内容的点赞关系 Hash 中
local relation_key = KEYS[1]
local counter_key = KEYS[2]
local user_likes_key = KEYS[3]
local user_id = ARGV[1]
local timestamp = ARGV[2]
local post_id = ARGV[3]

local exists = redis.call('HEXISTS', relation_key, user_id)

-- 2. 如果用户已经点赞，则直接返回当前计数，不执行任何写操作（幂等性保障）
if exists == 1 then
    local current_count = redis.call('GET', counter_key)
    -- 返回状态码 1（已点赞）和当前计数
    return {1, tonumber(current_count or 0)}
end

-- 3. 用户未点赞，执行点赞的写操作
-- 3.1 将用户 ID 和点赞时间戳写入关系 Hash
redis.call('HSET', relation_key, user_id, timestamp)

-- 3.2 将点赞计数器原子性地加 1，并获取最新的计数值
local new_count = redis.call('INCR', counter_key)

-- 3.3 如果传入了用户点赞历史的 ZSet Key（非空），则将内容 ID 和时间戳写入用户历史
--     使用 ZADD 命令，Score 为时间戳，Member 为内容 ID
if user_likes_key and user_likes_key ~= '' then
    redis.call('ZADD', user_likes_key, timestamp, post_id)
end

-- 4. 返回状态码 0（表示本次成功点赞）和最新的点赞总数
return {0, new_count}
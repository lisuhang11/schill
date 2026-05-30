--[[
    功能：取消收藏操作（原子性）
    参数说明：
        KEYS[1] - 收藏用户集合的 Key，格式: collect:post:{contentId}
        KEYS[2] - 收藏计数器的 Key，格式: collect:count:{contentId}
        ARGV[1] - 当前操作用户的 ID
    返回值：
        包含两个元素的 Lua 数组：{ status, count }
        - status: 1 表示用户之前未收藏（本次未执行操作），0 表示本次成功取消收藏
        - count:  当前最新的收藏总数（整数，最小值为 0）
--]]

-- 1. 检查用户是否存在于收藏集合中（实现幂等性的关键步骤）
local is_member = redis.call('SISMEMBER', KEYS[1], ARGV[1])

-- 2. 如果用户不在集合中（即未收藏），则无需执行任何取消操作
if is_member == 0 then
    -- 获取当前收藏计数，若计数器不存在则默认为 0
    local current_count = redis.call('GET', KEYS[2])
    -- 返回状态码 1（表示未收藏，操作无效）和当前计数
    return {1, tonumber(current_count or 0)}
end

-- 3. 用户已收藏，执行取消收藏的写操作
-- 从收藏集合中移除该用户 ID
redis.call('SREM', KEYS[1], ARGV[1])
-- 将计数器原子性地减 1，并获取减少后的最新数值
local new_count = redis.call('DECR', KEYS[2])

-- 4. 防御性编程：确保计数不会因异常情况（如数据不一致）变成负数
if new_count < 0 then
    -- 若出现负数，强制将计数器修正为 0
    new_count = 0
    redis.call('SET', KEYS[2], 0)
    -- 生产环境中建议在此处触发告警或记录日志，以便排查数据不一致的根本原因
end

-- 5. 返回状态码 0（表示本次成功取消收藏）和最新的收藏总数
return {0, new_count}

--[[
    功能：点赞操作（原子性 + TTL）
    KEYS[1] - 点赞用户集合的 Key
    KEYS[2] - 点赞计数器的 Key
    ARGV[1] - 当前操作用户的 ID
    ARGV[2] - TTL (秒)
    返回值：{ status, count }
--]]

local is_member = redis.call('SISMEMBER', KEYS[1], ARGV[1])

if is_member == 1 then
    local current_count = redis.call('GET', KEYS[2])
    -- Refresh TTL on access
    redis.call('EXPIRE', KEYS[1], ARGV[2])
    redis.call('EXPIRE', KEYS[2], ARGV[2])
    return {1, tonumber(current_count or 0)}
end

redis.call('SADD', KEYS[1], ARGV[1])
local new_count = redis.call('INCR', KEYS[2])
redis.call('EXPIRE', KEYS[1], ARGV[2])
redis.call('EXPIRE', KEYS[2], ARGV[2])

return {0, new_count}

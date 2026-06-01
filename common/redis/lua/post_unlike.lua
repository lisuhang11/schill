--[[
    功能：取消点赞操作（原子性 + TTL）
    KEYS[1] - 点赞用户集合的 Key
    KEYS[2] - 点赞计数器的 Key
    ARGV[1] - 当前操作用户的 ID
    ARGV[2] - TTL (秒)
    返回值：{ status, count }
--]]

local is_member = redis.call('SISMEMBER', KEYS[1], ARGV[1])

if is_member == 0 then
    local current_count = redis.call('GET', KEYS[2])
    return {1, tonumber(current_count or 0)}
end

redis.call('SREM', KEYS[1], ARGV[1])
local new_count = redis.call('DECR', KEYS[2])

if new_count < 0 then
    new_count = 0
    redis.call('SET', KEYS[2], 0)
end

-- If set is empty after removal, delete it (cleanup); otherwise refresh TTL
local card = redis.call('SCARD', KEYS[1])
if card == 0 then
    redis.call('DEL', KEYS[1])
    redis.call('DEL', KEYS[2])
else
    redis.call('EXPIRE', KEYS[1], ARGV[2])
    redis.call('EXPIRE', KEYS[2], ARGV[2])
end

return {0, new_count}

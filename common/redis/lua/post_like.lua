--[[
    功能：点赞操作（原子性）
    参数说明：
        KEYS[1] - 点赞用户集合的 Key，格式: like:post:{contentId}
        KEYS[2] - 点赞计数器的 Key，格式: like:count:{contentId}
        ARGV[1] - 当前操作用户的 ID
    返回值：
        一个包含两个元素的 Lua 数组：
        { status, count }
        - status: 1 表示用户之前已点赞（本次未执行操作），0 表示本次成功点赞
        - count:  当前最新的点赞总数（整数）
--]]

-- 1. 检查用户是否已经存在于点赞集合中（实现幂等性的关键步骤）
local is_member = redis.call('SISMEMBER', KEYS[1], ARGV[1])

-- 2. 如果用户已经点赞，则直接返回当前计数，不进行任何写操作
if is_member == 1 then
    -- 获取当前的点赞计数器数值，若计数器不存在则默认为 0
    local current_count = redis.call('GET', KEYS[2])
    -- 返回状态码 1（表示已点赞）和当前计数
    return {1, tonumber(current_count or 0)}
end

-- 3. 如果用户未点赞，执行点赞的写操作
-- 将用户 ID 加入点赞集合
redis.call('SADD', KEYS[1], ARGV[1])
-- 将点赞计数器原子性地加 1，并获取增加后的新数值
local new_count = redis.call('INCR', KEYS[2])

-- 4. 返回状态码 0（表示本次成功点赞）和最新的点赞总数
return {0, new_count}
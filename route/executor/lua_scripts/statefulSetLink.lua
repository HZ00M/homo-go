-- 设置Pod链接的Lua脚本
-- 参数: KEYS[1] = 链接键, ARGV[1] = Pod ID, ARGV[2] = 过期时间(秒), ARGV[3] = 服务名称, ARGV[4] = 命名空间, ARGV[5] = 用户ID
-- 返回: 1表示成功, 0表示失败

local key = KEYS[1]
local podId = tonumber(ARGV[1])
local expireSeconds = tonumber(ARGV[2])
local serviceName = ARGV[3]
local namespace = ARGV[4]
local uid = ARGV[5]

-- 检查参数
if not podId or not expireSeconds or not serviceName or not namespace or not uid then
    return 0
end

-- 构建链接数据
local linkData = {
    podId = podId,
    serviceName = serviceName,
    namespace = namespace,
    uid = uid,
    createTime = redis.call('TIME')[1],
    lastAccessTime = redis.call('TIME')[1]
}

-- 序列化数据
local linkJson = cjson.encode(linkData)

-- 设置到Redis
redis.call('SET', key, linkJson, 'EX', expireSeconds)

-- 记录操作日志
redis.call('LPUSH', 'logs:pod_links', string.format('SET %s -> pod %d at %s', key, podId, redis.call('TIME')[1]))

return 1

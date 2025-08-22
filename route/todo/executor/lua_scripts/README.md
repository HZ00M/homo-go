# Stateful Redis Lua脚本说明

本目录包含用于有状态服务管理的Redis Lua脚本，这些脚本确保Redis操作的原子性和一致性。

## 脚本列表

### 1. statefulSetLink.lua
**功能**：设置Pod链接
**参数**：
- KEYS[1]: UID与特定服务链接的Redis键
- KEYS[2]: UID链接信息的Redis键
- ARGV[1]: 服务名称
- ARGV[2]: Pod ID
- ARGV[3]: 持久时间（秒，-1表示永久）

**逻辑**：
1. 获取当前链接的Pod ID
2. 设置新的Pod链接
3. 如果持久时间为-1，则永久保存；否则设置过期时间
4. 更新哈希表中的链接信息

### 2. statefulTrySetLink.lua
**功能**：尝试设置Pod链接（仅在当前链接为持久链接时）
**参数**：同statefulSetLink.lua
**逻辑**：
1. 检查当前链接是否为持久链接
2. 如果是持久链接且与目标Pod不同，则返回当前Pod ID
3. 否则设置新链接并返回结果

### 3. statefulSetLinkIfAbsent.lua
**功能**：仅在不存在链接时设置Pod链接
**参数**：同statefulSetLink.lua
**逻辑**：
1. 检查是否存在链接
2. 如果不存在，则设置新链接
3. 返回当前链接的Pod ID

### 4. statefulRemoveLink.lua
**功能**：移除Pod链接
**参数**：
- KEYS[1]: UID与特定服务链接的Redis键
- KEYS[2]: UID链接信息的Redis键
- ARGV[1]: 服务名称
- ARGV[2]: 临时状态维持时间（秒）

**逻辑**：
1. 删除链接键
2. 如果指定了临时状态时间，则设置过期时间
3. 从哈希表中移除链接信息

### 5. statefulRemoveLinkWithId.lua
**功能**：仅在匹配指定Pod ID时移除Pod链接
**参数**：
- KEYS[1]: UID与特定服务链接的Redis键
- KEYS[2]: UID链接信息的Redis键
- ARGV[1]: 服务名称
- ARGV[2]: 临时状态维持时间（秒）
- ARGV[3]: 要移除的Pod ID

**逻辑**：
1. 检查当前链接是否匹配指定的Pod ID
2. 如果匹配，则执行移除操作
3. 返回操作结果（1表示成功，0表示失败）

### 6. statefulGetLinkIfPersist.lua
**功能**：仅在链接为持久链接时获取链接的Pod
**参数**：
- KEYS[1]: UID与特定服务链接的Redis键

**逻辑**：
1. 获取当前链接的Pod ID
2. 检查是否为持久链接（TTL为-1）
3. 如果是持久链接则返回Pod ID，否则返回-1

### 7. statefulComputeLinkIfAbsent.lua
**功能**：计算并设置Pod链接（如果不存在）
**参数**：
- KEYS[1]: 目标哈希表键
- KEYS[2]: 源哈希表键
- ARGV[1]: 目标键
- ARGV[2]: 最小权重阈值

**逻辑**：
1. 检查目标键是否存在
2. 如果不存在，从源哈希表中选择权重最低的Pod
3. 设置链接并返回选中的Pod ID

### 8. statefulGetLinkService.lua
**功能**：获取UID链接的所有服务
**参数**：
- KEYS[1]: UID链接信息的Redis键

**逻辑**：
1. 返回哈希表中的所有键值对
2. 键为服务名称，值为Pod ID

### 9. 其他脚本
- `statefulGetServicePod.lua`: 获取服务的Pod信息（空实现）
- `statefulGetService.lua`: 获取服务信息（空实现）
- `statefulSetState.lua`: 设置服务状态（空实现）
- `statefulGetLink.lua`: 获取链接信息（空实现）

## 使用说明

### 在Go代码中使用
```go
// 获取脚本管理器
scriptManager := NewStatefulLuaScriptManager()

// 执行脚本
result, err := scriptManager.ExecuteScriptInt(ctx, 
    scriptManager.GetStatefulSetLink(), 
    []string{uidSvcKey, uidKey}, 
    []interface{}{serviceName, podId, persistSeconds})
```

### 注意事项
1. 所有脚本都使用KEYS数组传递Redis键
2. 所有脚本都使用ARGV数组传递参数
3. 脚本返回的结果需要根据具体脚本进行解析
4. 空实现的脚本需要在Go代码中实现对应逻辑

## 兼容性
这些脚本与Java版本的StatefulRedisExecutor完全兼容，确保数据格式和操作逻辑的一致性。

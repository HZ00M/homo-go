# Entity 模块 6A 任务索引（按实施顺序）

## 任务状态概览

| 任务 | 状态 | 优先级 | 完成度 | 备注 |
|------|------|--------|--------|------|
| Task-01 | ✅ 已完成 | 🔴 高 | 100% | Actor队列机制，串行执行保障 |
| Task-02 | ✅ 已完成 | 🔴 高 | 100% | EntityMgr，实体生命周期管理 |
| Task-02-1 | ✅ 已完成 | 🔴 高 | 100% | Config接入，TTL卸载与周期保存 |
| Task-03 | 🔄 进行中 | 🔴 高 | 60% | AbilitySystem，能力系统初始化 |
| Task-04 | ❌ 未开始 | 🔴 高 | 0% | StorageSystem，存储系统管理 |
| Task-05 | ❌ 未开始 | 🔴 高 | 0% | StorageAbility，存储能力完善 |
| Task-06 | ❌ 未开始 | 🟡 中 | 0% | CallSystem，调用系统优化 |
| Task-07 | ❌ 未开始 | 🔴 高 | 0% | CallAbility服务端，中间件集成 |
| Task-08 | ❌ 未开始 | 🟡 中 | 0% | CallAbility客户端，强类型代理 |
| Task-09 | ❌ 未开始 | 🟢 低 | 0% | LocalCall，传输错误映射 |
| Task-10 | ❌ 未开始 | 🟡 中 | 0% | 生成器，可观测性体系 |

## 任务详细链接

### 已完成任务
- **Task-01**: 每实体串行与队列接入（Actor 接入） — [Task-01-actor-queue.md](./Task-01-actor-queue.md)
  - 状态: ✅ 已完成
  - 完成时间: 2025-08-14
  - 核心成果: Actor队列模型、Context取消支持、Panic恢复机制

- **Task-02**: 实体管理器 EntityMgr（生命周期/并发安全/接口完善） — [Task-02-entity-mgr.md](./Task-02-entity-mgr.md)
  - 状态: ✅ 已完成
  - 完成时间: 2025-08-14
  - 核心成果: 实体生命周期管理、并发安全控制、扩展能力

- **Task-02-1**: 接入 Config（TTL 卸载与周期保存） — [Task-03-config-ttl-save.md](./Task-03-config-ttl-save.md)
  - 状态: ✅ 已完成
  - 完成时间: 2025-08-14
  - 核心成果: TTL自动卸载、周期保存机制、动态配置更新

- **Task-03**: 实体能力系统 AbilitySystem（注册钩子/初始化策略） — [Task-03-ability-system.md](./Task-03-ability-system.md)
  - 状态: 🔄 进行中
  - 优先级: 🔴 高优先级
  - 核心功能: 统一能力管理、生命周期钩子、依赖注入支持

### 待开始任务
- **Task-04**: 存储系统 StorageSystem（管理 StorageAbility/定时保存/TTL） — [Task-04-storage-system.md](./Task-04-storage-system.md)
  - 状态: ❌ 未开始
  - 优先级: 🔴 高优先级
  - 核心功能: 统一存储管理、智能调度机制、错误处理与监控

- **Task-05**: 存储能力 StorageAbility（序列化/schema/错误处理） — [Task-05-storage-ability.md](./Task-05-storage-ability.md)
  - 状态: ❌ 未开始
  - 优先级: 🔴 高优先级
  - 核心功能: 稳定序列化、Schema版本管理、错误处理与监控

- **Task-06**: 调用系统 CallSystem（分发表/打包/LocalCall） — [Task-06-call-system.md](./Task-06-call-system.md)
  - 状态: ❌ 未开始
  - 优先级: 🟡 中优先级
  - 核心功能: 智能分发表管理、灵活编解码支持、高性能本地调用

- **Task-07**: 实体调用能力 CallAbility 服务端（鉴权/限流/恢复/日志/追踪/指标） — [Task-07-callability-server.md](./Task-07-callability-server.md)
  - 状态: ❌ 未开始
  - 优先级: 🔴 高优先级
  - 核心功能: 稳定服务保障、安全与限流、完整观测体系

- **Task-08**: 实体调用能力 CallAbility 客户端（强类型代理/selector 重试/广播） — [Task-08-callability-client.md](./Task-08-callability-client.md)
  - 状态: ❌ 未开始
  - 优先级: 🟡 中优先级
  - 核心功能: 强类型代理、智能选择器、广播支持

- **Task-09**: LocalCall 与传输错误映射（目标 pod metadata） — [Task-09-localcall-transport.md](./Task-09-localcall-transport.md)
  - 状态: ❌ 未开始
  - 优先级: 🟢 低优先级
  - 核心功能: LocalCall优化、错误映射完善、传输层优化

- **Task-10**: 生成器与可观测性（server 分发表、client 代理、日志/指标/Tracing） — [Task-10-generators-observability.md](./Task-10-generators-observability.md)
  - 状态: ❌ 未开始
  - 优先级: 🟡 中优先级
  - 核心功能: 代码生成器、可观测性体系、性能优化

## 6A 工作流规则

所有任务都遵循标准的6A工作流规则：

- **A1 目标（Aim）**: 定义任务的核心目标和业务价值
- **A2 分析（Analyze）**: 分析现状、差距和约束条件
- **A3 设计（Architect）**: 设计系统架构和实现方案
- **A4 行动（Act）**: 具体的实施计划和行动步骤
- **A5 验证（Assure）**: 验证实施结果的质量和正确性
- **A6 迭代（Advance）**: 后续优化和扩展计划

## 优先级说明

- 🔴 **高优先级**: 核心功能，必须优先完成
- 🟡 **中优先级**: 重要功能，按计划完成
- 🟢 **低优先级**: 增强功能，可延后完成

## 状态说明

- ✅ **已完成**: 功能完全实现并通过测试
- 🔄 **进行中**: 功能正在开发中
- ❌ **未开始**: 功能尚未开始开发
- ⚠️ **阻塞**: 功能开发被其他因素阻塞

## 下一步计划

1. **立即开始**: Task-03 AbilitySystem（当前进行中）
2. **优先启动**: Task-04 StorageSystem、Task-05 StorageAbility
3. **并行开发**: Task-06 CallSystem、Task-07 CallAbility服务端
4. **后续跟进**: Task-08 CallAbility客户端、Task-09 LocalCall
5. **收尾完善**: Task-10 生成器与可观测性

## 质量检查

所有任务必须通过以下质量检查清单：

- [ ] 代码质量检查完成（go vet, golangci-lint, go test -race）
- [ ] 文档质量检查完成（注释完整、接口一致、架构清晰）
- [ ] 测试质量检查完成（单元测试、集成测试、性能测试）

## 联系信息

如有问题或需要协助，请联系项目负责人或相关开发人员。  
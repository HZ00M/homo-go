package call

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/api/entity"
	"github.com/go-kratos/kratos/v2/entity/base"
	facade "github.com/go-kratos/kratos/v2/entity/facade"
	"github.com/go-kratos/kratos/v2/rpc"
)

func init() {
	// 预留：如需在包加载时进行全局注册可在此处进行
}

type CallSystemImpl struct {
	entityMgr facade.EntityMgr
	// type -> id -> CallAbleAbility
	t2Id2Call map[string]map[string]*CallAbleAbility
	// type -> id -> per-entity actor
	t2Id2Actor map[string]map[string]*base.Actor
	mu         sync.Mutex
	queueSize  int
}

// Init 将 CallAbleAbility 挂载流程注册到 EntityMgr，并保存引用
func (c *CallSystemImpl) Init(ctx context.Context, eMgr facade.EntityMgr) {
	c.entityMgr = eMgr
	if c.t2Id2Call == nil {
		c.t2Id2Call = make(map[string]map[string]*CallAbleAbility)
	}
	if c.t2Id2Actor == nil {
		c.t2Id2Actor = make(map[string]map[string]*base.Actor)
	}
	if c.queueSize <= 0 {
		c.queueSize = 8
	}
	// 按能力注册添加流程：为每个实体自动挂载 call 能力
	eMgr.RegisterAddProcess("call", func(ctx context.Context, owner facade.Entity) {
		abi := &CallAbleAbility{}
		_ = abi.Attach(ctx, owner)
		t := owner.Type()
		id := owner.ID()
		c.mu.Lock()
		mp, ok := c.t2Id2Call[t]
		if !ok {
			mp = make(map[string]*CallAbleAbility)
			c.t2Id2Call[t] = mp
		}
		mp[id] = abi
		c.mu.Unlock()
	})
	// 在移除时关闭并清理对应 actor
	eMgr.RegisterRemoveProcess("call", func(ctx context.Context, owner facade.Entity) {
		c.CloseActor(owner.Type(), owner.ID())
	})
}

// SetQueueSize 在 Init 之前设置默认队列大小
func (c *CallSystemImpl) SetQueueSize(size int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if size > 0 {
		c.queueSize = size
	}
}

func (c *CallSystemImpl) getActor(t, id string) *base.Actor {
	c.mu.Lock()
	defer c.mu.Unlock()
	mp, ok := c.t2Id2Actor[t]
	if !ok {
		mp = make(map[string]*base.Actor)
		c.t2Id2Actor[t] = mp
	}
	act := mp[id]
	if act == nil {
		act = base.NewActor(c.queueSize)
		mp[id] = act
	}
	return act
}

// CloseActor 关闭并移除某个实体的 actor（若存在）
func (c *CallSystemImpl) CloseActor(t, id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if mp, ok := c.t2Id2Actor[t]; ok {
		if act, ok2 := mp[id]; ok2 {
			act.Close()
			delete(mp, id)
		}
	}
}

// CloseAll 关闭所有 actor
func (c *CallSystemImpl) CloseAll() {
	c.mu.Lock()
	for _, mp := range c.t2Id2Actor {
		for id, act := range mp {
			act.Close()
			delete(mp, id)
		}
	}
	c.mu.Unlock()
}

// QueueLen 返回指定实体的队列长度（近似）
func (c *CallSystemImpl) QueueLen(t, id string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if mp, ok := c.t2Id2Actor[t]; ok {
		if act, ok2 := mp[id]; ok2 {
			return act.QueueLen()
		}
	}
	return 0
}

func (c *CallSystemImpl) Call(ctx context.Context, srcName string, funName string, req *entity.EntityRequest) ([][]byte, error) {
	t := req.Type
	id := req.Id

	// 确保实体已加载（并触发 AddProcess）
	owner, err := c.entityMgr.Get(ctx, t, id)
	if err != nil {
		return nil, err
	}

	// 获取/准备实体的 call 能力实例
	var callAbi *CallAbleAbility
	c.mu.Lock()
	if c.t2Id2Call != nil {
		if mp, ok := c.t2Id2Call[t]; ok {
			callAbi = mp[id]
		}
	}
	c.mu.Unlock()
	if callAbi == nil {
		// 懒挂载防御：若未通过流程挂载则即时挂载
		callAbi = &CallAbleAbility{}
		_ = callAbi.Attach(ctx, owner)
		c.mu.Lock()
		if c.t2Id2Call == nil {
			c.t2Id2Call = make(map[string]map[string]*CallAbleAbility)
		}
		mp, ok := c.t2Id2Call[t]
		if !ok {
			mp = make(map[string]*CallAbleAbility)
			c.t2Id2Call[t] = mp
		}
		mp[id] = callAbi
		c.mu.Unlock()
	}

	// 根据 funName 的分发表选择 packer，再构造对应的 RpcContent
	abiType, ok := fun2abiMap[funName]
	if !ok {
		return nil, facade.ErrMethodNotFound
	}
	disp := abi2callMap[abiType]
	if disp == nil {
		return nil, facade.ErrAbilityNotFound
	}
	fi := disp.RpcMethod[funName]
	var content rpc.RpcContent
	if fi != nil && fi.Packer != nil && fi.Packer.Name() == "json" { //todo 待优化
		var jsonStr string
		if len(req.Content) > 0 {
			jsonStr = string(req.Content[0])
		}
		content = &rpc.Content[string]{CType: rpc.RpcContentJson, Dt: jsonStr}
	} else {
		content = &rpc.Content[[][]byte]{CType: rpc.RpcContentBytes, Dt: req.Content}
	}

	// 按实体串行：通过 per-entity actor 排队执行
	actor := c.getActor(t, id)
	done := make(chan struct{})
	var ret rpc.RpcContent
	var callErr error
	if err := actor.Enqueue(ctx, func() {
		ret, callErr = callAbi.onCall(ctx, funName, content)
		close(done)
	}); err != nil {
		return nil, err
	}
	select {
	case <-done:
		// ok
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if callErr != nil {
		return nil, callErr
	}

	// 统一转换为 [][]byte 返回
	switch ret.Type() {
	case rpc.RpcContentBytes:
		if resp, ok := ret.Data().([][]byte); ok {
			return resp, nil
		}
		return nil, facade.ErrDecode
	case rpc.RpcContentJson:
		if s, ok := ret.Data().(string); ok {
			return [][]byte{[]byte(s)}, nil
		}
		return nil, facade.ErrDecode
	default:
		return nil, facade.ErrEncode
	}
}

func (c *CallSystemImpl) LocalCall(ctx context.Context, srcName string, funName string, params []any) ([]any, error) {
	// 可选：后续实现直接本地分发（不经编解码）
	return nil, facade.ErrMethodNotFound
}

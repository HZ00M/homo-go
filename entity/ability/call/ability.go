package call

import (
	"context"
	"reflect"

	"github.com/go-kratos/kratos/v2/entity/base"
	"github.com/go-kratos/kratos/v2/entity/facade"
	"github.com/go-kratos/kratos/v2/rpc"
)

var (
	// funName -> ability reflect.Type（统一使用指针类型）
	fun2abiMap = make(map[string]reflect.Type)
	// ability reflect.Type(ptr) -> CallDispatcher（可长期复用）
	abi2callMap = make(map[reflect.Type]*rpc.CallDispatcher)
)

// RegisterType 通过 Ability 实现类型进行注册（建议传入指针类型，例如：reflect.TypeOf((*MyAbility)(nil)))
func RegisterType(abilityType reflect.Type) {
	if abilityType == nil {
		return
	}
	// 规范化为指针类型，确保方法集完整
	ptrType := abilityType
	if ptrType.Kind() != reflect.Ptr {
		ptrType = reflect.PtrTo(abilityType)
	}
	// 构造零实例，并基于实例构建可复用的分发表
	inst := reflect.New(ptrType.Elem()).Interface()
	abi, _ := inst.(facade.Ability)
	abi2callMap[ptrType] = rpc.NewCallDispatcher(abi)

	// 基础能力类型集合，用于过滤（避免将 Name/Attach/Detach 以及 BaseAbility 的方法注册为业务方法）
	baseAbilityType := reflect.TypeOf((*base.BaseAbility)(nil)).Elem()

	// 扫描导出方法，建立 fun -> abilityType 映射，排除 Ability 基础方法
	for i := 0; i < ptrType.NumMethod(); i++ {
		m := ptrType.Method(i)
		if !m.IsExported() {
			continue
		}
		// 排除接口基础方法
		switch m.Name {
		case "Name", "Attach", "Detach":
			continue
		}
		// 排除由 BaseAbility 提供的方法（方法接收者为 BaseAbility 或其指针）
		if m.Func.Type().NumIn() > 0 {
			recv := m.Func.Type().In(0)
			// 形如 func (b *BaseAbility) Xxx(...)
			if recv == baseAbilityType || (recv.Kind() == reflect.Ptr && recv.Elem() == baseAbilityType) {
				continue
			}
		}
		fun2abiMap[m.Name] = ptrType
	}
}

type CallAbleAbility struct {
	base.BaseAbility
	typeName string
	owner    facade.Entity
}

func (a *CallAbleAbility) Name() string { return "call" }

// onCall 通过 funName 找到目标 Ability 类型，并使用已缓存的 dispatcher 分发
func (a *CallAbleAbility) onCall(ctx context.Context, funName string, req rpc.RpcContent) (rpc.RpcContent, error) {
	abiType, ok := fun2abiMap[funName]
	if !ok {
		return nil, facade.ErrMethodNotFound
	}
	disp := abi2callMap[abiType]
	if disp == nil {
		return nil, facade.ErrAbilityNotFound
	}
	return disp.Dispatch(ctx, funName, req)
}

package call

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"testing"

	"github.com/go-kratos/kratos/v2/api/entity"
	facade "github.com/go-kratos/kratos/v2/entity/facade"
)

type testEntity struct {
	id        string
	typeName  string
	loadTime  int64
	abilities map[string]facade.Ability
}

func (e *testEntity) Type() string                      { return e.typeName }
func (e *testEntity) ID() string                        { return e.id }
func (e *testEntity) SetID(id string)                   { e.id = id }
func (e *testEntity) Init(ctx context.Context) error    { return nil }
func (e *testEntity) Destroy(ctx context.Context) error { return nil }
func (e *testEntity) LoadTime() int64                   { return e.loadTime }
func (e *testEntity) GetAbility(name string) facade.Ability {
	if e.abilities == nil {
		return nil
	}
	return e.abilities[name]
}
func (e *testEntity) AddAbility(a facade.Ability) {
	if e.abilities == nil {
		e.abilities = make(map[string]facade.Ability)
	}
	e.abilities[a.Name()] = a
}

type addReq struct {
	A int `json:"a"`
	B int `json:"b"`
}
type addResp struct {
	Sum int `json:"sum"`
}

type adderAbility struct{ owner facade.Entity }

func (a *adderAbility) Name() string { return "adder" }
func (a *adderAbility) Attach(ctx context.Context, owner facade.Entity) error {
	a.owner = owner
	owner.AddAbility(a)
	return nil
}
func (a *adderAbility) Detach(ctx context.Context) error { a.owner = nil; return nil }
func (a *adderAbility) Add(ctx context.Context, req *addReq) (*addResp, error) {
	return &addResp{Sum: req.A + req.B}, nil
}

// fake manager implementing minimal subset
type fakeMgr struct {
	ents     map[string]map[string]facade.Entity // type->id->entity
	addProcs map[string][]func(ctx context.Context, e facade.Entity)
	mu       sync.Mutex
	added    map[string]map[string]bool // type->id->added
}

func (m *fakeMgr) Exists(ctx context.Context, entityType, id string) (bool, error) {
	_, ok := m.ents[entityType][id]
	return ok, nil
}
func (m *fakeMgr) Create(ctx context.Context, entityType, id string, ctor func() facade.Entity) (facade.Entity, error) {
	e := ctor()
	e.SetID(id)
	if m.ents == nil {
		m.ents = map[string]map[string]facade.Entity{}
	}
	if m.ents[entityType] == nil {
		m.ents[entityType] = map[string]facade.Entity{}
	}
	m.ents[entityType][id] = e
	return e, nil
}
func (m *fakeMgr) Get(ctx context.Context, entityType, id string) (facade.Entity, error) {
	if m.ents == nil || m.ents[entityType] == nil {
		return nil, facade.ErrNotFound
	}
	e, ok := m.ents[entityType][id]
	if !ok {
		return nil, facade.ErrNotFound
	}
	// simulate add process hooks only once per entity to avoid concurrent attach
	m.mu.Lock()
	if m.added == nil {
		m.added = make(map[string]map[string]bool)
	}
	if m.added[entityType] == nil {
		m.added[entityType] = make(map[string]bool)
	}
	did := m.added[entityType][id]
	if !did {
		m.added[entityType][id] = true
	}
	hs := m.addProcs["call"]
	m.mu.Unlock()
	if !did && len(hs) > 0 {
		for _, h := range hs {
			h(ctx, e)
		}
	}
	return e, nil
}
func (m *fakeMgr) GetNoKeepAlive(ctx context.Context, entityType, id string) (facade.Entity, error) {
	return m.Get(ctx, entityType, id)
}
func (m *fakeMgr) GetOrCreate(ctx context.Context, entityType, id string, ctor func() facade.Entity) (facade.Entity, error) {
	if e, err := m.Get(ctx, entityType, id); err == nil {
		return e, nil
	}
	return m.Create(ctx, entityType, id, ctor)
}
func (m *fakeMgr) Remove(ctx context.Context, entityType, id string) error {
	delete(m.ents[entityType], id)
	return nil
}
func (m *fakeMgr) DestroyAllType(ctx context.Context, id string) error { return nil }
func (m *fakeMgr) ReleaseAll(ctx context.Context) error                { return nil }
func (m *fakeMgr) IsAllLanded(ctx context.Context) (bool, error)       { return true, nil }
func (m *fakeMgr) RegisterNotFoundHook(entityType string, fn func(ctx context.Context, id string) (facade.Entity, error)) {
}
func (m *fakeMgr) RegisterCreateProcess(abilityName string, consumer func(ctx context.Context, e facade.Entity)) {
}
func (m *fakeMgr) RegisterAddProcess(abilityName string, consumer func(ctx context.Context, e facade.Entity)) {
	if m.addProcs == nil {
		m.addProcs = map[string][]func(ctx context.Context, e facade.Entity){}
	}
	m.addProcs[abilityName] = append(m.addProcs[abilityName], consumer)
}
func (m *fakeMgr) RegisterGetProcess(abilityName string, consumer func(ctx context.Context, e facade.Entity)) {
}
func (m *fakeMgr) RegisterRemoveProcess(abilityName string, consumer func(ctx context.Context, e facade.Entity)) {
}
func (m *fakeMgr) RegisterBeforeAddProcess(abilityName string, fun func(ctx context.Context, e facade.Entity) (facade.Ability, error)) {
}

func TestCallSystem_AddAbilityMethod(t *testing.T) {
	ctx := context.Background()
	// prepare entity and ability
	e := &testEntity{id: "1", typeName: "user"}
	adder := &adderAbility{}
	if err := adder.Attach(ctx, e); err != nil {
		t.Fatalf("attach adder: %v", err)
	}
	// register mapping for adder ability methods (by type)
	RegisterType(reflect.TypeOf((*adderAbility)(nil)))
	// prepare manager and register entity
	mgr := &fakeMgr{ents: map[string]map[string]facade.Entity{"user": {"1": e}}}
	// init call system
	cs := &CallSystemImpl{}
	cs.Init(ctx, mgr)
	// build request content (json for struct)
	arg := &addReq{A: 3, B: 5}
	b, _ := json.Marshal(arg)
	req := &entity.EntityRequest{Type: "user", Id: "1", FunName: "Add", Content: [][]byte{b}}
	// execute
	bytesList, err := cs.Call(ctx, "src", "Add", req)
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if len(bytesList) != 1 {
		t.Fatalf("unexpected return len: %d", len(bytesList))
	}
	var resp addResp
	if err := json.Unmarshal(bytesList[0], &resp); err != nil {
		t.Fatalf("decode resp: %v", err)
	}
	if resp.Sum != 8 {
		t.Fatalf("unexpected sum: %d", resp.Sum)
	}
}

func TestCallSystem_UnknownMethod(t *testing.T) {
	ctx := context.Background()
	e := &testEntity{id: "2", typeName: "user"}
	mgr := &fakeMgr{ents: map[string]map[string]facade.Entity{"user": {"2": e}}}
	cs := &CallSystemImpl{}
	cs.Init(ctx, mgr)
	req := &entity.EntityRequest{Type: "user", Id: "2", FunName: "NotExist", Content: nil}
	_, err := cs.Call(ctx, "src", "NotExist", req)
	if err == nil {
		t.Fatalf("expected error for unknown method")
	}
}

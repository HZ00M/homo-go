package call

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/go-kratos/kratos/v2/api/entity"
	facade "github.com/go-kratos/kratos/v2/entity/facade"
)

type orderReq struct {
	Seq int `json:"seq"`
}
type orderResp struct{}

type orderState struct {
	processed []int
}

type orderAbility struct {
	owner facade.Entity
	st    *orderState
}

func (a *orderAbility) Name() string { return "order" }
func (a *orderAbility) Attach(ctx context.Context, owner facade.Entity) error {
	a.owner = owner
	owner.AddAbility(a)
	return nil
}
func (a *orderAbility) Detach(ctx context.Context) error { a.owner = nil; return nil }
func (a *orderAbility) Step(ctx context.Context, req *orderReq) (*orderResp, error) {
	// record processing order
	a.st.processed = append(a.st.processed, req.Seq)
	return &orderResp{}, nil
}

func TestCallSystem_Order_SameEntity(t *testing.T) {
	ctx := context.Background()
	// entity with attached order ability
	e := &testEntity{id: "O1", typeName: "user"}
	st := &orderState{}
	ord := &orderAbility{st: st}
	if err := ord.Attach(ctx, e); err != nil {
		t.Fatalf("attach order ability: %v", err)
	}
	// register type so dispatcher knows method signatures
	RegisterType(reflect.TypeOf((*orderAbility)(nil)))
	// manager with the entity registered
	mgr := &fakeMgr{ents: map[string]map[string]facade.Entity{"user": {"O1": e}}}
	// init call system with larger queue size to avoid overflow
	cs := &CallSystemImpl{}
	cs.SetQueueSize(1024)
	cs.Init(ctx, mgr)

	// issue concurrent calls for same entity
	const N = 50
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		i := i
		go func() {
			defer wg.Done()
			b, _ := json.Marshal(&orderReq{Seq: i})
			req := &entity.EntityRequest{Type: "user", Id: "O1", FunName: "Step", Content: [][]byte{b}}
			_, err := cs.Call(ctx, "src", "Step", req)
			if err != nil {
				t.Errorf("call err: %v", err)
			}
		}()
	}
	wg.Wait()

	// verify order equals 0..N-1 strictly
	if len(st.processed) != N {
		t.Fatalf("processed len=%d want=%d", len(st.processed), N)
	}
	// ensure sequence is strictly increasing by 1
	for i := 0; i < N; i++ {
		if st.processed[i] != i {
			t.Fatalf("order mismatch at %d: got=%d want=%d (processed=%v)", i, st.processed[i], i, st.processed)
		}
	}
	// sanity: order differs from sorted check means any disorder would be caught
	sorted := append([]int(nil), st.processed...)
	sort.Ints(sorted)
	if !(len(sorted) == len(st.processed)) {
		t.Fatalf("sanity len mismatch")
	}
}

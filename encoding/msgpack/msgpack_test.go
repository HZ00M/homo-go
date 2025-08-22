package msgpack

import (
	"testing"

	"github.com/go-kratos/kratos/v2/encoding"
)

func TestMsgpackCodec_RegisterAndRoundtrip(t *testing.T) {
	c := encoding.GetCodec(Name)
	if c == nil || c.Name() != Name {
		t.Fatalf("codec not registered: %v", c)
	}
	type demo struct {
		A int
		B string
	}
	in := demo{A: 42, B: "ok"}
	b, err := c.Marshal(in)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var out demo
	if err := c.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if out != in {
		t.Fatalf("roundtrip mismatch: %+v != %+v", out, in)
	}
}

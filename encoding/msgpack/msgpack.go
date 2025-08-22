package msgpack

import (
	mp "github.com/vmihailenco/msgpack/v5"

	"github.com/go-kratos/kratos/v2/encoding"
)

// Name is the name registered for the msgpack codec.
const Name = "msgpack"

func init() {
	encoding.RegisterCodec(codec{})
}

// codec is a Codec implementation with msgpack.
type codec struct{}

// Marshal encodes v into msgpack bytes.
func (codec) Marshal(v any) ([]byte, error) { return mp.Marshal(v) }

// Unmarshal decodes msgpack bytes into v.
func (codec) Unmarshal(data []byte, v any) error { return mp.Unmarshal(data, v) }

// Name returns the codec name.
func (codec) Name() string { return Name }

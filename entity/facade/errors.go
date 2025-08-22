package facade

import "errors"

// 标准错误集合（与文档对齐）
var (
	ErrNotFound          = errors.New("entity: not found")
	ErrAlreadyInOtherPod = errors.New("entity: already in other pod")
	ErrNoAvailablePod    = errors.New("entity: no available pod")
	ErrMethodNotFound    = errors.New("entity: method not found")
	ErrAbilityNotFound   = errors.New("entity: ability not found")
	ErrDecode            = errors.New("entity: decode error")
	ErrEncode            = errors.New("entity: encode error")
	ErrTimeout           = errors.New("entity: timeout")
	ErrQueueOverflow     = errors.New("entity: queue overflow")
)

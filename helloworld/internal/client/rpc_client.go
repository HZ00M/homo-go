package client

import (
	"bytes"
	"context"
	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	v1 "helloworld/api/helloworld/v1"
	"sync"
	"time"
)

type MsgContent struct {
	content any
	encoding.Codec
}
type RpcProxy struct {
	conn     *grpc.ClientConn
	client   v1.RouterServiceClient
	stream   v1.RouterService_StreamCallClient // 有状态服务保留 Stream
	stateful bool
	mu       sync.Mutex
	service  string // 服务名或唯一标识
	reqIdGen func() string
}

func NewRpcProxy(ctx context.Context, target string, opts ...grpc.DialOption) (*RpcProxy, error) {
	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, err
	}
	stateful := true
	client := v1.NewRouterServiceClient(conn)
	rc := &RpcProxy{
		conn:     conn,
		client:   client,
		stateful: stateful,
		reqIdGen: func() string { return uuid.New().String() },
	}

	// 如果是有状态服务，建立 stream
	if stateful {
		stream, err := client.StreamCall(ctx)
		if err != nil {
			conn.Close()
			return nil, err
		}
		rc.stream = stream
	}

	return rc, nil
}

func (c *RpcProxy) Call(ctx context.Context, msgId string, msgContent MsgContent) (any, error) {
	if c.stateful {
		// 有状态 stream 双向通信
		c.mu.Lock()
		defer c.mu.Unlock()

		data, ok := msgContent.content.([][]byte)
		if !ok {
			return nil, errors.New(-1, " msgContent.content.([][]byte)", "expect [][]byte for stream request")
		}
		reqId := c.reqIdGen()
		trace := &v1.TraceInfo{Traceid: time.Now().UnixNano(), Spanid: 1, Sample: true}
		err := c.stream.Send(&v1.StreamReq{
			SrcService: c.service,
			MsgId:      msgId,
			MsgContent: data,
			ReqId:      reqId,
			Traceinfo:  trace,
		})
		if err != nil {
			return nil, err
		}

		// 简化处理：等待响应
		for {
			rsp, err := c.stream.Recv()
			if err != nil {
				return nil, err
			}
			if rsp.ReqId == reqId {
				return bytes.Join(rsp.MsgContent, []byte{}), nil
			}
		}
	}

	switch msgContent.Codec.Name() {
	case "json":
		data := msgContent.content
		jsonData, err := msgContent.Codec.Marshal(data)
		if err != nil {
			return nil, err
		}
		rsp, err := c.client.JsonMessage(ctx, &v1.JsonReq{
			SrcService: c.service,
			MsgId:      msgId,
			MsgContent: string(jsonData),
		})
		if err != nil {
			return nil, err
		}
		return []byte(rsp.MsgContent), nil

	case "bytes":
		data, ok := msgContent.content.([][]byte)
		if !ok {
			return nil, errors.New(-2, " msgContent.content.([][]byte)", "expect [][]byte for stream request")
		}
		rsp, err := c.client.RpcCall(ctx, &v1.Req{
			SrcService: c.service,
			MsgId:      msgId,
			MsgContent: data,
		})
		if err != nil {
			return nil, err
		}
		return bytes.Join(rsp.MsgContent, []byte{}), nil

	default:
		return nil, errors.New(-3, "unknown", "fail")
	}
}

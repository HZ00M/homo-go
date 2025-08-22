package rpc

import (
	"context"

	"github.com/go-kratos/kratos/v2/api/rpc"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RouterServiceServerImpl RouterService的gRPC服务器实现
type RouterServiceServerImpl struct {
	rpc.UnimplementedRouterServiceServer
	rpcServer RpcServerDriver
}

// NewRouterServiceServer 创建RouterService服务器
func NewRouterServiceServer(rpcServer RpcServerDriver) rpc.RouterServiceServer {
	return &RouterServiceServerImpl{
		rpcServer: rpcServer,
	}
}

// JsonMessage 处理JSON消息
func (s *RouterServiceServerImpl) JsonMessage(ctx context.Context, req *rpc.JsonReq) (*rpc.JsonRsp, error) {
	log.Infof("Received JSON message from %s, msgId: %s", req.SrcService, req.MsgId)

	// 创建RPC内容
	content := &Content[string]{
		CType: RpcContentJson,
		Dt:    req.MsgContent,
	}

	// 调用RPC服务器
	result, err := s.rpcServer.OnCall(ctx, req.SrcService, req.MsgId, content)
	if err != nil {
		log.Errorf("RPC call failed: %v", err)
		return nil, status.Errorf(codes.Internal, "RPC call failed: %v", err)
	}

	// 处理返回结果
	if result == nil {
		return &rpc.JsonRsp{
			MsgId:      req.MsgId,
			MsgContent: "",
		}, nil
	}

	if result.Type() != RpcContentJson {
		return nil, status.Errorf(codes.Internal, "unexpected response type: %v", result.Type())
	}

	responseData, ok := result.Data().(string)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid response data type")
	}

	return &rpc.JsonRsp{
		MsgId:      req.MsgId,
		MsgContent: responseData,
	}, nil
}

// RpcCall 处理RPC调用
func (s *RouterServiceServerImpl) RpcCall(ctx context.Context, req *rpc.Req) (*rpc.Rsp, error) {
	log.Infof("Received RPC call from %s, msgId: %s", req.SrcService, req.MsgId)

	// 创建RPC内容
	content := &Content[[][]byte]{
		CType: RpcContentBytes,
		Dt:    req.MsgContent,
	}

	// 调用RPC服务器
	result, err := s.rpcServer.OnCall(ctx, req.SrcService, req.MsgId, content)
	if err != nil {
		log.Errorf("RPC call failed: %v", err)
		return nil, status.Errorf(codes.Internal, "RPC call failed: %v", err)
	}

	// 处理返回结果
	if result == nil {
		return &rpc.Rsp{
			MsgId:      req.MsgId,
			MsgContent: [][]byte{},
		}, nil
	}

	if result.Type() != RpcContentBytes {
		return nil, status.Errorf(codes.Internal, "unexpected response type: %v", result.Type())
	}

	responseData, ok := result.Data().([][]byte)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid response data type")
	}

	return &rpc.Rsp{
		MsgId:      req.MsgId,
		MsgContent: responseData,
	}, nil
}

// StreamCall 处理流式调用
func (s *RouterServiceServerImpl) StreamCall(stream rpc.RouterService_StreamCallServer) error {
	log.Info("Stream call started")

	// 处理流式请求
	for {
		req, err := stream.Recv()
		if err != nil {
			log.Errorf("Stream receive error: %v", err)
			return err
		}

		log.Infof("Received stream request from %s, msgId: %s, reqId: %s",
			req.SrcService, req.MsgId, req.ReqId)

		// 创建RPC内容
		content := &Content[[][]byte]{
			CType: RpcContentBytes,
			Dt:    req.MsgContent,
		}

		// 调用RPC服务器
		result, err := s.rpcServer.OnCall(stream.Context(), req.SrcService, req.MsgId, content)
		if err != nil {
			log.Errorf("Stream RPC call failed: %v", err)
			// 发送错误响应
			errRsp := &rpc.StreamRsp{
				MsgId:      req.MsgId,
				MsgContent: [][]byte{},
				ReqId:      req.ReqId,
				Spnid:      req.Traceinfo.Spanid,
			}
			if sendErr := stream.Send(errRsp); sendErr != nil {
				log.Errorf("Failed to send error response: %v", sendErr)
			}
			continue
		}

		// 处理返回结果
		var responseData [][]byte
		if result != nil && result.Type() == RpcContentBytes {
			if data, ok := result.Data().([][]byte); ok {
				responseData = data
			}
		}

		// 发送响应
		rsp := &rpc.StreamRsp{
			MsgId:      req.MsgId,
			MsgContent: responseData,
			ReqId:      req.ReqId,
			Spnid:      req.Traceinfo.Spanid,
		}

		if err := stream.Send(rsp); err != nil {
			log.Errorf("Failed to send stream response: %v", err)
			return err
		}
	}
}

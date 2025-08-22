package service

import (
	"context"
	"io"

	pb "helloworld/api/helloworld/v1"
)

type RouterServiceService struct {
	pb.UnimplementedRouterServiceServer
}

func NewRouterServiceService() *RouterServiceService {
	return &RouterServiceService{}
}

func (s *RouterServiceService) JsonMessage(ctx context.Context, req *pb.JsonReq) (*pb.JsonRsp, error) {
	return &pb.JsonRsp{}, nil
}
func (s *RouterServiceService) RpcCall(ctx context.Context, req *pb.Req) (*pb.Rsp, error) {
	return &pb.Rsp{}, nil
}
func (s *RouterServiceService) StreamCall(conn pb.RouterService_StreamCallServer) error {
	for {
		_, err := conn.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		err = conn.Send(&pb.StreamRsp{})
		if err != nil {
			return err
		}
	}
}

package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"

	pb "helloworld/api/helloworld/v1"
)

type Demo2Service struct {
	pb.UnimplementedDemo2Server
	log *log.Helper
}

func NewDemo2Service() *Demo2Service {
	return &Demo2Service{
		log: log.NewHelper(log.DefaultLogger),
	}
}

func (s *Demo2Service) Hi(ctx context.Context, req *pb.Demo2Request) (*pb.Demo2Reply, error) {
	s.log.Infof("Demo2Service.Hi: %v", req)
	return &pb.Demo2Reply{Message: req.Name}, nil
}
func (s *Demo2Service) Hello(ctx context.Context, req *pb.HelloReq) (*pb.HelloRep, error) {
	s.log.Infof("Demo2Service.Hello: %v", req)
	return &pb.HelloRep{
		Message: "Hello, " + req.Name,
		Code:    0,
	}, nil
}

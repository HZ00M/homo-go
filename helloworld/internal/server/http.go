package server

import (
	"context"
	"encoding/json"
	"github.com/go-kratos/kratos/v2/api/metadata"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/protobuf/encoding/protojson"
	v1 "helloworld/api/helloworld/v1"
	"helloworld/internal/conf"
	"helloworld/internal/service"

	stdhttp "net/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server,
	grpc *grpc.Server,
	greeter *service.GreeterService,
	demo *service.DemoService,
	demo2 *service.Demo2Service,
	router *service.RouterServiceService,
	logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			//middleware2.CORS(),  //时序问题 不可用 在mux.router匹配阶段已经被拦截了
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	httpServer := http.NewServer(opts...)
	//RegisterCorsHandler(httpServer) //必须最先注册
	RegisterCorsHandlerWithFallback(httpServer)
	v1.RegisterGreeterHTTPServer(httpServer, greeter)
	v1.RegisterDemo2HTTPServer(httpServer, demo2)
	v1.RegisterRouterServiceHTTPServer(httpServer, router)
	RegisterMetadataListServiceHTTP(httpServer, grpc.Metadata)
	RegisterMetadataDescHTTP(httpServer, grpc.Metadata)
	//RegisterMetadataHTTPServer(httpServer, grpc.Metadata)
	return httpServer
}

// RegisterCorsHandlerWithFallback 注册兜底 OPTIONS Handler
func RegisterCorsHandlerWithFallback(s *http.Server) {
	router := s.GetMuxRouter()

	// OPTIONS 请求未匹配路由时兜底
	router.NotFoundHandler = stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if r.Method == stdhttp.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.WriteHeader(stdhttp.StatusNoContent)
			return
		}
		stdhttp.NotFound(w, r)
	})

	// 方法不允许的请求（例如 POST 到 GET 接口）
	router.MethodNotAllowedHandler = stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if r.Method == stdhttp.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.WriteHeader(stdhttp.StatusNoContent)
			return
		}
		stdhttp.Error(w, "Method Not Allowed", stdhttp.StatusMethodNotAllowed)
	})
}

//	func RegisterCorsHandler(s *http.Server) {
//		s.HandleFunc("/{any:.*}", func(w http.ResponseWriter, r *http.Request) {
//			if r.Method == np.MethodOptions {
//				w.Header().Set("Access-Control-Allow-Origin", "*")
//				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
//				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Token")
//				w.Header().Set("Access-Control-Allow-Credentials", "true")
//				w.WriteHeader(np.StatusNoContent)
//				return
//			}
//		})
//	}
func Metadata_Services_Handler(m *metadata.Server) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in metadata.ListServicesRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, metadata.Metadata_ListServices_FullMethodName)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return m.ListServices(ctx, req.(*metadata.ListServicesRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*metadata.ListServicesReply)
		return ctx.Result(200, reply)
	}
}
func Metadata_Desc_Handler(m *metadata.Server) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in metadata.GetServiceDescRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, metadata.Metadata_GetServiceDesc_FullMethodName)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return m.GetServiceDesc(ctx, req.(*metadata.GetServiceDescRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*metadata.GetServiceDescReply)
		return ctx.Result(200, reply)
	}
}
func RegisterMetadataListServiceHTTP(s *http.Server, srv *metadata.Server) {
	r := s.Route("/")
	r.GET("/metadata/services", Metadata_Services_Handler(srv))
}
func RegisterMetadataDescHTTP(s *http.Server, srv *metadata.Server) {
	r := s.Route("/")
	r.GET("/metadata/desc", Metadata_Desc_Handler(srv))
}
func RegisterMetadataHTTPServer(s *http.Server, m *metadata.Server) {
	s.HandleFunc("/metadata/desc", func(w http.ResponseWriter, r *http.Request) {
		reply, err := m.ListServices(r.Context(), nil)
		if err != nil {
			//http.Error(w, err.Error(), 500)
			return
		}
		_ = json.NewEncoder(w).Encode(reply)
		marshaler := protojson.MarshalOptions{Multiline: true}
		b, _ := marshaler.Marshal(reply)
		log.Info("metadata/services", "reply", string(b))
	})

	s.HandleFunc("/metadata/desc", func(w http.ResponseWriter, r *http.Request) {
		service := r.URL.Query().Get("name")
		if service == "" {
			//http.Error(w, "missing ?name=service.name", 400)
			return
		}
		reply, err := m.GetServiceDesc(r.Context(), &metadata.GetServiceDescRequest{Name: service})
		if err != nil {
			//http.Error(w, err.Error(), 500)
			return
		}
		// 使用 protojson 输出 descriptor
		marshaler := protojson.MarshalOptions{Multiline: true}
		b, _ := marshaler.Marshal(reply)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		log.Info("metadata/desc", "reply", string(b))
	})
}

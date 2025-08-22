package middleware

import (
	"context"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"strings"
)

func CORS() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				if httpTr, ok := tr.(*http.Transport); ok {
					header := httpTr.ReplyHeader()
					header.Set("Access-Control-Allow-Origin", "*")
					header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					header.Set("Access-Control-Expose-Headers", "*")
					header.Set("Access-Control-Allow-Credentials", "true")

					// 处理预检请求（OPTIONS）
					if strings.ToUpper(httpTr.Request().Method) == "OPTIONS" {
						// 直接返回响应，不再进入后续 handler
						return nil, nil
					}
				}
			}
			return next(ctx, req)
		}
	}
}

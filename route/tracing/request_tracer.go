package tracing

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RequestTracer 请求链路追踪器
type RequestTracer struct {
	tracer trace.Tracer
	logger log.Logger
}

// TraceInfo 追踪信息
type TraceInfo struct {
	TraceID    string
	SpanID     string
	ParentID   string
	Operation  string
	StartTime  time.Time
	EndTime    time.Time
	Attributes map[string]string
}

// NewRequestTracer 创建新的请求追踪器
func NewRequestTracer(logger log.Logger) *RequestTracer {
	return &RequestTracer{
		tracer: otel.Tracer("stateful-route"),
		logger: logger,
	}
}

// StartSpan 开始追踪span
func (rt *RequestTracer) StartSpan(ctx context.Context, operation string, attributes map[string]string) (context.Context, trace.Span) {
	// 转换属性
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	// 创建span
	ctx, span := rt.tracer.Start(ctx, operation, trace.WithAttributes(attrs...))

	// 记录开始时间
	span.SetAttributes(attribute.Int64("start_time", time.Now().UnixNano()))

	return ctx, span
}

// EndSpan 结束追踪span
func (rt *RequestTracer) EndSpan(span trace.Span, err error) {
	if err != nil {
		span.SetStatus(trace.StatusError, err.Error())
		span.RecordError(err)
	} else {
		span.SetStatus(trace.StatusOK, "")
	}

	// 记录结束时间
	span.SetAttributes(attribute.Int64("end_time", time.Now().UnixNano()))

	span.End()
}

// AddEvent 添加事件
func (rt *RequestTracer) AddEvent(span trace.Span, name string, attributes map[string]string) {
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		attrs = append(attrs, attribute.String(k, v))
	}

	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// TraceRequest 追踪请求
func (rt *RequestTracer) TraceRequest(ctx context.Context, operation string, fn func(context.Context) error) error {
	ctx, span := rt.StartSpan(ctx, operation, nil)
	defer rt.EndSpan(span, nil)

	startTime := time.Now()
	err := fn(ctx)
	duration := time.Since(startTime)

	// 记录请求持续时间
	span.SetAttributes(attribute.Duration("duration", duration))

	if err != nil {
		rt.EndSpan(span, err)
	}

	return err
}

// GetTraceInfo 获取追踪信息
func (rt *RequestTracer) GetTraceInfo(ctx context.Context) *TraceInfo {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return nil
	}

	spanCtx := span.SpanContext()

	return &TraceInfo{
		TraceID:    spanCtx.TraceID().String(),
		SpanID:     spanCtx.SpanID().String(),
		ParentID:   "", // 需要从上下文获取
		Operation:  span.Name(),
		StartTime:  time.Now(), // 需要从span获取
		EndTime:    time.Now(), // 需要从span获取
		Attributes: make(map[string]string),
	}
}

// AddSpanAttribute 添加span属性
func (rt *RequestTracer) AddSpanAttribute(span trace.Span, key, value string) {
	span.SetAttributes(attribute.String(key, value))
}

// AddSpanAttributes 批量添加span属性
func (rt *RequestTracer) AddSpanAttributes(span trace.Span, attributes map[string]string) {
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		attrs = append(attrs, attribute.String(k, v))
	}
	span.SetAttributes(attrs...)
}

// CreateChildSpan 创建子span
func (rt *RequestTracer) CreateChildSpan(ctx context.Context, operation string, attributes map[string]string) (context.Context, trace.Span) {
	return rt.StartSpan(ctx, operation, attributes)
}

// RecordError 记录错误
func (rt *RequestTracer) RecordError(span trace.Span, err error, attributes map[string]string) {
	if attributes != nil {
		rt.AddSpanAttributes(span, attributes)
	}
	span.RecordError(err)
	span.SetStatus(trace.StatusError, err.Error())
}

// SetSpanStatus 设置span状态
func (rt *RequestTracer) SetSpanStatus(span trace.Span, code trace.StatusCode, description string) {
	span.SetStatus(code, description)
}

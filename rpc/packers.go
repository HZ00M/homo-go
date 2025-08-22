package rpc

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/go-kratos/kratos/v2/encoding"

	"github.com/go-kratos/kratos/v2/errors"
	"google.golang.org/protobuf/proto"
)

// BytesPacker Proto类型的打包器
type BytesPacker struct{}

func (p *BytesPacker) Name() string {
	return "bytes"
}

func (p *BytesPacker) Match(fun *FunInfo) bool {
	// 只要不是“全部 JSON 结构体参数 + 返回”，其余情况均由 bytes 处理
	paramTypes := fun.ParamTypes()
	returnTypes := fun.ReturnTypes()
	if isAllJson(paramTypes) && isAllJsonReturn(returnTypes) {
		return false
	}
	return len(paramTypes) > 0 || len(returnTypes) > 0
}

func isAllProto(types []reflect.Type) bool {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	protoMessageType := reflect.TypeOf((*proto.Message)(nil)).Elem()

	for _, t := range types {
		// 跳过 error 类型的返回值
		if t.Implements(errorType) {
			continue
		}

		// 检查指针类型是否实现了proto.Message接口
		if t.Kind() == reflect.Ptr {
			if !t.Implements(protoMessageType) {
				return false
			}
		} else {
			// 如果不是指针类型，检查其指针类型是否实现了proto.Message接口
			ptrType := reflect.PtrTo(t)
			if !ptrType.Implements(protoMessageType) {
				return false
			}
		}
	}
	return len(types) > 0 // 至少有一个参数
}

func (p *BytesPacker) UnSerialize(content RpcContent, fun *FunInfo) ([]interface{}, error) {
	if content.Type() != RpcContentBytes {
		return nil, errors.New(400, "INVALID_CONTENT_TYPE", "expect bytes content type for bytes packer")
	}
	data, ok := content.Data().([][]byte)
	if !ok {
		return nil, errors.New(400, "INVALID_DATA_TYPE", "expect [][]byte data for bytes content")
	}
	args := make([]interface{}, len(fun.Param))

	for i, paramInfo := range fun.Param {
		t := paramInfo.Type

		if t.Kind() == reflect.Interface {
			return nil, errors.New(400, "INVALID_PARAM_TYPE", "interface type param not supported without factory")
		}

		if i < len(data) {
			// 有数据，尝试反序列化
			if t.Kind() == reflect.Ptr {
				// 指针类型，创建指针实例
				paramInstance := reflect.New(t.Elem()).Interface()
				if err := encoding.GetCodec(paramInfo.Codec).Unmarshal(data[i], paramInstance); err != nil {
					return nil, err
				}
				args[i] = paramInstance // 直接使用指针实例
			} else {
				// 值类型，创建值实例
				paramInstance := reflect.New(t).Interface()
				if err := encoding.GetCodec(paramInfo.Codec).Unmarshal(data[i], paramInstance); err != nil {
					return nil, err
				}
				args[i] = reflect.ValueOf(paramInstance).Elem().Interface() // 获取值
			}
		} else {
			// 没有数据，设置零值
			args[i] = reflect.Zero(t).Interface()
		}
	}
	return args, nil
}

func (p *BytesPacker) Serialize(fun *FunInfo, returnValues []any) (RpcContent, error) {
	result := make([][]byte, 0, len(returnValues))
	for i, ret := range fun.ReturnParam {
		// 跳过 error
		if ret.Type.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			continue
		}
		if i >= len(returnValues) {
			continue
		}
		bytes, err := encoding.GetCodec(ret.Codec).Marshal(returnValues[i])
		if err != nil {
			return nil, err
		}
		result = append(result, bytes)
	}
	return &Content[[][]byte]{CType: RpcContentBytes, Dt: result}, nil
}

func (p *BytesPacker) Call(ctx context.Context, req RpcContent, fun *FunInfo) (RpcContent, error) {
	args, err := p.UnSerialize(req, fun)
	if err != nil {
		return nil, err
	}
	callArgs := make([]reflect.Value, 0, len(args)+1)
	mt := fun.Method.Type()
	if mt.NumIn() > 0 && mt.In(0) == reflect.TypeOf((*context.Context)(nil)).Elem() {
		callArgs = append(callArgs, reflect.ValueOf(ctx))
	}
	callArgs = append(callArgs, sliceToValues(args)...)
	results := fun.Method.Call(callArgs)

	for _, result := range results {
		if result.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !result.IsNil() {
				return nil, result.Interface().(error)
			}
		}
	}

	return p.Serialize(fun, sliceToInterfaces(results))
}

// JsonPacker JSON打包器
type JsonPacker struct{}

func (p *JsonPacker) Name() string {
	return "json"
}

func (p *JsonPacker) Match(fun *FunInfo) bool {
	paramTypes := fun.ParamTypes()
	returnTypes := fun.ReturnTypes()

	paramMatch := isAllJson(paramTypes)
	returnMatch := isAllJsonReturn(returnTypes)

	return paramMatch && returnMatch
}

func isAllJson(types []reflect.Type) bool {
	// 如果没有参数，返回false
	if len(types) == 0 {
		return false
	}

	for _, t := range types {
		u := t
		if u.Kind() == reflect.Ptr {
			u = u.Elem()
		}
		if u.Kind() != reflect.Struct || u.Name() == "" {
			return false
		}
	}
	return true
}

func isAllJsonReturn(types []reflect.Type) bool {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	for _, t := range types {
		if t.Implements(errorType) {
			continue
		}
		u := t
		if u.Kind() == reflect.Ptr {
			u = u.Elem()
		}
		if u.Kind() != reflect.Struct || u.Name() == "" {
			return false
		}
	}
	return true
}

func (p *JsonPacker) UnSerialize(content RpcContent, fun *FunInfo) ([]interface{}, error) {
	if content.Type() != RpcContentJson {
		return nil, errors.New(400, "INVALID_CONTENT_TYPE", "expect JSON content type")
	}
	data, ok := content.Data().(string)
	if !ok {
		return nil, errors.New(400, "INVALID_DATA_TYPE", "expect string data for JSON content")
	}

	// 直接解析JSON字符串到对应的结构体
	args := make([]interface{}, 0, len(fun.Param))
	for _, paramInfo := range fun.Param {
		paramInstance := reflect.New(paramInfo.Type).Interface()
		if err := json.Unmarshal([]byte(data), paramInstance); err != nil {
			return nil, err
		}
		args = append(args, reflect.ValueOf(paramInstance).Elem().Interface())
	}
	return args, nil
}

func (p *JsonPacker) Serialize(fun *FunInfo, returnValues []any) (RpcContent, error) {
	nonErrorValues := make([]any, 0, len(returnValues))
	returnTypes := fun.ReturnTypes()
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	for i, returnType := range returnTypes {
		if returnType.Implements(errorType) {
			continue
		}
		if i < len(returnValues) {
			nonErrorValues = append(nonErrorValues, returnValues[i])
		}
	}

	// 如果有多个返回值，使用第一个非error的返回值
	if len(nonErrorValues) > 0 {
		jsonBytes, err := json.Marshal(nonErrorValues[0])
		if err != nil {
			return nil, err
		}
		return &Content[string]{CType: RpcContentJson, Dt: string(jsonBytes)}, nil
	}

	// 如果没有返回值，返回空JSON对象
	return &Content[string]{CType: RpcContentJson, Dt: "{}"}, nil
}

func (p *JsonPacker) Call(ctx context.Context, req RpcContent, fun *FunInfo) (RpcContent, error) {
	args, err := p.UnSerialize(req, fun)
	if err != nil {
		return nil, err
	}
	callArgs := make([]reflect.Value, 0, len(args)+1)
	mt := fun.Method.Type()
	if mt.NumIn() > 0 && mt.In(0) == reflect.TypeOf((*context.Context)(nil)).Elem() {
		callArgs = append(callArgs, reflect.ValueOf(ctx))
	}
	callArgs = append(callArgs, sliceToValues(args)...)
	results := fun.Method.Call(callArgs)

	for _, result := range results {
		if result.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !result.IsNil() {
				return nil, result.Interface().(error)
			}
		}
	}

	return p.Serialize(fun, sliceToInterfaces(results))
}

// 工具函数
func sliceToInterfaces(values []reflect.Value) []interface{} {
	results := make([]interface{}, len(values))
	for i, v := range values {
		if v.IsValid() {
			results[i] = v.Interface()
		} else {
			results[i] = nil
		}
	}
	return results
}

func sliceToValues(args []interface{}) []reflect.Value {
	values := make([]reflect.Value, len(args))
	for i, arg := range args {
		values[i] = reflect.ValueOf(arg)
	}
	return values
}

// 初始化函数，注册所有 packer
func init() {
	RegisterPacker(&BytesPacker{})
	RegisterPacker(&JsonPacker{})
}

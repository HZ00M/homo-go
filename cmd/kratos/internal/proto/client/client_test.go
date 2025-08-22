package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_NoArgs(t *testing.T) {
	// 捕获标准输出
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	run(nil, []string{})

	_ = w.Close()
	os.Stdout = old

	output := make([]byte, 1024)
	n, _ := r.Read(output)
	assert.Contains(t, string(output[:n]), "Please enter the proto file or directory")
}

func TestRun_BasicProto(t *testing.T) {
	tmpDir := t.TempDir()

	protoFile := filepath.Join(tmpDir, "simple.proto")
	err := os.WriteFile(protoFile, []byte(`
syntax = "proto3";
package test;
service Hello {
  rpc Say (Request) returns (Reply);
}
message Request {}
message Reply {}
`), 0o644)
	require.NoError(t, err)

	// 切换当前目录为 proto 文件所在目录
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// 直接调用 run，传入参数为 proto 文件路径
	run(nil, []string{protoFile})
}

func TestPathExists(t *testing.T) {
	tmpDir := t.TempDir()
	exists := pathExists(tmpDir)
	assert.True(t, exists)

	notExists := pathExists(filepath.Join(tmpDir, "not-exist"))
	assert.False(t, notExists)
}

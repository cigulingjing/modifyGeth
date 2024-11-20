package merkeldag

import (
	"bytes"
	"testing"
)

func TestNewNode(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid data",
			data:    []byte("test data"),
			wantErr: false,
		},
		{
			name:    "nil data",
			data:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := NewNode(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && node == nil {
				t.Error("NewNode() returned nil node for valid data")
			}
		})
	}
}

func TestNode_calculateHash(t *testing.T) {
	// 创建一个简单的树结构
	root, _ := NewNode([]byte("root"))
	left, _ := NewNode([]byte("left"))
	right, _ := NewNode([]byte("right"))
	root.SetLeft(left)
	root.SetRight(right)

	// 测试叶子节点哈希
	leftHash := left.calculateHash()
	if len(leftHash) == 0 {
		t.Error("calculateHash() failed for leaf node")
	}

	// 测试父节点哈希
	rootHash := root.calculateHash()
	if len(rootHash) == 0 {
		t.Error("calculateHash() failed for parent node")
	}

	// 确保父节点哈希不同于子节点
	if bytes.Equal(rootHash, leftHash) {
		t.Error("parent and child nodes should have different hashes")
	}
}

func TestInsertNode(t *testing.T) {
	tests := []struct {
		name    string
		index   uint64
		data    []byte
		wantErr bool
	}{
		{
			name:    "insert at index 0",
			index:   0,
			data:    []byte("data0"),
			wantErr: false,
		},
		{
			name:    "insert at index 1",
			index:   1,
			data:    []byte("data1"),
			wantErr: false,
		},
		{
			name:    "insert with nil data",
			index:   2,
			data:    nil,
			wantErr: true,
		},
	}

	var root *Node
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newRoot, err := InsertNode(root, tt.data, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("InsertNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				root = newRoot
				if root == nil {
					t.Error("InsertNode() returned nil root for valid data")
				}
			}
		})
	}
}

func TestNode_Verify(t *testing.T) {
	// 创建一个有效的树
	root, _ := NewNode([]byte("root"))
	left, _ := NewNode([]byte("left"))
	right, _ := NewNode([]byte("right"))
	root.SetLeft(left)
	root.SetRight(right)

	if !root.Verify() {
		t.Error("Verify() failed for valid tree")
	}

	// 手动破坏哈希值
	left.hash = []byte("invalid hash")
	if root.Verify() {
		t.Error("Verify() should fail for invalid hash")
	}
}

package merkeldag

import (
	"testing"
)

func TestMerkelDAG_Insert(t *testing.T) {
	dag, err := NewMerkelDAG()
	if err != nil {
		t.Fatalf("Failed to create DAG: %v", err)
	}

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "first insert",
			data:    []byte("data1"),
			wantErr: false,
		},
		{
			name:    "second insert",
			data:    []byte("data2"),
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
			err := dag.Insert(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Insert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestMerkelDAG_GetCurrentIndex(t *testing.T) {
	dag, err := NewMerkelDAG()
	if err != nil {
		t.Fatalf("Failed to create DAG: %v", err)
	}

	// 初始索引应该是0
	if idx := dag.GetCurrentIndex(); idx != 0 {
		t.Errorf("Initial index should be 0, got %d", idx)
	}

	// 插入一些数据
	dag.Insert([]byte("data1"))
	if idx := dag.GetCurrentIndex(); idx != 1 {
		t.Errorf("After one insert, index should be 1, got %d", idx)
	}

	dag.Insert([]byte("data2"))
	if idx := dag.GetCurrentIndex(); idx != 2 {
		t.Errorf("After two inserts, index should be 2, got %d", idx)
	}
}

func TestMerkelDAG_SetCurrentIndex(t *testing.T) {
	dag, err := NewMerkelDAG()
	if err != nil {
		t.Fatalf("Failed to create DAG: %v", err)
	}

	tests := []struct {
		name    string
		index   uint64
		wantErr bool
	}{
		{
			name:    "valid index",
			index:   5,
			wantErr: false,
		},
		{
			name:    "max index",
			index:   MaxLeafNodes - 1,
			wantErr: false,
		},
		{
			name:    "overflow index",
			index:   MaxLeafNodes,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dag.SetCurrentIndex(tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetCurrentIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && dag.GetCurrentIndex() != tt.index {
				t.Errorf("SetCurrentIndex() = %v, want %v", dag.GetCurrentIndex(), tt.index)
			}
		})
	}
}

func TestMerkelDAG_IsEmpty(t *testing.T) {
	dag, err := NewMerkelDAG()
	if err != nil {
		t.Fatalf("Failed to create DAG: %v", err)
	}

	// 新创建的DAG应该是空的
	if !dag.IsEmpty() {
		t.Error("New DAG should be empty")
	}

	// 插入数据后不应该是空的
	dag.Insert([]byte("data"))
	if dag.IsEmpty() {
		t.Error("DAG should not be empty after insert")
	}
}

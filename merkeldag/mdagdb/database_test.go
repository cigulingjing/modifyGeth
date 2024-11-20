package mdagdb

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/merkeldag"
)

func TestMerkleDAGDB_SaveAndLoadDAG(t *testing.T) {
	// 创建内存数据库
	memDB := rawdb.NewMemoryDatabase()
	dagDB := NewMerkleDAGDB(memDB)

	// 创建一个测试 DAG
	dag := merkeldag.NewMerkelDAG(memDB)

	// 插入一些测试数据
	testData := [][]byte{
		[]byte("data1"),
		[]byte("data2"),
		[]byte("data3"),
	}

	for _, data := range testData {
		if err := dag.Insert(data); err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}
	}

	// 保存 DAG
	if err := dagDB.SaveDAG(dag); err != nil {
		t.Fatalf("Failed to save DAG: %v", err)
	}

	// 加载 DAG
	loadedDAG, err := dagDB.LoadDAG()
	if err != nil {
		t.Fatalf("Failed to load DAG: %v", err)
	}

	// 验证加载的 DAG
	if loadedDAG.GetCurrentIndex() != dag.GetCurrentIndex() {
		t.Errorf("Loaded DAG index mismatch: got %d, want %d",
			loadedDAG.GetCurrentIndex(), dag.GetCurrentIndex())
	}

	// 验证根节点
	if !bytes.Equal(loadedDAG.GetRoot().GetHash(), dag.GetRoot().GetHash()) {
		t.Error("Root hash mismatch between original and loaded DAG")
	}
}

func TestMerkleDAGDB_SaveAndLoadEmptyDAG(t *testing.T) {
	memDB := rawdb.NewMemoryDatabase()
	dagDB := NewMerkleDAGDB(memDB)
	emptyDAG := merkeldag.NewMerkelDAG(memDB)

	// 保存空 DAG
	if err := dagDB.SaveDAG(emptyDAG); err != nil {
		t.Fatalf("Failed to save empty DAG: %v", err)
	}

	// 加载空 DAG
	loadedDAG, err := dagDB.LoadDAG()
	if err != nil {
		t.Fatalf("Failed to load empty DAG: %v", err)
	}

	// 验证空 DAG
	if !loadedDAG.IsEmpty() {
		t.Error("Loaded DAG should be empty")
	}
}

func TestMerkleDAGDB_CacheOperations(t *testing.T) {
	memDB := rawdb.NewMemoryDatabase()
	dagDB := NewMerkleDAGDB(memDB)
	dag := merkeldag.NewMerkelDAG(memDB)

	// 插入测试数据
	if err := dag.Insert([]byte("test data")); err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// 保存 DAG
	if err := dagDB.SaveDAG(dag); err != nil {
		t.Fatalf("Failed to save DAG: %v", err)
	}

	// 清除缓存
	dagDB.ClearCache()

	// 重新加载，验证是否正确加载
	loadedDAG, err := dagDB.LoadDAG()
	if err != nil {
		t.Fatalf("Failed to load DAG after cache clear: %v", err)
	}

	if loadedDAG.GetCurrentIndex() != dag.GetCurrentIndex() {
		t.Error("Index mismatch after cache clear")
	}
}

func TestMerkleDAGDB_SaveIndex(t *testing.T) {
	memDB := rawdb.NewMemoryDatabase()
	dagDB := NewMerkleDAGDB(memDB)

	testIndexes := []uint64{0, 1, 100, 255}

	for _, idx := range testIndexes {
		// 保存索引
		if err := dagDB.saveIndex(idx); err != nil {
			t.Errorf("Failed to save index %d: %v", idx, err)
			continue
		}

		// 加载并验证索引
		loadedIdx, err := dagDB.loadIndex()
		if err != nil {
			t.Errorf("Failed to load index %d: %v", idx, err)
			continue
		}

		if loadedIdx != idx {
			t.Errorf("Index mismatch: got %d, want %d", loadedIdx, idx)
		}
	}
}

func TestMerkleDAGDB_LoadNonExistentDAG(t *testing.T) {
	memDB := rawdb.NewMemoryDatabase()
	dagDB := NewMerkleDAGDB(memDB)

	// 尝试加载不存在的 DAG
	loadedDAG, err := dagDB.LoadDAG()
	if err != nil {
		t.Fatalf("Loading non-existent DAG should not return error: %v", err)
	}

	// 验证返回了一个空的 DAG
	if !loadedDAG.IsEmpty() {
		t.Error("Loading non-existent DAG should return empty DAG")
	}

	if loadedDAG.GetCurrentIndex() != 0 {
		t.Error("Loading non-existent DAG should return DAG with index 0")
	}
}

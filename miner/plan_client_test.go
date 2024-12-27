package miner

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core"
)

func TestStartPlanClient(t *testing.T) {
	// 创建真实的PlanPool
	planPool := core.NewPlanPool()

	// 启动服务器
	port := 22325 // 使用不同的端口避免冲突
	err := StartPlanClient(planPool, port)
	if err != nil {
		t.Fatalf("启动Plan客户端失败: %v", err)
	}

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 测试用例1：发送有效的Plan数据
	t.Run("Valid Plan", func(t *testing.T) {
		// 创建测试Plan数据
		plan := &core.Plan{
			Height:     100,
			Difficulty: big.NewInt(1000),
			MinPowGas:  50000,
			MaxPowGas:  100000,
		}

		update := PlanUpdate{
			Action: "update",
			Level:  "parameter",
			Group:  "execution",
			Plan:   plan,
		}

		// 发送数据
		err := sendUpdate(port, update)
		if err != nil {
			t.Fatalf("发送数据失败: %v", err)
		}

		// 等待处理完成
		time.Sleep(100 * time.Millisecond)

		// 验证Plan是否被正确添加
		plans := planPool.GetPlansByHeight(100)
		if len(plans) == 0 {
			t.Fatal("Plan未被添加到Pool中")
		}

		found := false
		for _, p := range plans {
			if p.Height == plan.Height &&
				p.Difficulty.Cmp(plan.Difficulty) == 0 &&
				p.MinPowGas == plan.MinPowGas &&
				p.MaxPowGas == plan.MaxPowGas {
				found = true
				break
			}
		}
		if !found {
			t.Error("添加的Plan与原始Plan不匹配")
		}
	})

	// 测试用例2：发送错误的group值
	t.Run("Invalid Group", func(t *testing.T) {
		initialPlans := planPool.GetPlansByHeight(200)
		initialCount := len(initialPlans)

		update := PlanUpdate{
			Action: "update",
			Level:  "parameter",
			Group:  "invalid",
			Plan: &core.Plan{
				Height: 200,
			},
		}

		// 发送数据
		err := sendUpdate(port, update)
		if err != nil {
			t.Fatalf("发送数据失败: %v", err)
		}

		// 等待处理完成
		time.Sleep(100 * time.Millisecond)

		// 验证Plan未被添加
		currentPlans := planPool.GetPlansByHeight(200)
		if len(currentPlans) != initialCount {
			t.Error("无效的Plan被错误地添加到Pool中")
		}
	})

	// 测试用例3：发送空Plan
	t.Run("Empty Plan", func(t *testing.T) {
		initialMinHeight := planPool.GetMinHeight()

		update := PlanUpdate{
			Action: "update",
			Level:  "parameter",
			Group:  "execution",
			Plan:   nil,
		}

		// 发送数据
		err := sendUpdate(port, update)
		if err != nil {
			t.Fatalf("发送数据失败: %v", err)
		}

		// 等待处理完成
		time.Sleep(100 * time.Millisecond)

		// 验证空Plan未被添加（最小高度应该保持不变）
		if planPool.GetMinHeight() != initialMinHeight {
			t.Error("空Plan改变了Pool的状态")
		}
	})
}

// sendUpdate 辅助函数：发送更新数据到服务器
func sendUpdate(port int, update PlanUpdate) error {
	// 建立连接
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	defer conn.Close()

	// 序列化数据
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}

	// 发送长度
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))
	_, err = conn.Write(lengthBuf)
	if err != nil {
		return err
	}

	// 发送数据
	_, err = conn.Write(data)
	return err
}

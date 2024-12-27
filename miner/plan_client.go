package miner

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/ethereum/go-ethereum/core"
)

// PlanUpdate represents the received plan update message
type PlanUpdate struct {
	Action string     `json:"action"`
	Level  string     `json:"level"`
	Group  string     `json:"group"`
	Plan   *core.Plan `json:"plan"`
}

// StartPlanClient starts a TCP server to listen for plan updates
func StartPlanClient(planPool *core.PlanPool, port int) error {
	// 在新的goroutine中启动服务器
	go func() {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Printf("无法启动Plan客户端: %v", err)
			return
		}
		defer listener.Close()

		log.Printf("Plan客户端已启动，监听端口 %d...", port)

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("接受连接错误: %v", err)
				continue
			}
			// 为每个连接创建一个goroutine
			go handleConnection(conn, planPool)
		}
	}()

	return nil
}

func handleConnection(conn net.Conn, planPool *core.PlanPool) {
	defer conn.Close()

	// 读取长度信息（4字节）
	lengthBuf := make([]byte, 4)
	_, err := conn.Read(lengthBuf)
	if err != nil {
		log.Printf("读取数据长度错误: %v", err)
		return
	}

	// 将4字节转换为整数
	length := binary.BigEndian.Uint32(lengthBuf)

	// 读取实际数据
	dataBuf := make([]byte, length)
	_, err = conn.Read(dataBuf)
	if err != nil {
		log.Printf("读取数据错误: %v", err)
		return
	}

	// 解析JSON数据
	var update PlanUpdate
	err = json.Unmarshal(dataBuf, &update)
	if err != nil {
		log.Printf("JSON解析错误: %v", err)
		return
	}

	// 验证group字段
	if update.Group != "execution" {
		log.Printf("无效的group值: %s, 期望值为 'execution'", update.Group)
		return
	}

	// 验证必要字段
	if update.Plan == nil {
		log.Printf("Plan字段为空")
		return
	}

	// 添加plan到planPool
	planID := planPool.AddPlan(update.Plan)
	log.Printf("成功添加Plan，ID: %d", planID)

	// 打印接收到的数据
	log.Printf("收到Plan更新请求:\n")
	log.Printf("Action: %s\n", update.Action)
	log.Printf("Level: %s\n", update.Level)
	log.Printf("Group: %s\n", update.Group)
	log.Printf("Plan ID: %d\n", update.Plan.ID)
	log.Printf("Plan Height: %d\n", update.Plan.Height)
}

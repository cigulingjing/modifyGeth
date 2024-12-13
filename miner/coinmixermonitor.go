package miner

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type TransferMessage struct {
	To    common.Address
	Value *big.Int
}

// coin mixer contract address
var CoinMixerContractAddress = common.HexToAddress("0x445aB2C84c4144297f2F08fd8AC05406F14ff790")
var DepositMadeEventHash = crypto.Keccak256Hash([]byte("DepositMade(address,bytes,bytes,uint256)"))
var WithdrawMade2EventHash = crypto.Keccak256Hash([]byte("WithdrawMade2(address,bytes,uint256)"))

type CoinMixerMonitor struct {
	mux         *event.TypeMux
	eth         Backend
	chainConfig *params.ChainConfig
	nonceLock   sync.Mutex

	serving atomic.Bool
	server  *grpc.Server

	// transfer消息订阅
	transferCh  chan *TransferMessage
	transferSub event.Subscription

	quit chan struct{}

	mixerEventCh chan types.Log
}

func NewCoinMixerMonitor(eth Backend, config *params.ChainConfig, mux *event.TypeMux) *CoinMixerMonitor {
	m := &CoinMixerMonitor{
		mux:          mux,
		eth:          eth,
		chainConfig:  config,
		transferCh:   make(chan *TransferMessage, 10),
		quit:         make(chan struct{}),
		mixerEventCh: make(chan types.Log),
		server:       grpc.NewServer(),
	}

	return m
}

func (m *CoinMixerMonitor) start() {
	if !m.serving.Load() {
		// !!! 这一段应该进入配置文件
		listen, err := net.Listen("tcp", "127.0.0.1:9294") // will be included in config
		if err != nil {
			fmt.Println(err)
			panic("coin mixer monitor cannot listen!")
		}
		m.serving.Store(true)
		go m.server.Serve(listen)
	}
	go m.bindCoinMixer()
	go m.loop()
}

func (m *CoinMixerMonitor) Stop() {
	m.server.Stop()
	close(m.quit)
	m.transferSub.Unsubscribe()
	m.serving.Store(false)
}

func (m *CoinMixerMonitor) UTXODeposit(ctx context.Context, req *pb.UTXODepositRequest) (*pb.Empty, error) {
	fmt.Println("get UTXODeposit request")
	// 1. 解析交易要发送给CoinMixer合约的调用交易
	pbtxbytes := req.GetTx()
	if pbtxbytes == nil {
		return nil, fmt.Errorf("tx is nil")
	}
	// 反序列化
	tx := new(types.Transaction)
	pbTx := new(pb.Transaction)
	err := proto.Unmarshal(pbtxbytes, pbTx)
	if err != nil {
		return nil, fmt.Errorf("pb tx unmarshal failed: %v", err)
	}
	err = tx.UnmarshalBinary(pbTx.Payload)
	if err != nil {
		return nil, fmt.Errorf("tx unmarshal failed: %v", err)
	}
	// 2. 检查交易地址是否能对齐
	from, err := types.Sender(types.LatestSignerForChainID(m.eth.BlockChain().Config().ChainID), tx)
	if err != nil {
		return nil, fmt.Errorf("tx sender failed: %v", err)
	}
	if from.Hex() != common.BytesToAddress(req.GetAddr()).Hex() {
		return nil, fmt.Errorf("tx sender is not correct")
	}

	if tx.To().Hex() != CoinMixerContractAddress.Hex() {
		return nil, fmt.Errorf("tx to address is not coin mixer contract address")
	}

	// 3. 构造给CoinMixer合约add Balance的交易
	addBalanceTx := m.createTransaction()
	// 4. 将交易添加到txpool
	m.eth.TxPool().Add([]*types.Transaction{addBalanceTx}, true, false)

	// 5. 将交易添加到txpool
	m.eth.TxPool().Add([]*types.Transaction{tx}, true, false)

	return &pb.Empty{}, nil
}

func (m *CoinMixerMonitor) loop() {
	for {
		select {

		case mixerLog := <-m.mixerEventCh:
			// 处理币混合器事件
			m.handleMixerEvent(mixerLog)

		case <-m.quit:
			return
		}
	}
}

func (m *CoinMixerMonitor) bindCoinMixer() {
	logsCh := make(chan []*types.Log)
	sub := m.eth.BlockChain().SubscribeLogsEvent(logsCh)
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			log.Info("Error while listening for logs", "err", err)
		case logs := <-logsCh:
			// 过滤logs的地址和topic
			for _, log := range logs {
				if log.Address == CoinMixerContractAddress && log.Topics[0] == DepositMadeEventHash {
					m.mixerEventCh <- *log
				}
				if log.Address == CoinMixerContractAddress && log.Topics[0] == WithdrawMade2EventHash {
					m.mixerEventCh <- *log
				}
			}
		case <-m.quit:
			return
		}
	}
}

func (m *CoinMixerMonitor) handleMessageToTransfer(msg *TransferMessage) {
	fmt.Println("Handling transfer message in monitor!")
	fmt.Println(msg)
	// 发送消息给转账区
}

func (m *CoinMixerMonitor) createTransaction() *types.Transaction {

	var from common.Address
	// 使用第一个账户作为发送者
	if len(m.eth.AccountManager().Accounts()) != 0 {
		from = m.eth.AccountManager().Accounts()[0]
	}

	// 获取当前的 nonce
	m.nonceLock.Lock()
	nonce := m.eth.TxPool().Nonce(from)
	m.nonceLock.Unlock()

	// 获取当前的 gas price
	gasPrice := m.eth.BlockChain().CurrentHeader().BaseFee

	// 构造交易数据 - 0x0D05 + account address
	data := make([]byte, 0)
	// 添加 0x0D05
	prefix := []byte{0x0D, 0x05}
	data = append(data, prefix...)
	// 添加账户地址
	data = append(data, from.Bytes()...)

	// 创建交易对象
	tx := types.NewTransaction(
		nonce,                    // nonce
		CoinMixerContractAddress, // to address
		big.NewInt(0),            // value
		100000,                   // gas limit
		gasPrice,                 // gas price
		data,                     // data
	)

	// 获取钱包
	account := accounts.Account{Address: from}
	wallet, err := m.eth.AccountManager().Find(account)
	if err != nil {
		log.Error("Failed to find wallet for account", "err", err)
		return nil
	}

	// 签名交易
	signedTx, err := wallet.SignTx(account, tx, m.chainConfig.ChainID)
	if err != nil {
		log.Error("Failed to sign transaction", "err", err)
		return nil
	}

	return signedTx
}

func (m *CoinMixerMonitor) handleMixerEvent(log types.Log) {
	fmt.Println("Handling coin mixer event in monitor!")
	fmt.Println(log)
	// 在这里添加具体的处理逻辑
	if log.Topics[0] == DepositMadeEventHash {
		// 处理DepositMade事件
		fmt.Println("DepositMade event")
	} else if log.Topics[0] == WithdrawMade2EventHash {
		msg := &TransferMessage{
			To:    common.BytesToAddress(log.Topics[1].Bytes()),
			Value: big.NewInt(0).SetUint64(100000000000000000),
		}
		m.handleMessageToTransfer(msg)
		fmt.Println("WithdrawMade2 event")
	}
}

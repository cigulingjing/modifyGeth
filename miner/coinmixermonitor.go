package miner

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

type TransferMessage struct {
	To    common.Address
	Value *big.Int
}

// coin mixer contract address
var CoinMixerContractAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")
var CoinMixerEventHash = crypto.Keccak256Hash([]byte("pullcode(string)"))

type CoinMixerMonitor struct {
	mux          *event.TypeMux
	eth          Backend
	chainConfig  *params.ChainConfig
	etherbase    common.Address
	nonceLock    sync.Mutex
	mu           sync.Mutex // protect miner address
	currentNonce uint64

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
	}

	return m
}

func (m *CoinMixerMonitor) Start() {
	// 启动主循环
	go m.loop()
	go m.bindCoinMixer()
}

func (m *CoinMixerMonitor) Stop() {
	close(m.quit)
	m.transferSub.Unsubscribe()
}

// setEtherbase sets the etherbase used to initialize the block coinbase field.
func (m *CoinMixerMonitor) setEtherbase(addr common.Address) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.etherbase = addr
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
				if log.Address == CoinMixerContractAddress && log.Topics[0] == CoinMixerEventHash {
					m.mixerEventCh <- *log
				}
			}
		case <-m.quit:
			return
		}
	}
}

func (m *CoinMixerMonitor) handleTransferMessage(msg *TransferMessage) {
	fmt.Println("Handling transfer message in monitor!")
	fmt.Println(msg)
}

func (m *CoinMixerMonitor) createTransaction(data []byte) *types.Transaction {
	to := CoinMixerContractAddress
	m.nonceLock.Lock()
	defer m.nonceLock.Unlock()

	// 获取nonce
	nonce := m.eth.TxPool().Nonce(m.etherbase)
	if nonce > m.currentNonce {
		m.currentNonce = nonce
	}

	// 构造交易
	gasPrice := big.NewInt(30000000000) // 30 Gwei
	gasLimit := uint64(21000)           // 标准转账gas限制

	tx := types.NewTransaction(
		m.currentNonce,
		to,
		new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)),
		gasLimit,
		gasPrice,
		data,
	)

	// 签名交易
	signer := types.NewEIP155Signer(m.chainConfig.ChainID)
	// 这里应该通过传递一个参数来读取keystore文件
	keyjson, err := ioutil.ReadFile("../build/chain/node1/keystore/UTC--2024-12-03T06-58-06.965988566Z--b0725bdd29091782aadd05d693370408f46174db")
	if err != nil {
		log.Error("Failed to read key file", "err", err)
		return nil
	}

	account, err := keystore.DecryptKey(keyjson, "123456")
	if err != nil {
		log.Error("Failed to decrypt key", "err", err)
		return nil
	}

	signedTx, err := types.SignTx(tx, signer, account.PrivateKey)
	if err != nil {
		log.Error("Failed to sign transaction", "err", err)
		return nil
	}

	m.currentNonce++
	return signedTx
}

func (m *CoinMixerMonitor) sendTransaction(tx *types.Transaction) {
	err := m.eth.TxPool().Add([]*types.Transaction{tx}, true, false)
	if err != nil {
		log.Error("Failed to send transaction", "err", err)
	}
}

func (m *CoinMixerMonitor) handleMixerEvent(log types.Log) {
	fmt.Println("Handling coin mixer event in monitor!")
	fmt.Println(log)
	// 在这里添加具体的处理逻辑

}

// RPC 方法示例
func (m *CoinMixerMonitor) SendTransaction(data []byte) error {
	tx := m.createTransaction(data)
	if tx == nil {
		return fmt.Errorf("failed to create transaction")
	}

	m.sendTransaction(tx)
	return nil
}

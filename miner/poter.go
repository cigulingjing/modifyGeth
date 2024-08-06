package miner

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/proto/pb"
	"google.golang.org/grpc"
)

type poter struct {
	eth   Backend // blockchain and txpool
	chain *core.BlockChain

	// running atomic.Bool // a functional judge
	serving atomic.Bool

	poterClient pb.PoTExecutorClient

	server *grpc.Server // server pointer to the running server
	pb.UnimplementedPoTExecutorServer
}

func newPoter(eth Backend, cli pb.PoTExecutorClient) *poter {
	poter := &poter{
		eth:   eth,
		chain: eth.BlockChain(),
		// running: atomic.Bool{},
		serving: atomic.Bool{},

		server: grpc.NewServer(),
	}
	poter.poterClient = cli
	// Register the grpc server
	s := grpc.NewServer()
	pb.RegisterPoTExecutorServer(s, poter)
	poter.server = s // then we can handle the server

	poter.serving.Store(false)

	return poter
}

// start sets the running status as 1 and triggers new work submitting.
func (p *poter) start() {

	if !p.serving.Load() {
		// !!! 这一段应该进入配置文件
		listen, err := net.Listen("tcp", "127.0.0.1:9877") // will be included in config
		if err != nil {
			fmt.Println(err)
			panic("cannot listen!")
		}
		p.serving.Store(true)
		go p.server.Serve(listen)
	}

}

// func (p *poter) isRunning() bool {
// 	return p.running.Load()
// }

func (p *poter) close() {
	p.server.Stop()
	p.serving.Store(false)
}

func (p *poter) GetTxs(ctx context.Context, getTxRq *pb.GetTxRequest) (*pb.GetTxResponse, error) {
	// get the transaction from the txpool
	res := &pb.GetTxResponse{
		Start:  getTxRq.GetStartHeight(),
		End:    p.eth.BlockChain().CurrentHeader().Number.Uint64(),
		Blocks: make([]*pb.ExecuteBlock, 0),
	}

	for i := getTxRq.GetStartHeight(); i <= p.eth.BlockChain().CurrentHeader().Number.Uint64(); i++ {
		block := p.eth.BlockChain().GetBlockByNumber(i)
		header := &pb.ExecuteHeader{Height: block.Header().Number.Uint64()}
		txs := make([]*pb.ExecutedTx, 0)

		for _, tx := range block.Transactions() {
			txs = append(txs, &pb.ExecutedTx{
				TxHash: tx.Hash().Bytes(),
				Height: block.Header().Number.Uint64(),
				Data:   tx.Data(),
			})
		}

		res.Blocks = append(res.Blocks, &pb.ExecuteBlock{
			Header: header,
			Txs:    txs,
		})
	}

	return res, nil
}

func (p *poter) VerifyTxs(ctx context.Context, veriRq *pb.VerifyTxRequest) (*pb.VerifyTxResponse, error) {
	res := &pb.VerifyTxResponse{
		Txs:  veriRq.GetTxs(),
		Flag: make([]bool, len(veriRq.GetTxs())),
	}

	for _, txData := range veriRq.GetTxs() {
		// verify the transaction
		lookup, _, err := p.eth.BlockChain().GetTransactionLookup(common.Hash(txData.TxHash))
		if err != nil {
			return nil, err
		}
		if lookup == nil {
			res.Flag = append(res.Flag, false)
			continue
		}
		if lookup.BlockIndex == txData.ExecutedHeight {
			res.Flag = append(res.Flag, true)
		}
	}

	return res, nil

}

func (p *poter) CommitTxs(context.Context, *pb.CommitTxsRequest) (*pb.CommitTxsResponse, error) {
	return nil, nil
}

// incensentive

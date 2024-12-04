package voucher

// Used to solve some problems caused by abi coding
import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	// Method declaration
	BalanceOf     Method
	Buy           Method
	Use           Method
	CreateVoucher Method
	RootAddress   = &common.Address{}
	gas           = uint64(100000000)
)

// Init method in Voucher
func VoucherMethodInit(Contractabi *abi.ABI) {
	BalanceOf = NewMethod(Contractabi, "balanceOf", true, gas)
	Use = NewMethod(Contractabi, "use", false, gas)
	Buy = NewMethod(Contractabi, "buy", false, gas)
	CreateVoucher = NewMethod(Contractabi, "createVoucher", false, gas)
}


// TODO Extra attribute for voucher
// type VoucherInfos struct {
// 	OriginToken   common.Address
// 	MinAmount     *big.Int
// 	ExchangeRate  *big.Int
// 	BuyExpiration *big.Int
// 	UseExpiration *big.Int
// }
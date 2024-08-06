package vm

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
)

var errStore = errors.New("error pangu add : input length must be 64 bytes")
var errVerify = errors.New("error pangu verify: input length is insufficient")
var errProve = errors.New("error pangu prove: input length must be 64 bytes")

const Mysha3Size = 224

func Mysha3(data []byte) []byte {
	h := sha256.New224()
	h.Write(data)
	return h.Sum(nil)
}

func hmac(key []byte, data []byte) []byte {
	h := sha256.New()
	h.Write(key)
	h.Write(data)
	return h.Sum(nil)
}

func Setup(nBits int) (*big.Int, *big.Int) {
	p, _ := rand.Prime(rand.Reader, nBits/2)
	q, _ := rand.Prime(rand.Reader, nBits/2)
	return p, q
}

func evalTrap(x []byte, n *big.Int, e *big.Int) []byte {
	xInt := new(big.Int).SetBytes(x)
	r := new(big.Int).Exp(xInt, e, n)
	return r.Bytes()
}

func eval(x []byte, n *big.Int, t uint) []byte {
	g := new(big.Int).SetBytes(x)
	exp := new(big.Int).SetUint64(1 << t)
	result := new(big.Int).Exp(g, exp, n)
	return result.Bytes()
}

// 使用陷门排列store生成并存储数据的身份验证信息
func store(c []byte, d []byte, p *big.Int, q *big.Int, t int, k int) ([]byte, []byte) {
	fmt.Println("p:", p.Int64())
	fmt.Println("q:", q.Int64())
	fmt.Println("t, k:", t, k)
	one := big.NewInt(1)
	n := new(big.Int).Mul(p, q)
	phi := new(big.Int).Mul(new(big.Int).Sub(p, one), new(big.Int).Sub(q, one))
	e := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(t)), nil)
	e.Mod(e, phi)

	var cs, vs []byte
	//迭代k次
	fmt.Println("start loop")
	for i := 0; i <= k; i++ {
		fmt.Printf("store 迭代次数:%d\n", i)
		//计算文件的hmac
		v := hmac(c, d)
		//appends c to the cs byte slice and v to the vs byte slice.
		cs = append(cs, c...)
		vs = append(vs, v...)
		//
		c = Mysha3(evalTrap(Mysha3(v), n, e))
		fmt.Printf("已经进行了%d轮\n", i)
	}
	return Mysha3(cs), Mysha3(vs)
}

// 使用 Store 函数生成的身份验证信息为给定文件 d 生成真实性证明
func prove(c []byte, d []byte, n *big.Int, t uint, k int) ([]byte, []byte) {
	var cs, vs []byte
	for i := 0; i <= k; i++ {
		//fmt.Printf("prove 迭代次数:%d\n", i)
		v := hmac(c, d)
		cs = append(cs, c...)
		vs = append(vs, v...)
		c = Mysha3(eval(Mysha3(v), n, t))
		//fmt.Printf("已经进行了%d轮\n", i)
	}
	return Mysha3(cs), Mysha3(vs)
}

func verify(c []byte, b []byte, a []byte, n *big.Int, t uint) bool {
	// Calculate the number of iterations
	k := len(a)/(Mysha3Size+int(t/8)) - 1

	// Recompute c from b and d
	d := make([]byte, len(c)+len(a))
	copy(d, c)
	copy(d[len(c):], a)
	cRecomputed := Mysha3(d)

	// Recompute v from b and d
	v := make([]byte, 0, Mysha3Size*(k+1))
	for i := 0; i <= k; i++ {
		offset := len(c) + i*(Mysha3Size+int(t/8))
		v = append(v, b[offset:offset+Mysha3Size]...)
	}
	vRecomputed := Mysha3(v)

	// Verify the commitments
	if !bytes.Equal(c, cRecomputed) {
		return false
	}

	// Verify the challenge values
	if !bytes.Equal(v, vRecomputed) {
		return false
	}

	// Verify the responses
	for i := 0; i <= k; i++ {
		offset := len(c) + i*(Mysha3Size+int(t/8))
		r := new(big.Int).SetBytes(a[offset : offset+int(t/8)])
		x := new(big.Int).SetBytes(a[offset+int(t/8) : offset+Mysha3Size+int(t/8)])
		fx := evalTrap(x.Bytes(), n, big.NewInt(int64(t)))
		ax := new(big.Int).SetBytes(fx)
		gx := new(big.Int).Exp(ax, r, n)
		//gx  := new(big.Int).Exp(new(big.Int).SetBytes(fx), r, n)
		if !bytes.Equal(x.Bytes(), Mysha3(gx.Bytes())) {
			return false
		}
	}

	return true
}

// func getHash() string {
// 	file, err := os.Open("random.txt")
// 	if err != nil {
// 		fmt.Println("Error opening file:", err)
// 	}
// 	defer file.Close()

// 	hash := sha256.New()
// 	if _, err := io.Copy(hash, file); err != nil {
// 		fmt.Println("Error calculating file hash:", err)
// 	}

// 	fmt.Printf("File hash: %x\n", hash.Sum(nil))
// 	return fmt.Sprintf("%x", hash.Sum(nil))
// }

type panguStore struct{}

func (p *panguStore) RequiredGas(input []byte) uint64 {
	// 自定义Gas计算方法
	// Input为 tx msg 中的 data，如果需要按操作计算Gas，需要自行解析
	return 10
}

func (p *panguStore) Run(input []byte, blkCtx BlockContext) ([]byte, error) {
	fmt.Println("get in", len(input))
	if len(input) < 10 { // 确保至少有两个长度前缀和一些数据
		return nil, errStore
	}

	offset := 0

	// 解析 c 和 d 的长度
	cLen := int(input[offset])<<8 | int(input[offset+1])
	offset += 2
	dLen := int(input[offset])<<8 | int(input[offset+1])
	offset += 2

	// 确保输入数据足够长以包含 c, d, p, q, t, k
	if len(input) < offset+cLen+dLen+4*32 {
		fmt.Println("input not enough")
		return nil, errStore
	}

	fmt.Println("1")
	// 解析 c 和 d
	c := input[offset : offset+cLen]
	offset += cLen
	d := input[offset : offset+dLen]
	offset += dLen

	fmt.Println("2")
	// 解析 p, q, t, k
	pBytes := input[offset : offset+32]
	offset += 32
	qBytes := input[offset : offset+32]
	offset += 32
	tBytes := input[offset : offset+32]
	offset += 32
	kBytes := input[offset : offset+32]
	offset += 32

	fmt.Println("3")
	// 将字节数据转换为大数和整数
	pBig := new(big.Int).SetBytes(pBytes)
	qBig := new(big.Int).SetBytes(qBytes)
	tBig := new(big.Int).SetBytes(tBytes)
	kBig := new(big.Int).SetBytes(kBytes)

	tInt := int(tBig.Int64())
	kInt := int(kBig.Int64())

	fmt.Println("4")
	// 调用 store 函数
	cs, vs := store(c, d, pBig, qBig, tInt, kInt)
	fmt.Println("over")
	return append(cs, vs...), nil
}

type panguStoreVerify struct{}

func (p *panguStoreVerify) RequiredGas(input []byte) uint64 {
	// 自定义Gas计算方法
	// Input为 tx msg 中的 data，如果需要按操作计算Gas，需要自行解析
	return 30
}

func (p *panguStoreVerify) Run(input []byte, blkCtx BlockContext) ([]byte, error) {
	if len(input) < 12 { // 确保至少有三个长度前缀和一些数据
		return nil, errVerify
	}

	offset := 0

	// 解析 c, b, a 的长度
	cLen := int(input[offset])<<8 | int(input[offset+1])
	offset += 2
	bLen := int(input[offset])<<8 | int(input[offset+1])
	offset += 2
	aLen := int(input[offset])<<8 | int(input[offset+1])
	offset += 2

	// 确保输入数据足够长以包含 c, b, a, n, t
	if len(input) < offset+cLen+bLen+aLen+32+4 {
		return nil, errVerify
	}

	// 解析 c, b, a
	c := input[offset : offset+cLen]
	offset += cLen
	b := input[offset : offset+bLen]
	offset += bLen
	a := input[offset : offset+aLen]
	offset += aLen

	// 解析 n, t
	nBytes := input[offset : offset+32]
	offset += 32
	tBytes := input[offset : offset+4]
	offset += 4

	// 将字节数据转换为大数和整数
	nBig := new(big.Int).SetBytes(nBytes)
	tInt := int(new(big.Int).SetBytes(tBytes).Uint64())

	// 调用 verify 函数
	valid := verify(c, b, a, nBig, uint(tInt))

	// 返回验证结果
	if valid {
		return []byte{1}, nil
	} else {
		return []byte{0}, nil
	}
}

type panguStoreProve struct{}

func (p *panguStoreProve) RequiredGas(input []byte) uint64 {
	// 自定义Gas计算方法
	// Input为 tx msg 中的 data，如果需要按操作计算Gas，需要自行解析
	return 20
}

func (p *panguStoreProve) Run(input []byte, blkCtx BlockContext) ([]byte, error) {
	if len(input) < 10 { // 确保至少有两个长度前缀和一些数据
		return nil, errProve
	}

	offset := 0

	// 解析 c 和 d 的长度
	cLen := int(input[offset])<<8 | int(input[offset+1])
	offset += 2
	dLen := int(input[offset])<<8 | int(input[offset+1])
	offset += 2

	// 确保输入数据足够长以包含 c, d, n, t, k
	if len(input) < offset+cLen+dLen+3*32 {
		return nil, errProve
	}

	// 解析 c 和 d
	c := input[offset : offset+cLen]
	offset += cLen
	d := input[offset : offset+dLen]
	offset += dLen

	// 解析 n, t, k
	nBytes := input[offset : offset+32]
	offset += 32
	tBytes := input[offset : offset+32]
	offset += 32
	kBytes := input[offset : offset+32]
	offset += 32

	nBig := new(big.Int).SetBytes(nBytes)
	tBig := new(big.Int).SetBytes(tBytes)
	kBig := new(big.Int).SetBytes(kBytes)
	// 将字节数据转换为大数和整数

	tInt := int(tBig.Int64())
	kInt := int(kBig.Int64())

	// 调用 prove 函数
	cs, vs := prove(c, d, nBig, uint(tInt), kInt)
	return append(cs, vs...), nil
}

package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"math"
	"math/big"
)

const targetBits = 24

var (
	maxNonce = math.MaxInt64
)

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{b, target}
	return pow
}

func (pow *ProofOfWork) Run() (nonce int, hashRes []byte, err error) {
	var hashInt big.Int
	nonce = 0

	logrus.Infof("Mining the block containing '%s'", pow.block.Data)
	for nonce < maxNonce {
		var data []byte
		data, err = pow.prepareData(nonce)
		if err != nil {
			return
		}
		hash := sha256.Sum256(data)

		// to log the hash generate
		if math.Remainder(float64(nonce), 100000) == 0 {
			//logrus.Infof("Hash is \r%x", hash)
		}
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			logrus.Infof("Mining block is success, Hash is \r%x", hash)
			hashRes = hash[:]
			break
		} else {
			nonce++
		}
	}
	return
}

func (pow *ProofOfWork) prepareData(nonce int) ([]byte, error) {
	timeHex, err := intToHex(pow.block.Timestamp)
	if err != nil {
		return nil, err
	}
	targetBitsHex, err := intToHex(int64(targetBits))
	if err != nil {
		return nil, err
	}
	nonceHex, err := intToHex(int64(nonce))
	if err != nil {
		return nil, err
	}
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			timeHex,
			targetBitsHex,
			nonceHex,
		},
		[]byte{},
	)
	return data, nil
}

func intToHex(num int64) ([]byte, error) {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data, err := pow.prepareData(pow.block.Nonce)
	if err != nil {
		return false
	}
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(pow.target) == -1
}

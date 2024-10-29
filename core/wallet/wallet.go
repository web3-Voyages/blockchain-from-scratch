package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"
)

const version = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewWallet() *Wallet {
	private, public := newKeyPair()
	return &Wallet{private, public}
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	public := append(privateKey.PublicKey.X.Bytes(), privateKey.Y.Bytes()...)
	return *privateKey, public
}

// BitCoin public address generate: https://3bcaf57.webp.li/myblog/BtcPublicKeyGenerate.png
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checkSum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checkSum...)
	return Base58Encode(fullPayload)
}

func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	// 使用 SHA-256 代替 RIPEMD-160
	publicSHA256Again := sha256.Sum256(publicSHA256[:])
	return publicSHA256Again[:20] // 取前20字节作为哈希值
}

func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

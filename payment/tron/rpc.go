package tron

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"github.com/chenqgp/abc/conf"
	"log"
	"math/big"
	"time"

	"github.com/chenqgp/abc/third/telegram"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

//var TRCAddress = "TUziUk5ga946kTFJQsZMF4aJfz4rPhS1wF"
//var TRCPrivate = "70cf43308cfd8c3376271b706b80cf6df5a3f28b80a619317e79e17b1eb21a91"
//var ColdTRCAddress = "TQ8xcu7gg61ZkUbeXxBWGEv2ZPKvJ2Uy4q"
//var USDTContract = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
//var USDTContractHex = "a614f803b6fd780986a42c78ec9c7f77e6ded13c"

func signTransation(transaction *core.Transaction, privateKeyHex string) *core.Transaction {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		panic("tron signTransation1" + err.Error())
	}
	rawData, err := proto.Marshal(transaction.GetRawData())
	if err != nil {
		panic("tron signTransation2" + err.Error())
	}

	signed, err := crypto.Sign(s256(rawData), privateKey)
	if err != nil {
		panic("tron signTransation3" + err.Error())
	}
	transaction.Signature = append(transaction.Signature, signed)

	return transaction
}

func s256(s []byte) []byte {
	h := sha256.New()
	h.Write(s)
	return h.Sum(nil)
}

func tob58(s string) string {
	addr, _ := hex.DecodeString(s)
	h := s256(s256(addr))
	secret := h[:4]
	addr = append(addr, secret...)
	return string(Base58Encode(addr))
}

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func Base58Encode(input []byte) []byte {
	var result []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	// https://en.bitcoin.it/wiki/Base58Check_encoding#Version_bytes
	if input[0] == 0x00 {
		result = append(result, b58Alphabet[0])
	}

	length := len(result)
	for i := 0; i < length/2; i++ {
		result[i], result[length-1-i] = result[length-1-i], result[i]
	}

	return result
}

func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)

	for _, b := range input {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()

	if input[0] == b58Alphabet[0] {
		decoded = append([]byte{0x00}, decoded...)
	}

	return decoded
}

func newClient() *client.GrpcClient {
	conn := client.NewGrpcClient("grpc.trongrid.io:50051")
	err := conn.Start(grpc.WithInsecure())
	if err != nil {
		panic("tron newClient" + err.Error())
	}
	return conn
}

func CreateWallet() (string, string) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		panic("tron CreateWallet1" + err.Error())
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	//log.Println("privateKeyBytes: ", privateKeyBytes)
	//log.Println("privateKeyHex: ", hexutil.Encode(privateKeyBytes)[2:])

	publicKey := privateKey.Public()
	publicECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("tron CreateWallet error casting public key to ECDSA")
	}
	//publicKeyBytes := crypto.FromECDSAPub(publicECDSA)
	//log.Println("publicKeyByte:", publicKeyBytes)
	//log.Println("publicKeyHex:", hexutil.Encode(publicKeyBytes)[4:])

	address := crypto.PubkeyToAddress(*publicECDSA).Hex()
	//log.Println("address:", address)
	address58 := tob58("41" + address[2:])
	//log.Println("address58:", address58)

	conn := newClient()
	tx, err := conn.CreateAccount(conf.TRCAddress, address58)
	if err != nil {
		telegram.SendMsg(telegram.TEXT, telegram.TEST, "tron CreateWallet"+err.Error())
		log.Println("tron CreateWallet2" + err.Error())
		return "", ""
	}

	trx := signTransation(tx.Transaction, conf.TRCPrivate)
	r, err := conn.Broadcast(trx)
	if err != nil {
		panic("tron CreateWallet3" + err.Error())
	}
	if r.Code != 0 {
		panic("tron CreateWallet4" + string(r.GetMessage()))
	}

	//hash := sha3.NewLegacyKeccak256()
	//log.Println("publicKeyBytes[1:]: ", publicKeyBytes[1:])
	//hash.Write(publicKeyBytes[1:])
	//log.Println("hashRaw: ", hexutil.Encode(hash.Sum(nil)))
	//log.Println("hash: ", hexutil.Encode(hash.Sum(nil)[12:]))
	return address58, hexutil.Encode(privateKeyBytes)[2:]
}

func TransferTrx(from, to, key string, amount int64) {
	conn := newClient()
	tx, err := conn.Transfer(from, to, amount*int64(1000000))
	if err != nil {
		telegram.SendMsg(telegram.TEXT, telegram.TEST, "tron TransferTrx"+err.Error())
		log.Println("tron TransferTrx1" + err.Error())
		return
	}

	trx := signTransation(tx.Transaction, key)
	r, err := conn.Broadcast(trx)
	if err != nil {
		panic("tron TransferTrx2" + err.Error())
	}
	if r.Code != 0 {
		panic("tron TransferTrx3" + string(r.GetMessage()))
	}
	txConfirmation(tx.Transaction)
	return
}

func USDTTransfer(from, to, key string, amount, fee int64) {
	log.Println("tron", from, amount, fee)
	conn := newClient()
	tx, err := conn.TRC20Send(from, to, conf.USDTContract, big.NewInt(amount), fee*int64(1000000))
	if err != nil {
		telegram.SendMsg(telegram.TEXT, telegram.TEST, "tron USDTTransfer"+err.Error())
		log.Println("tron USDTTransfer1" + err.Error())
		return
	}

	trx := signTransation(tx.Transaction, key)
	r, err := conn.Broadcast(trx)
	if err != nil {
		panic("tron USDTTransfer2" + err.Error())
	}
	if r.Code != 0 {
		panic("tron USDTTransfer3" + string(r.GetMessage()))
	}
	txConfirmation(tx.Transaction)
	return
}

func USDTBalance(from string) int64 {
	conn := newClient()
	r, err := conn.TRC20ContractBalance(from, conf.USDTContract)
	if err != nil {
		panic("tron USDTBalance" + err.Error())
	}

	return r.Int64()
}

func TrxBalance(from string) int64 {
	conn := newClient()
	r, err := conn.GetAccountDetailed(from)
	if err != nil {
		panic("tron TrxBalance" + err.Error())
	}

	return r.Balance
}

func txConfirmation(tx *core.Transaction) {
	conn := newClient()
	txHash, err := TransactionHash(tx)
	if err != nil {
		panic("tron txConfirmation1" + err.Error())
	}
	start := 10
	for {
		if txi, err := conn.GetTransactionInfoByID(txHash); err == nil {
			if txi.Result != 0 {
				panic("tron txConfirmation2" + string(txi.ResMessage))
			}
			log.Println("tron txConfirmation3", txi.Fee, txi.Receipt.NetFee)
			return
		}
		if start < 0 {
			return
		}
		time.Sleep(time.Second)
		start--
	}
}

func TransactionHash(tx *core.Transaction) (string, error) {
	rawData, err := proto.Marshal(tx.GetRawData())
	if err != nil {
		return "", err
	}
	h256h := sha256.New()
	h256h.Write(rawData)
	hash := h256h.Sum(nil)
	return common.ToHex(hash), nil
}

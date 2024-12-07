package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"
	"timechain-gateway/internal/config"
	"timechain-gateway/pkg/models"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Contract ABI 常量，需要根据你的智能合约生成
const ContractABI = `[{"inputs":[{"components":[...],"name":"data","type":"tuple[][]"}],"name":"sendData","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

// 合约地址常量
// const ContractAddress = "0x742d35Cc6634C0532925a3b844Bc454e4438f44e" // 替换为你的合约地址

type Client struct {
	config     *config.Config
	eth        *ethclient.Client
	contract   *bind.BoundContract
	privateKey *ecdsa.PrivateKey
	address    common.Address
	gasLimit   uint64
}

func NewClient(cfg *config.Config) *Client {
	// 连接到以太坊网络
	client, err := ethclient.Dial(cfg.BlockchainConfig.NodeURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// 加载私钥
	privateKey, err := crypto.HexToECDSA(cfg.BlockchainConfig.PrivateKey)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// 从私钥获取公钥和地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Failed to get public key")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 创建合约实例
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	contractAddress := common.HexToAddress(cfg.BlockchainConfig.ContractAddress)
	contract := bind.NewBoundContract(contractAddress, parsed, client, client, client)

	return &Client{
		config:     cfg,
		eth:        client,
		contract:   contract,
		privateKey: privateKey,
		address:    address,
		gasLimit:   cfg.BlockchainConfig.GasLimit,
	}
}

func (c *Client) SendData(data [][]models.SensorData, batchNum int) error {
	// 获取nonce
	nonce, err := c.eth.PendingNonceAt(context.Background(), c.address)
	if err != nil {
		return err
	}

	// 获取gas价格
	gasPrice, err := c.eth.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	// 准备交易选项
	auth, err := bind.NewKeyedTransactorWithChainID(c.privateKey, big.NewInt(c.config.BlockchainConfig.ChainID)) // Sepolia chain ID
	if err != nil {
		return err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)                         // 不发送ETH
	auth.GasLimit = c.config.BlockchainConfig.GasLimit // 设置gas限制
	auth.GasPrice = gasPrice

	// 准备合约调用数据
	payload := getLSH(data)

	// 调用合约方法
	tx, err := c.contract.Transact(auth, "sendData", payload)
	if err != nil {
		return err
	}

	// 等待交易被确认
	receipt, err := bind.WaitMined(context.Background(), c.eth, tx)
	if err != nil {
		return err
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction failed")
	}

	return nil
}

// 添加其他辅助方法
func (c *Client) GetBalance() (*big.Int, error) {
	return c.eth.BalanceAt(context.Background(), c.address, nil)
}

func (c *Client) Close() {
	c.eth.Close()
}

func getLSH(data [][]models.SensorData) []common.Hash {
	if len(data) == 0 || len(data[0]) != 32 {
		log.Fatal("Invalid data dimensions")
	}

	var rootHashes []common.Hash
	for _, batch := range data {
		var leaves []common.Hash
		for _, sensorData := range batch {
			leaf := crypto.Keccak256Hash(
				[]byte(fmt.Sprintf("%v", sensorData.Timestamp)),
				[]byte(fmt.Sprintf("%v", sensorData.Value)),
				[]byte(fmt.Sprintf("%v", sensorData.DeviceID)),
				[]byte(fmt.Sprintf("%v", sensorData.SensorID)),
				[]byte(fmt.Sprintf("%v", sensorData.Type)),
				[]byte(fmt.Sprintf("%v", sensorData.Unit)),
			)
			leaves = append(leaves, leaf)
		}
		rootHash := calculateMerkleRoot(leaves)
		rootHashes = append(rootHashes, rootHash)
	}
	return rootHashes
}

func calculateMerkleRoot(leaves []common.Hash) common.Hash {
	if len(leaves) == 0 {
		return common.Hash{}
	}
	for len(leaves) > 1 {
		var newLevel []common.Hash
		for i := 0; i < len(leaves); i += 2 {
			if i+1 < len(leaves) {
				newLevel = append(newLevel, crypto.Keccak256Hash(leaves[i].Bytes(), leaves[i+1].Bytes()))
			} else {
				newLevel = append(newLevel, leaves[i])
			}
		}
		leaves = newLevel
	}
	return leaves[0]
}

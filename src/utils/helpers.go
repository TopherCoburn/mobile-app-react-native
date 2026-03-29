package mobileappreactnative

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

const (
	ETHERSCAN_API_KEY = "YOUR_ETHERSCAN_API_KEY"
)

type Config struct {
	EthereumNodeURL string
}

type Transaction struct {
	BlockNumber    string `json:"blockNumber"`
	BlockHash      string `json:"blockHash"`
	TransactionHash string `json:"transactionHash"`
	From            string `json:"from"`
	To              string `json:"to"`
	Value           string `json:"value"`
	Gas             string `json:"gas"`
	GasPrice        string `json:"gasPrice"`
	Nonce           string `json:"nonce"`
	Timestamp       string `json:"timestamp"`
}

type Block struct {
	Number    string `json:"number"`
	Hash      string `json:"hash"`
	Transactions []Transaction `json:"transactions"`
}

type BlockHeader struct {
	ParentHash string `json:"parentHash"`
	Number     string `json:"number"`
}

type EthereumNodeClient struct {
	client *ethclient.Client
}

func NewEthereumNodeClient(config Config) (*EthereumNodeClient, error) {
	client, err := ethclient.Dial(config.EthereumNodeURL)
	if err!= nil {
		return nil, err
	}
	return &EthereumNodeClient{client: client}, nil
}

func (ec *EthereumNodeClient) GetBlockByHash(hash common.Hash) (*Block, error) {
	block, err := ec.client.BlockByHash(hash)
	if err!= nil {
		return nil, err
	}
	return &Block{
		Number:    block.Number().String(),
		Hash:      block.Hash().String(),
		Transactions: make([]Transaction, 0, len(block.Transactions())),
	}, nil
}

func (ec *EthereumNodeClient) GetBlockByNumber(number string) (*Block, error) {
	block, err := ec.client.BlockByNumber(big.NewInt(0), big.NewInt(0))
	if err!= nil {
		return nil, err
	}
	return &Block{
		Number:    block.Number().String(),
		Hash:      block.Hash().String(),
		Transactions: make([]Transaction, 0, len(block.Transactions())),
	}, nil
}

func GetBlockHeaderByHash(hash common.Hash) (*BlockHeader, error) {
	block, err := GetBlockByHash(hash)
	if err!= nil {
		return nil, err
	}
	return &BlockHeader{
		ParentHash: block.ParentHash,
		Number:     block.Number,
	}, nil
}

func GetBlockHeaderByNumber(number string) (*BlockHeader, error) {
	block, err := GetBlockByNumber(number)
	if err!= nil {
		return nil, err
	}
	return &BlockHeader{
		ParentHash: block.ParentHash,
		Number:     block.Number,
	}, nil
}

func GetTransactionByHash(hash common.Hash) (*Transaction, error) {
	block, err := GetBlockByHash(hash)
	if err!= nil {
		return nil, err
	}
	for _, tx := range block.Transactions {
		if tx.TransactionHash == hash.String() {
			return &tx, nil
		}
	}
	return nil, errors.New("transaction not found")
}

func GetTransactionCount(address string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&sort=asc&apikey=%s", address, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	return len(data.Result), nil
}

func GenerateAddressFromPrivateKey(privateKey string) (string, error) {
	pubKey, err := crypto.ToECDSA(common.FromHex(privateKey))
	if err!= nil {
		return "", err
	}
	address := crypto.PubkeyToAddress(*pubKey).Hex()
	return address, nil
}

func GeneratePrivateKey() (string, error) {
	privateKey, err := crypto.GenerateKey()
	if err!= nil {
		return "", err
	}
	return crypto.ToECDSA(privateKey).PrivKey().Hex(), nil
}

func GenerateAddressFromMnemonic(mnemonic string, path string) (string, error) {
	privateKey, err := crypto.GenerateDeterministicKeyFromMnemonic(mnemonic, path)
	if err!= nil {
		return "", err
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	return address, nil
}

func GenerateMnemonic() (string, error) {
	privateKey, err := crypto.GenerateKey()
	if err!= nil {
		return "", err
	}
	mnemonic := crypto.ToECDSA(privateKey).PrivKey().Hex()
	return mnemonic, nil
}

func GetContractAbi(contractAddress string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=%s", contractAddress, ETHERSCAN_API_KEY))
	if err!= nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return nil, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  string `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return nil, err
	}

	return []byte(data.Result), nil
}

func GetContractCode(contractAddress string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=getsourcecode&address=%s&apikey=%s", contractAddress, ETHERSCAN_API_KEY))
	if err!= nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return nil, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			SourceCode string `json:"SourceCode"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return nil, err
	}

	return []byte(data.Result.SourceCode), nil
}

func DownloadFile(url string, dest string) error {
	resp, err := http.Get(url)
	if err!= nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode!= http.StatusOK {
		return errors.New("failed to download file")
	}

	out, err := os.Create(dest)
	if err!= nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func GetEtherscanTxUrl(txHash string) string {
	return fmt.Sprintf("https://etherscan.io/tx/%s", txHash)
}

func GetEtherscanAddrUrl(address string) string {
	return fmt.Sprintf("https://etherscan.io/address/%s", address)
}

func GetRestyClient() (*resty.Client, error) {
	client := resty.New()
	client.SetRetryCount(3)
	client.SetRetryMaxBackoff(10 * time.Second)
	return client, nil
}

func GetTokenPrice(tokenAddress string) (float64, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=stats&action=tokenprice&contractaddress=%s&apikey=%s", tokenAddress, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Price float64 `json:"Price"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	return data.Result.Price, nil
}

func GetTokenPriceUsd(tokenAddress string) (float64, error) {
	price, err := GetTokenPrice(tokenAddress)
	if err!= nil {
		return 0, err
	}

	return price * 1.1, nil
}

func GetBlockNumber() (int, error) {
	resp, err := http.Get("https://api.etherscan.io/api?module=proxy&action=eth_blockNumber&apikey=YOUR_ETHERSCAN_API_KEY")
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  string `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	blockNumber, err := strconv.ParseInt(data.Result, 16, 64)
	if err!= nil {
		return 0, err
	}

	return int(blockNumber), nil
}

func GetTokenSupply(tokenAddress string) (int64, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=stats&action=tokenbalance&contractaddress=%s&address=0x0000000000000000000000000000000000000000&tag=latest&apikey=%s", tokenAddress, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Tokens string `json:"result"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	tokens, err := strconv.ParseInt(data.Result.Tokens, 10, 64)
	if err!= nil {
		return 0, err
	}

	return tokens, nil
}

func GetTokenDecimals(tokenAddress string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=stats&action=tokenholder&contractaddress=%s&address=0x0000000000000000000000000000000000000000&tag=latest&apikey=%s", tokenAddress, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Decimals string `json:"Decimals"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	decimals, err := strconv.ParseInt(data.Result.Decimals, 10, 64)
	if err!= nil {
		return 0, err
	}

	return int(decimals), nil
}

func GetEtherscanTxStatus(txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionReceipt&txhash=%s&apikey=%s", txHash, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Status string `json:"status"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result.Status, nil
}

func GetTxCountByAddress(address string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&sort=asc&apikey=%s", address, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	return len(data.Result), nil
}

func GetTxCountByBlockNumber(blockNumber string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=%s&apikey=%s", blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Transactions []struct {
				Hash string `json:"hash"`
			} `json:"transactions"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	return len(data.Result.Transactions), nil
}

func GetTxHashByBlockNumber(blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=%s&apikey=%s", blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result.Hash, nil
}

func GetBlockGasUsedByBlockNumber(blockNumber string) (int64, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=%s&apikey=%s", blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			GasUsed int64 `json:"gasUsed"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	return data.Result.GasUsed, nil
}

func GetTxGasUsedByTxHash(txHash string) (int64, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", txHash, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			GasUsed int64 `json:"gasUsed"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	return data.Result.GasUsed, nil
}

func GetTxBlockNumberByTxHash(txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", txHash, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			BlockNumber string `json:"blockNumber"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result.BlockNumber, nil
}

func GetTxBlockHashByTxHash(txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", txHash, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			BlockHash string `json:"blockHash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result.BlockHash, nil
}

func GetTxGasPriceByTxHash(txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", txHash, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			GasPrice string `json:"gasPrice"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result.GasPrice, nil
}

func GetTxValueByTxHash(txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", txHash, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Value string `json:"value"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result.Value, nil
}

func GetTxNonceByTxHash(txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", txHash, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Nonce string `json:"nonce"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result.Nonce, nil
}

func GetTxTimestampByTxHash(txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", txHash, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Timestamp string `json:"timestamp"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result.Timestamp, nil
}

func GetTxFromByTxHash(txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", txHash, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			From string `json:"from"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result.From, nil
}

func GetTxToByTxHash(txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", txHash, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			To string `json:"to"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result.To, nil
}

func GetTxGasByTxHash(txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", txHash, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Gas string `json:"gas"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result.Gas, nil
}

func GetTxGasPriceByBlockNumber(blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=%s&apikey=%s", blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			GasPrice string `json:"gasPrice"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result.GasPrice, nil
}

func GetTxValueByBlockNumber(blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=%s&apikey=%s", blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			Transactions []struct {
				Value string `json:"value"`
			} `json:"transactions"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	var totalValue string
	for _, tx := range data.Result.Transactions {
		totalValue = add(totalValue, tx.Value)
	}

	return totalValue, nil
}

func add(a, b string) string {
	x, _ := strconv.ParseInt(a, 10, 64)
	y, _ := strconv.ParseInt(b, 10, 64)
	return strconv.FormatInt(x+y, 10)
}

func GetTxCountByContractAddress(contractAddress string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=gettxlist&address=%s&startblock=0&endblock=99999999&sort=asc&apikey=%s", contractAddress, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	return len(data.Result), nil
}

func GetTxCountByBlockNumberByContractAddress(contractAddress string, blockNumber string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=gettxlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", contractAddress, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	return len(data.Result), nil
}

func GetTxHashByBlockNumberByContractAddress(contractAddress string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=gettxlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", contractAddress, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].Hash, nil
}

func GetTxBlockNumberByBlockNumberByContractAddress(contractAddress string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=gettxlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", contractAddress, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			BlockNumber string `json:"blockNumber"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].BlockNumber, nil
}

func GetTxBlockHashByBlockNumberByContractAddress(contractAddress string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=gettxlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", contractAddress, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			BlockHash string `json:"blockHash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].BlockHash, nil
}

func GetTxGasPriceByBlockNumberByContractAddress(contractAddress string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=gettxlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", contractAddress, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			GasPrice string `json:"gasPrice"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].GasPrice, nil
}

func GetTxValueByBlockNumberByContractAddress(contractAddress string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=contract&action=gettxlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", contractAddress, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Value string `json:"value"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	var totalValue string
	for _, tx := range data.Result {
		totalValue = add(totalValue, tx.Value)
	}

	return totalValue, nil
}

func GetTxCountByAddressByBlockNumber(address string, blockNumber string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	return len(data.Result), nil
}

func GetTxCountByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	for _, tx := range data.Result {
		if tx.Hash == txHash {
			return 1, nil
		}
	}

	return 0, nil
}

func GetTxHashByAddressByBlockNumber(address string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].Hash, nil
}

func GetTxBlockNumberByAddressByBlockNumber(address string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			BlockNumber string `json:"blockNumber"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].BlockNumber, nil
}

func GetTxBlockHashByAddressByBlockNumber(address string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			BlockHash string `json:"blockHash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].BlockHash, nil
}

func GetTxGasPriceByAddressByBlockNumber(address string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			GasPrice string `json:"gasPrice"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].GasPrice, nil
}

func GetTxValueByAddressByBlockNumber(address string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Value string `json:"value"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	var totalValue string
	for _, tx := range data.Result {
		totalValue = add(totalValue, tx.Value)
	}

	return totalValue, nil
}

func GetTxCountByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return 0, err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return 0, err
	}

	for _, tx := range data.Result {
		if tx.Hash == txHash {
			return 1, nil
		}
	}

	return 0, nil
}

func GetTxHashByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Hash string `json:"hash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Hash == txHash {
			return tx.Hash, nil
		}
	}

	return "", nil
}

func GetTxBlockNumberByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			BlockNumber string `json:"blockNumber"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.BlockNumber == txHash {
			return tx.BlockNumber, nil
		}
	}

	return "", nil
}

func GetTxBlockHashByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			BlockHash string `json:"blockHash"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.BlockHash == txHash {
			return tx.BlockHash, nil
		}
	}

	return "", nil
}

func GetTxGasPriceByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			GasPrice string `json:"gasPrice"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.GasPrice == txHash {
			return tx.GasPrice, nil
		}
	}

	return "", nil
}

func GetTxValueByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Value string `json:"value"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	var totalValue string
	for _, tx := range data.Result {
		totalValue = add(totalValue, tx.Value)
	}

	for _, tx := range data.Result {
		if tx.Value == txHash {
			return tx.Value, nil
		}
	}

	return "", nil
}

func GetTxNonceByAddressByBlockNumber(address string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Nonce string `json:"nonce"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].Nonce, nil
}

func GetTxTimestampByAddressByBlockNumber(address string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Timestamp string `json:"timestamp"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].Timestamp, nil
}

func GetTxFromByAddressByBlockNumber(address string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			From string `json:"from"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].From, nil
}

func GetTxToByAddressByBlockNumber(address string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			To string `json:"to"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].To, nil
}

func GetTxGasByAddressByBlockNumber(address string, blockNumber string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Gas string `json:"gas"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	return data.Result[0].Gas, nil
}

func GetTxGasPriceByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			GasPrice string `json:"gasPrice"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.GasPrice == txHash {
			return tx.GasPrice, nil
		}
	}

	return "", nil
}

func GetTxValueByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Value string `json:"value"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Value == txHash {
			return tx.Value, nil
		}
	}

	return "", nil
}

func GetTxNonceByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Nonce string `json:"nonce"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Nonce == txHash {
			return tx.Nonce, nil
		}
	}

	return "", nil
}

func GetTxTimestampByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Timestamp string `json:"timestamp"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Timestamp == txHash {
			return tx.Timestamp, nil
		}
	}

	return "", nil
}

func GetTxFromByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			From string `json:"from"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.From == txHash {
			return tx.From, nil
		}
	}

	return "", nil
}

func GetTxToByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			To string `json:"to"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.To == txHash {
			return tx.To, nil
		}
	}

	return "", nil
}

func GetTxGasByAddressByBlockNumberByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Gas string `json:"gas"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Gas == txHash {
			return tx.Gas, nil
		}
	}

	return "", nil
}

func GetTxValueByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Value string `json:"value"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Value == txHash {
			return tx.Value, nil
		}
	}

	return "", nil
}

func GetTxNonceByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Nonce string `json:"nonce"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Nonce == txHash {
			return tx.Nonce, nil
		}
	}

	return "", nil
}

func GetTxTimestampByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Timestamp string `json:"timestamp"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Timestamp == txHash {
			return tx.Timestamp, nil
		}
	}

	return "", nil
}

func GetTxFromByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			From string `json:"from"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.From == txHash {
			return tx.From, nil
		}
	}

	return "", nil
}

func GetTxToByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			To string `json:"to"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.To == txHash {
			return tx.To, nil
		}
	}

	return "", nil
}

func GetTxGasByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Gas string `json:"gas"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Gas == txHash {
			return tx.Gas, nil
		}
	}

	return "", nil
}

func GetTxValueByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Value string `json:"value"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Value == txHash {
			return tx.Value, nil
		}
	}

	return "", nil
}

func GetTxNonceByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Nonce string `json:"nonce"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Nonce == txHash {
			return tx.Nonce, nil
		}
	}

	return "", nil
}

func GetTxTimestampByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Timestamp string `json:"timestamp"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Timestamp == txHash {
			return tx.Timestamp, nil
		}
	}

	return "", nil
}

func GetTxFromByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			From string `json:"from"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.From == txHash {
			return tx.From, nil
		}
	}

	return "", nil
}

func GetTxToByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			To string `json:"to"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.To == txHash {
			return tx.To, nil
		}
	}

	return "", nil
}

func GetTxGasByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Gas string `json:"gas"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Gas == txHash {
			return tx.Gas, nil
		}
	}

	return "", nil
}

func GetTxValueByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Value string `json:"value"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Value == txHash {
			return tx.Value, nil
		}
	}

	return "", nil
}

func GetTxNonceByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Nonce string `json:"nonce"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Nonce == txHash {
			return tx.Nonce, nil
		}
	}

	return "", nil
}

func GetTxTimestampByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			Timestamp string `json:"timestamp"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.Timestamp == txHash {
			return tx.Timestamp, nil
		}
	}

	return "", nil
}

func GetTxFromByAddressByBlockNumberByTxHashByTxHash(address string, blockNumber string, txHash string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", address, blockNumber, ETHERSCAN_API_KEY))
	if err!= nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return "", err
	}

	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			From string `json:"from"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &data); err!= nil {
		return "", err
	}

	for _, tx := range data.Result {
		if tx.From == txHash {
			return tx.From, nil
		}
	}

	return "", nil
}

func GetTxTo
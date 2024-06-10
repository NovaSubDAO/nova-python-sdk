package optimism

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/NovaSubDAO/nova-sdk/go/pkg/config"
	"github.com/NovaSubDAO/nova-sdk/go/pkg/constants"
	optimismContracts "github.com/NovaSubDAO/nova-sdk/go/pkg/sdk/optimism/abis"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type SdkOptimism struct {
	Config   *config.Config
	Contract *optimismContracts.SavingsDaiCaller
}

func NewSdkOptimism(cfg *config.Config) (*SdkOptimism, error) {
	client, err := ethclient.Dial(cfg.RpcEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Optimism client: %w", err)
	}

	contract, err := optimismContracts.NewSavingsDaiCaller(common.HexToAddress(cfg.SDai), client)
	if err != nil {
		return nil, fmt.Errorf("Failed to instantiate contract caller: %w", err)
	}

	return &SdkOptimism{Config: cfg, Contract: contract}, nil
}

func (sdk *SdkOptimism) getPriceFromInput(input optimismContracts.IMixedRouteQuoterV1QuoteExactInputSingleV2Params) (*big.Int, error) {
	client, err := ethclient.Dial(sdk.Config.RpcEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Error loading client: %w", err)
	}

	contractAddress := common.HexToAddress("0xa4ac92a0F54f1a447c55a4082c90742F5E76Df62")

	contractAbi, err := abi.JSON(strings.NewReader(optimismContracts.MixedRouteQuoterV1ABI))
	if err != nil {
		return nil, fmt.Errorf("Failed to parse contract ABI: %w", err)
	}

	data, err := contractAbi.Pack("quoteExactInputSingleV2", input)
	if err != nil {
		return nil, fmt.Errorf("Failed to pack arguments: %w", err)
	}

	// Setting up the call message
	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}
	ctx := context.Background()

	// Making the call
	output, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to call contract: %w", err)
	}

	// Process the output data (this depends on what the function returns)
	results, err := contractAbi.Unpack("quoteExactInputSingleV2", output)
	if err != nil {
		return nil, fmt.Errorf("Failed to unpack function output: %w", err)
	}

	// Type assert the result as *big.Int
	result, ok := results[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("result type assertion to *big.Int failed")
	}

	return result, nil
}

func (sdk *SdkOptimism) GetPrice(stable constants.Stablecoin) (*big.Int, error) {
	// Packing the input arguments
	stableAddress := constants.StablecoinAddresses[sdk.Config.ChainId][stable]
	tokenIn := common.HexToAddress(stableAddress)
	tokenOut := common.HexToAddress(sdk.Config.SDai)

	input := optimismContracts.IMixedRouteQuoterV1QuoteExactInputSingleV2Params{
		TokenIn:  tokenIn,
		TokenOut: tokenOut,
		Stable:   false,
		AmountIn: big.NewInt(1),
	}

	result, err := sdk.getPriceFromInput(input)
	if err != nil {
		return nil, fmt.Errorf("Failed to get price from input params: %w", err)
	}

	resultFloat := new(big.Float).SetInt(result)
	factorFloat := new(big.Float).SetInt(big.NewInt(1e12))
	decimalsFactorFloat := new(big.Float).SetInt(big.NewInt(1e18))
	price := new(big.Float).Quo(factorFloat, resultFloat)
	price.Mul(price, decimalsFactorFloat)

	priceInt := new(big.Int)
	price.Int(priceInt)

	return priceInt, nil
}

func (sdk *SdkOptimism) GetPosition(stable constants.Stablecoin, address common.Address) (*big.Int, error) {
	balance, err := sdk.Contract.BalanceOf(nil, address)
	if err != nil {
		return big.NewInt(0), err
	}

	price, err := sdk.GetPrice(stable)
	if err != nil {
		return big.NewInt(0), err
	}

	value := new(big.Int).Mul(balance, price)
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(sdk.Config.VaultDecimals)), nil)
	valueNormalized := new(big.Int).Div(value, factor)

	return valueNormalized, nil
}

func (sdk *SdkOptimism) GetTotalValue(stable constants.Stablecoin) (*big.Int, error) {
	totalSupply, err := sdk.Contract.TotalSupply(nil)
	if err != nil {
		return big.NewInt(0), err
	}

	price, err := sdk.GetPrice(stable)
	if err != nil {
		return big.NewInt(0), err
	}

	totalValue := new(big.Int).Mul(totalSupply, price)
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(sdk.Config.VaultDecimals)), nil)
	totalValueNormalized := new(big.Int).Div(totalValue, factor)

	return totalValueNormalized, nil
}

func (sdk *SdkOptimism) GetSlippage(stable constants.Stablecoin, amount *big.Int) (float64, error) {
	// Packing the input arguments
	stableAddress := constants.StablecoinAddresses[sdk.Config.ChainId][stable]
	tokenIn := common.HexToAddress(stableAddress)
	tokenOut := common.HexToAddress(sdk.Config.SDai)

	input := optimismContracts.IMixedRouteQuoterV1QuoteExactInputSingleV2Params{
		TokenIn:  tokenIn,
		TokenOut: tokenOut,
		Stable:   false,
		AmountIn: big.NewInt(1),
	}
	resultOne, err := sdk.getPriceFromInput(input)
	if err != nil {
		return 0, fmt.Errorf("Failed to get price from input params: %w", err)
	}

	input = optimismContracts.IMixedRouteQuoterV1QuoteExactInputSingleV2Params{
		TokenIn:  tokenIn,
		TokenOut: tokenOut,
		Stable:   false,
		AmountIn: amount,
	}
	resultAmount, err := sdk.getPriceFromInput(input)
	if err != nil {
		return 0, fmt.Errorf("Failed to get price from input params: %w", err)
	}

	if amount == big.NewInt(0) || resultOne == big.NewInt(0) {
		return float64(0), nil
	} else {
		resultOneFloat := new(big.Float).SetInt(resultOne)
		resultAmountFloat := new(big.Float).SetInt(resultAmount)
		amountFloat := new(big.Float).SetInt(amount)
		resultAmountFloat.Quo(resultAmountFloat, amountFloat)
		diff := new(big.Float).Sub(resultAmountFloat, resultOneFloat)
		percentageChange, _ := new(big.Float).Quo(diff, resultOneFloat).Float64()
		return percentageChange, nil
	}
}

func (sdk *SdkOptimism) CreateDepositTransaction(stable constants.Stablecoin, fromAddress common.Address, amount *big.Int, referral *big.Int) (string, error) {
	stableAddress := common.HexToAddress(constants.StablecoinAddresses[sdk.Config.ChainId][stable])

	client, err := ethclient.Dial(sdk.Config.RpcEndpoint)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to the Optimism client: %v", err)
	}

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("Failed to get nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("Failed to suggest gas price: %v", err)
	}

	contractAddress := common.HexToAddress(sdk.Config.VaultAddress)

	contractAbi, err := abi.JSON(strings.NewReader(optimismContracts.NovaVaultMetaData.ABI))
	if err != nil {
		return "", fmt.Errorf("Failed to parse contract ABI: %w", err)
	}

	referralUint16 := uint16(referral.Uint64())
	data, err := contractAbi.Pack("deposit", stableAddress, amount, referralUint16)
	if err != nil {
		return "", fmt.Errorf("ABI pack failed: %v", err)
	}

	// Estimating the gas needed for the transaction
	msg := ethereum.CallMsg{From: fromAddress, To: &contractAddress, GasPrice: gasPrice, Value: big.NewInt(0), Data: data}
	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		log.Printf("Gas estimation failed, using fallback gas limit: %v", err)
		gasLimit = 2000000
	}

	tx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), gasLimit, gasPrice, data)

	txJSON, err := json.Marshal(tx)
	if err != nil {
		return "", fmt.Errorf("Failed to marshal transaction: %w", err)
	}

	return string(txJSON), nil
}

func (sdk *SdkOptimism) CreateWithdrawTransaction(stable constants.Stablecoin, fromAddress common.Address, amount *big.Int, referral *big.Int) (string, error) {
	stableAddress := common.HexToAddress(constants.StablecoinAddresses[sdk.Config.ChainId][stable])

	client, err := ethclient.Dial(sdk.Config.RpcEndpoint)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to the Ethereum client: %v", err)
	}

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("Failed to get nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("Failed to suggest gas price: %v", err)
	}

	contractAbi, err := abi.JSON(strings.NewReader(optimismContracts.NovaVaultMetaData.ABI))
	if err != nil {
		return "", fmt.Errorf("Failed to parse contract ABI: %w", err)
	}

	contractAddress := common.HexToAddress(sdk.Config.VaultAddress)

	data, err := contractAbi.Pack("withdraw", stableAddress, amount)
	if err != nil {
		return "", fmt.Errorf("ABI pack failed: %v", err)
	}

	// Estimating the gas needed for the transaction
	msg := ethereum.CallMsg{From: fromAddress, To: &contractAddress, GasPrice: gasPrice, Value: big.NewInt(0), Data: data}
	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		log.Printf("Gas estimation failed, using fallback gas limit: %v", err)
		gasLimit = 2000000
	}

	tx := types.NewTransaction(nonce, contractAddress, big.NewInt(0), gasLimit, gasPrice, data)

	txJSON, err := json.Marshal(tx)
	if err != nil {
		return "", fmt.Errorf("Failed to marshal transaction: %w", err)
	}

	return string(txJSON), nil
}

func (sdk *SdkOptimism) Deposit(stable constants.Stablecoin, assets *big.Int, receiver common.Address, referral big.Int) (*types.Transaction, error) {
	return nil, fmt.Errorf("Not yet implemented")
}

func (sdk *SdkOptimism) Withdraw(stable constants.Stablecoin, assets *big.Int, receiver common.Address, referral big.Int) (*types.Transaction, error) {
	return nil, fmt.Errorf("Not yet implemented")
}

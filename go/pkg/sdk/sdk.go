package sdk

import (
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/NovaSubDAO/nova-sdk/go/pkg/config"
	"github.com/NovaSubDAO/nova-sdk/go/pkg/contracts"
)

type NovaSDK struct {
    Contract *contracts.ContractsCaller
    Client   *ethclient.Client
}

func NewNovaSDK(client *ethclient.Client, contractAddress string) (*NovaSDK, error) {
    contract, err := contracts.NewContractsCaller(common.HexToAddress(contractAddress), client)
    if err != nil {
        return nil, err
    }

    return &NovaSDK{
        Contract: contract,
        Client:   client,
    }, nil
}

func (sdk *NovaSDK) GetPrice() (*big.Int, error) {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal("Error loading configuration:", err)
    }

    factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(cfg.SDaiDecimals)), nil)

    assets, err :=sdk.Contract.ConvertToAssets(nil, factor)
    if err != nil {
        return big.NewInt(0), err
    }

    return assets, nil
}

func (sdk *NovaSDK) GetPosition(address common.Address) (*big.Int, error) {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal("Error loading configuration:", err)
    }

    balance, err := sdk.Contract.BalanceOf(nil, address)
    if err != nil {
        return big.NewInt(0), err
    }

    price, err := sdk.GetPrice()
    if err != nil {
        return big.NewInt(0), err
    }

    value := new(big.Int).Mul(balance, price)
    factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(cfg.SDaiDecimals)), nil)
    valueNormalized := new(big.Int).Div(value, factor)

    return valueNormalized, nil
}


// func (sdk *NovaSDK) GetTotalValue() (*big.Int, error) {
//     cfg, err := config.LoadConfig()
//     if err != nil {
//         log.Fatal("Error loading configuration:", err)
//     }

//     totalSupply, err := sdk.Contract.TotalSupply(nil)
//     if err != nil {
//         return big.NewInt(0), err
//     }

//     price, err := sdk.GetPrice()
//     if err != nil {
//         return big.NewInt(0), err
//     }

//     totalValue := new(big.Int).Mul(totalSupply, price)
//     factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(cfg.SDaiDecimals)), nil)
//     totalValueNormalized := new(big.Int).Div(totalValue, factor)

//     return totalValueNormalized, nil
// }

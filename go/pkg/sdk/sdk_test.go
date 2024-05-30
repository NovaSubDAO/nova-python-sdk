package sdk

import (
    "math/big"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/NovaSubDAO/nova-sdk/go/pkg/config"
)

func setup(t *testing.T) (*NovaSDK, func()) {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        t.Fatalf("Error loading configuration: %v", err)
    }

    // Initialize the Ethereum client
    ethClient, err := ethclient.Dial(cfg.RpcEndpoint)
    if err != nil {
        t.Fatalf("Failed to connect to Ethereum client: %v", err)
    }

    // Initialize the SDK
    novaSDK, err := NewNovaSDK(ethClient, cfg.SDaiAddress)
    if err != nil {
        t.Fatalf("Failed to initialize the Nova SDK: %v", err)
    }

    // Return the initialized SDK and a cleanup function to close the client
    return novaSDK, func() {
        ethClient.Close()
    }
}

func TestGetPrice(t *testing.T) {
    novaSDK, cleanup := setup(t)
    defer cleanup()

    // Use the SDK to get sDAI price
    output, err := novaSDK.GetPrice()
    if err != nil {
        t.Fatalf("Failed to get sDAI price: %v", err)
    }

    expected := big.NewInt(1083353724548051573)

    if output != expected {
        // TODO: uncomment when tests will be run with mainnet forking
        // t.Errorf("Output from GetPrice: %d. It should be %d", output, expected)
        t.Log("This test is to be run with mainnet forking", output)
    }
}

func TestGetPosition(t *testing.T) {
    novaSDK, cleanup := setup(t)
    defer cleanup()

    // Use the SDK to get the position of a user
    addressStr := "0x4aa42145aa6ebf72e164c9bbc74fbd3788045016"
    address := common.HexToAddress(addressStr)
    output, err := novaSDK.GetPosition(address)
    if err != nil {
        t.Fatalf("Failed to get position: %v", err)
    }

    t.Logf("Output from GetPosition: %f", output)
}

// func TestGetTotalValue(t *testing.T) {
//     novaSDK, cleanup := setup(t)
//     defer cleanup()

//     // Use the SDK to get the total value of the vault
//     output, err := novaSDK.GetTotalValue()
//     if err != nil {
//         t.Fatalf("Failed to get position: %v", err)
//     }

//     t.Logf("Output from GetPosition: %f", output)
// }

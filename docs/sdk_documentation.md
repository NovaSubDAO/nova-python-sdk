# SDK Documentation

The NovaVault SDK for Go enables developers to interact with the NovaVault contract for managing stablecoins.
The SDK simplifies tasks like depositing, withdrawing, and querying financial data within a liquidity pool.

## Overview of Functions

This SDK includes the following functions:

- **Deposit**: Deposit stablecoins into the NovaVault contract.
- **Withdraw**: Withdraw stablecoins from the NovaVault contract.
- **GetPrice**: Retrieve the current price of sDAI in the liquidity pool.
- **GetSlippage**: Calculate the current slippage percentage for transactions based on the specified amount.
- **GetPosition**: Retrieve the position value of a given user.

## Function Details

### **Deposit**

Deposits a specified amount of stablecoins into the NovaVault.
This function requires the stablecoin type, amount, and an optional referral identifier.

**Parameters**

- `amount` (int): The amount of stablecoin to deposit.
- `stablecoin` (string): The type of stablecoin to deposit.
- `referral` (int, optional): A referral code for tracking or rewards.

**Example**

```go
package main

import (
    "github.com/NovaSubDAO/nova-sdk/go/pkg/sdk"
    "fmt"
)

func main() {
    err := sdk.Deposit(100, "USDC", 123)
    if err != nil {
        fmt.Println("Error depositing stablecoin:", err)
    }
}
```

### **Withdraw**

Withdraws a specified amount of stablecoins from the NovaVault. It requires the stablecoin type and amount.

**Parameters**

- `amount` (int): The amount of stablecoin to withdraw.
- `stablecoin` (string): The type of stablecoin to withdraw.

**Example**

```go
package main

import (
    "github.com/NovaSubDAO/nova-sdk/go/pkg/sdk"
    "fmt"
)

func main() {
    err := sdk.Withdraw(50, "USDC")
    if err != nil {
        fmt.Println("Error withdrawing stablecoin:", err)
    }
}
```

### **GetPrice**

Fetches the current price of sDAI from the liquidity pool.

**Example**

```go
package main

import (
    "github.com/NovaSubDAO/nova-sdk/go/pkg/sdk"
    "fmt"
)

func main() {
    price, err := sdk.GetPrice()
    if err != nil {
        fmt.Println("Error getting price:", err)
        return
    }
    fmt.Println("Current price of sDAI:", price)
}
```

### **GetSlippage**

Calculates the slippage for a given transaction amount in the liquidity pool.

**Parameters**

- `amount` (int): The transaction amount for which to calculate slippage.

**Example**

```go
package main

import (
    "github.com/NovaSubDAO/nova-sdk/go/pkg/sdk"
    "fmt"
)

func main() {
    slippage, err := sdk.GetSlippage(100)
    if err != nil {
        fmt.Println("Error calculating slippage:", err)
        return
    }
    fmt.Println("Estimated slippage:", slippage)
}
```

### **GetPosition**

Calculates the value of the position of a given user.
This function calculates the balance of sDAI the user, retrieves the current sDAI price, and returns the value of sDAI position held by the user.

**Parameters**

- `address` (common.Address): The blockchain address of the user whose position is being queried.

**Example**

```go
package main

import (
    "github.com/NovaSubDAO/nova-sdk/go/pkg/sdk"
    "github.com/ethereum/go-ethereum/common"
    "fmt"
    "math/big"
)

func main() {
    address := common.HexToAddress("0xYourAddressHere")
    position, err := sdk.GetPosition(address)
    if err != nil {
        fmt.Println("Error retrieving position:", err)
        return
    }

    fmt.Println("Position value:", position)
}
```

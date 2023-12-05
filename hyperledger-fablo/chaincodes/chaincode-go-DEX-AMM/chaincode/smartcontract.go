package chaincode

import (
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type LiquidityPool struct {
	TokenA          float64 `json:"tokenA"`
	TokenB          float64 `json:"tokenB"`
	LastTransaction int64   `json:"lastTx"`
	LiquidityLocked bool    `json:"liquidityLocked"`
	LockedUntil     int64   `json:"lockedUntil"`
	TotalLPTokens   float64 `json:"totalLPTokens"`
}

const (
	MAX_SLIPPAGE       = 0.01
	TRANSACTION_FEE    = 0.001
	MINIMUM_LIQUIDITY  = 10
	RATE_LIMIT_SECONDS = 10
	MIN_LP_TOKENS      = 1
)

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	pool := LiquidityPool{
		TokenA:          10000,
		TokenB:          10000,
		LastTransaction: time.Now().Unix(),
		LiquidityLocked: false,
		LockedUntil:     0,
		TotalLPTokens:   0,
	}
	poolBytes, _ := json.Marshal(pool)
	return ctx.GetStub().PutState("pool", poolBytes)
}

func (s *SmartContract) AddLiquidity(ctx contractapi.TransactionContextInterface, amountA float64, amountB float64, provider string) (float64, error) {
	poolBytes, err := ctx.GetStub().GetState("pool")
	if err != nil || poolBytes == nil {
		return 0, errors.New("pool not found")
	}

	var pool LiquidityPool
	err = json.Unmarshal(poolBytes, &pool)
	if err != nil {
		return 0, errors.New("error unmarshalling pool data")
	}

	// Add tokens to the pool
	pool.TokenA += amountA
	pool.TokenB += amountB

	// Calculate LP tokens to issue - proportional to the increase in liquidity
	var issuedLPTokens float64
	if pool.TotalLPTokens == 0 {
		issuedLPTokens = math.Sqrt(amountA * amountB) // Initial mint
		pool.TotalLPTokens = issuedLPTokens
	} else {
		issuedLPTokens = math.Min(amountA/pool.TokenA, amountB/pool.TokenB) * pool.TotalLPTokens
		pool.TotalLPTokens += issuedLPTokens
	}

	// Update pool data
	poolBytes, _ = json.Marshal(pool)
	err = ctx.GetStub().PutState("pool", poolBytes)
	if err != nil {
		return 0, errors.New("failed to update pool data")
	}

	return issuedLPTokens, nil
}

func (s *SmartContract) Swap(ctx contractapi.TransactionContextInterface, amountA float64, maxSlippage float64, swapper string) (float64, error) {
	poolBytes, err := ctx.GetStub().GetState("pool")
	if err != nil || poolBytes == nil {
		return 0, errors.New("pool not found")
	}

	var pool LiquidityPool
	err = json.Unmarshal(poolBytes, &pool)
	if err != nil {
		return 0, errors.New("error unmarshalling pool data")
	}

	// Calculate amount of TokenB to be swapped
	k := pool.TokenA * pool.TokenB // Constant product
	newTokenA := pool.TokenA + amountA
	newTokenB := k / newTokenA
	amountB := pool.TokenB - newTokenB

	// Apply slippage check
	actualSlippage := (amountB / amountA) - 1
	if math.Abs(actualSlippage) > maxSlippage {
		return 0, errors.New("slippage too high")
	}

	// Apply transaction fee
	fee := amountB * TRANSACTION_FEE
	amountB -= fee

	// Update pool data
	pool.TokenA = newTokenA
	pool.TokenB = newTokenB
	pool.LastTransaction = time.Now().Unix()
	poolBytes, _ = json.Marshal(pool)
	err = ctx.GetStub().PutState("pool", poolBytes)
	if err != nil {
		return 0, errors.New("failed to update pool data")
	}

	return amountB, nil
}

func (s *SmartContract) RemoveLiquidity(ctx contractapi.TransactionContextInterface, lpTokens float64, provider string) (float64, float64, error) {
	if lpTokens < MIN_LP_TOKENS {
		return 0, 0, errors.New("minimum LP tokens not met for withdrawal")
	}

	poolBytes, err := ctx.GetStub().GetState("pool")
	if err != nil || poolBytes == nil {
		return 0, 0, errors.New("pool not found")
	}

	var pool LiquidityPool
	err = json.Unmarshal(poolBytes, &pool)
	if err != nil {
		return 0, 0, errors.New("error unmarshalling pool data")
	}

	// Calculate the amount of TokenA and TokenB to withdraw
	share := lpTokens / pool.TotalLPTokens
	amountA := share * pool.TokenA
	amountB := share * pool.TokenB

	// Update pool data
	pool.TokenA -= amountA
	pool.TokenB -= amountB
	pool.TotalLPTokens -= lpTokens
	poolBytes, _ = json.Marshal(pool)
	err = ctx.GetStub().PutState("pool", poolBytes)
	if err != nil {
		return 0, 0, errors.New("failed to update pool data")
	}

	return amountA, amountB, nil
}

// Additional helper functions or modifications can be added here

// func main() {
// 	err := shim.Start(new(SmartContract))
// 	if err != nil {
// 		fmt.Printf("Error creating new Smart Contract: %s", err)
// 	}
// }

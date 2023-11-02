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
	TokenA float64 `json:"tokenA"`
	TokenB float64 `json:"tokenB"`
	LastTransaction int64 `json:"lastTx"`
	LiquidityLocked bool `json:"liquidityLocked"`
	LockedUntil int64 `json:"lockedUntil"`
	TotalLPTokens float64 `json:"totalLPTokens"`
}

const (
	MAX_SLIPPAGE = 0.01
	TRANSACTION_FEE = 0.001
	MINIMUM_LIQUIDITY = 10
	RATE_LIMIT_SECONDS = 10
	MIN_LP_TOKENS = 1
)

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()

	pool := LiquidityPool {
		TokenA: 10000,
		TokenB: 10000,
		LastTransaction: txTimestamp.Seconds,
		LiquidityLocked: false,
		LockedUntil: 0,
		TotalLPTokens: 0,
	}

	poolBytes, _ := json.Marshal(pool)

	return ctx.GetStub().PutState("pool", poolBytes)
}

func (s *SmartContract) Swap(ctx contractapi.TransactionContextInterface, amountA float64, maxSlippage float64) (float64, error) {

	if maxSlippage > MAX_SLIPPAGE {
		return 0, errors.New("exceeds max allowable slippage")
	}

	poolBytes, _ := ctx.GetStub().GetState("pool")

	if poolBytes == nil {
		return 0, errors.New("pool not found")
	}

	var pool LiquidityPool
	_ = json.Unmarshal(poolBytes, &pool)

	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()

	if txTimestamp.Seconds-pool.LastTransaction < RATE_LIMIT_SECONDS {
		return 0, errors.New("rate limit exceeded")
	}

	invariant := pool.TokenA * pool.TokenB
	amountB := pool.TokenB - (invariant / (pool.TokenA + amountA))

	actualSlippage := (amountB / amountA) - 1

	if math.Abs(actualSlippage) > maxSlippage {
		return 0, errors.New("actual slippage exceeds user-defined max slippage")
	}

	fee := amountB * TRANSACTION_FEE
	amountB -= fee

	txTimestamp2, _ := ctx.GetStub().GetTxTimestamp()

	pool.TokenA += amountA
	pool.TokenB -= amountB
	pool.LastTransaction = txTimestamp2.Seconds

	poolBytes, _ = json.Marshal(pool)
	ctx.GetStub().PutState("pool", poolBytes)

	return amountB, nil
}

func (s *SmartContract) AddLiquidity(ctx contractapi.TransactionContextInterface, amountA float64, amountB float64) (float64, error) {
	poolBytes, _ := ctx.GetStub().GetState("pool")

	if poolBytes == nil {
		return 0, errors.New("pool not found")
	}

	var pool LiquidityPool
	_ = json.Unmarshal(poolBytes, &pool)

	if pool.LiquidityLocked && time.Now().Unix() < pool.LockedUntil {
		return 0, errors.New("liquidity is locked")
	}

	pool.TokenA += amountA
	pool.TokenB += amountB

	totalValueBefore := pool.TotalLPTokens
	pool.TotalLPTokens += amountA
	issuedLPTokens := pool.TotalLPTokens - totalValueBefore

	poolBytes, _ = json.Marshal(pool)
	ctx.GetStub().PutState("pool", poolBytes)

	return issuedLPTokens, nil
}

func (s *SmartContract) RemoveLiquidity(ctx contractapi.TransactionContextInterface, lpTokens float64) (float64, float64, error) {

	if lpTokens < MIN_LP_TOKENS {
		return 0, 0, errors.New("minimum LP tokens not met for withdrawal")
	}

	poolBytes, _ := ctx.GetStub().GetState("pool")

	if poolBytes == nil {
		return 0, 0, errors.New("pool not found")
	}

	var pool LiquidityPool
	_ = json.Unmarshal(poolBytes, &pool)

	if pool.LiquidityLocked && time.Now().Unix() < pool.LockedUntil {
		return 0, 0, errors.New("liquidity is locked")
	}

	if lpTokens > pool.TotalLPTokens {
		return 0, 0, errors.New("insufficient LP tokens in the pool")
	}

	amountA := (lpTokens / pool.TotalLPTokens) * pool.TokenA
	amountB := (lpTokens / pool.TotalLPTokens) * pool.TokenB

	pool.TokenA -= amountA
	pool.TokenB -= amountB
	pool.TotalLPTokens -= lpTokens

	poolBytes, _ = json.Marshal(pool)

	ctx.GetStub().PutState("pool", poolBytes)

	return amountA, amountB, nil
}

// func main() {

// 	chaincode, err := contractapi.NewChaincode(new(SmartContract))

// 	if err != nil {
// 		fmt.Printf("Error create SmartContract chaincode: %s", err.Error())
// 		return
// 	}

// 	if err := chaincode.Start(); err != nil {
// 		fmt.Printf("Error starting SmartContract chaincode: %s", err.Error())
// 	}
// }

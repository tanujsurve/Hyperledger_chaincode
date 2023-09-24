package chaincode

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// NFT represents an art collectible with details
type NFT struct {
	ID    string `json:"id"`
	Owner string `json:"owner"`
	Price int    `json:"price"` // Price is in FabBits
}

func (t *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// Initialization logic, if any.
	return nil
}

func (t *SmartContract) Mint(ctx contractapi.TransactionContextInterface, id string, priceInFabCoin float64) error {
	priceInFabBits := int(priceInFabCoin * 1000) // Convert FabCoin to FabBits

	ownerId, _ := ctx.GetClientIdentity().GetID()

	nft := NFT{
		ID:    id,
		Owner: ownerId,
		Price: priceInFabBits,
	}
	nftBytes, _ := json.Marshal(nft)
	return ctx.GetStub().PutState(id, nftBytes)
}

func (t *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	// NFT transfer logic
	nftBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return err
	}
	nft := NFT{}
	_ = json.Unmarshal(nftBytes, &nft)
	nft.Owner = newOwner
	updatedNftBytes, _ := json.Marshal(nft)
	return ctx.GetStub().PutState(id, updatedNftBytes)
}

// Utility functions to manage the contract's balance in FabBits
func getContractBalance(ctx contractapi.TransactionContextInterface) (int, error) {
	balanceBytes, err := ctx.GetStub().GetState("ContractBalance")
	if err != nil {
		return 0, err
	}
	if balanceBytes == nil {
		return 0, nil // return 0 if balance is not set yet
	}
	return strconv.Atoi(string(balanceBytes))
}

func setContractBalance(ctx contractapi.TransactionContextInterface, balance int) error {
	return ctx.GetStub().PutState("ContractBalance", []byte(strconv.Itoa(balance)))
}

func (t *SmartContract) Withdraw(ctx contractapi.TransactionContextInterface, targetAccount string) error {
	contractBalance, err := getContractBalance(ctx)
	if err != nil {
		return fmt.Errorf("Failed to fetch contract balance: %v", err)
	}
	if contractBalance <= 0 {
		return fmt.Errorf("No balance in the contract")
	}
	// Transfer balance to targetAccount
	targetBalanceBytes, err := ctx.GetStub().GetState(targetAccount)
	if err != nil {
		return fmt.Errorf("Failed to fetch target account balance: %v", err)
	}
	targetBalance, _ := strconv.Atoi(string(targetBalanceBytes))
	newTargetBalance := targetBalance + contractBalance
	err = ctx.GetStub().PutState(targetAccount, []byte(strconv.Itoa(newTargetBalance)))
	if err != nil {
		return fmt.Errorf("Failed to transfer balance to target account: %v", err)
	}
	// Reset contract's balance to 0
	err = setContractBalance(ctx, 0)
	if err != nil {
		return fmt.Errorf("Failed to reset contract balance: %v", err)
	}
	return nil
}

// func main() {
// 	chaincode, err := contractapi.NewChaincode(&SmartContract{})
// 	if err != nil {
// 		fmt.Printf("Error creating NFT chaincode: %s", err.Error())
// 		return
// 	}
// 	if err := chaincode.Start(); err != nil {
// 		fmt.Printf("Error starting NFT chaincode: %s", err.Error())
// 	}
// }

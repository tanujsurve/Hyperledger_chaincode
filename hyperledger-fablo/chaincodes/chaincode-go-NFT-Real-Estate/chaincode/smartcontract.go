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

type PropertyNFT struct {
	PropertyID  string `json:"propertyID"`
	Description string `json:"description"`
	TotalShares int    `json:"totalShares"`
}

type PropertyShare struct {
	ShareID    string `json:"shareID"`
	PropertyID string `json:"propertyID"`
	Owner      string `json:"owner"`
	NumShares  int    `json:"numShares"`
}

// Initialize the Smart Contract with sample data
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	properties := []PropertyNFT{
		{PropertyID: "Prop001", Description: "Luxury Apartment in New York", TotalShares: 1000},
		{PropertyID: "Prop002", Description: "Beach House in California", TotalShares: 500},
	}

	for _, property := range properties {
		propertyAsBytes, _ := json.Marshal(property)
		err := ctx.GetStub().PutState(property.PropertyID, propertyAsBytes)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}
	return nil
}

// Mint a new PropertyNFT
func (s *SmartContract) MintPropertyNFT(ctx contractapi.TransactionContextInterface, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("incorrect number of arguments. expecting 3")
	}

	propertyID := args[0]
	description := args[1]
	totalShares, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid total shares value")
	}

	property := PropertyNFT{
		PropertyID:  propertyID,
		Description: description,
		TotalShares: totalShares,
	}

	propertyAsBytes, _ := json.Marshal(property)
	return ctx.GetStub().PutState(propertyID, propertyAsBytes)
}

// Buy shares in a PropertyNFT
func (s *SmartContract) BuyShares(ctx contractapi.TransactionContextInterface, shareID, propertyID, buyer string, numShares, priceInCrypto int) error {
	propertyAsBytes, err := ctx.GetStub().GetState(propertyID)
	if err != nil {
		return fmt.Errorf("failed to get property: %s", err.Error())
	}
	if propertyAsBytes == nil {
		return fmt.Errorf("property not found")
	}

	var property PropertyNFT
	err = json.Unmarshal(propertyAsBytes, &property)
	if err != nil {
		return fmt.Errorf("failed to unmarshal property: %s", err.Error())
	}

	if numShares < 1 || numShares > property.TotalShares {
		return fmt.Errorf("invalid number of shares requested")
	}

	// Create and store the property share
	share := PropertyShare{
		ShareID:    shareID,
		PropertyID: propertyID,
		Owner:      buyer,
		NumShares:  numShares,
	}
	shareAsBytes, _ := json.Marshal(share)
	return ctx.GetStub().PutState(shareID, shareAsBytes)
}

// Transfer shares from one owner to another
func (s *SmartContract) TransferShares(ctx contractapi.TransactionContextInterface, shareID, currentOwner, newOwner string) error {
	shareAsBytes, err := ctx.GetStub().GetState(shareID)
	if err != nil {
		return fmt.Errorf("failed to get property share: %s", err.Error())
	}
	if shareAsBytes == nil {
		return fmt.Errorf("property share not found")
	}

	var share PropertyShare
	err = json.Unmarshal(shareAsBytes, &share)
	if err != nil {
		return fmt.Errorf("failed to unmarshal property share: %s", err.Error())
	}

	if share.Owner != currentOwner {
		return fmt.Errorf("transfer of shares can only be initiated by the current owner")
	}

	share.Owner = newOwner
	updatedShareAsBytes, _ := json.Marshal(share)
	return ctx.GetStub().PutState(shareID, updatedShareAsBytes)
}

// Main function
// func main() {
// 	chaincode, err := contractapi.NewChaincode(new(SmartContract))
// 	if err != nil {
// 		fmt.Printf("Error creating new Smart Contract: %s", err)
// 		return
// 	}

// 	if err := chaincode.Start(); err != nil {
// 		fmt.Printf("Error starting Smart Contract: %s", err)
// 	}
// }

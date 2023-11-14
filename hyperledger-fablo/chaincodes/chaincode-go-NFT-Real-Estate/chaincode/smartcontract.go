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

func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// func (s *SmartContract) Invoke(ctx contractapi.TransactionContextInterface) error {
// 	function, args := ctx.GetStub().GetFunctionAndParameters()

// 	switch function {
// 		case "mintPropertyNFT":
// 			return s.mintPropertyNFT(ctx, args)
// 		case "buyShares":
// 			return s.buyShares(ctx, args)
// 		case "transferShares":
// 			return s.transferShares(ctx, args)
// 		default:
// 			return fmt.Errorf("Invalid function name.")
// 	}
// }

func (t *SmartContract) MintPropertyNFT(ctx contractapi.TransactionContextInterface, args []string) error {
	// args: propertyID, description
	if len(args) != 2 {
		return fmt.Errorf("Incorrect number of arguments. Expecting 2")
	}

	property := PropertyNFT{
		PropertyID:  args[0],
		Description: args[1],
		TotalShares: 1000, // defaulting to 1,000 shares for fractional ownership
	}

	propertyAsBytes, _ := json.Marshal(property)
	err := ctx.GetStub().PutState(args[0], propertyAsBytes)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to mint property NFT: %s", args[0]))
	}

	return nil
}

func (t *SmartContract) BuyShares(ctx contractapi.TransactionContextInterface, args []string) error {
	// args: shareID, propertyID, owner, numShares
	if len(args) != 4 {
		return fmt.Errorf("Incorrect number of arguments. Expecting 4")
	}

	propertyAsBytes, _ := ctx.GetStub().GetState(args[1])

	if propertyAsBytes == nil {
		return fmt.Errorf("Property not found")
	}

	var property PropertyNFT
	json.Unmarshal(propertyAsBytes, &property)

	requestedShares, _ := strconv.Atoi(args[3])

	if requestedShares < 1 || requestedShares > property.TotalShares {
		return fmt.Errorf("Invalid number of shares requested")
	}

	share := PropertyShare{
		ShareID:    args[0],
		PropertyID: args[1],
		Owner:      args[2],
		NumShares:  requestedShares,
	}

	shareAsBytes, _ := json.Marshal(share)
	err := ctx.GetStub().PutState(args[0], shareAsBytes)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to buy shares: %s", args[0]))
	}

	property.TotalShares -= requestedShares
	propertyAsBytes, _ = json.Marshal(property)
	ctx.GetStub().PutState(args[1], propertyAsBytes)

	return nil
}

func (t *SmartContract) TransferShares(ctx contractapi.TransactionContextInterface, args []string) error {
	// args: shareID, currentOwner, newOwner
	if len(args) != 3 {
		return fmt.Errorf("Incorrect number of arguments. Expecting 3")
	}

	shareAsBytes, _ := ctx.GetStub().GetState(args[0])

	if shareAsBytes == nil {
		return fmt.Errorf("Shares not found")
	}

	var share PropertyShare
	json.Unmarshal(shareAsBytes, &share)

	// Check if the current owner is the one initiating the transfer
	if share.Owner != args[1] {
		return fmt.Errorf("Only the current owner can transfer the shares")
	}

	share.Owner = args[2]
	shareAsBytes, _ = json.Marshal(share)
	err := ctx.GetStub().PutState(args[0], shareAsBytes)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to transfer shares: %s", args[0]))
	}

	return nil
}

// func main() {
// 	err := shim.Start(new(SmartContract))

// 	if err != nil {
// 		fmt.Printf("Error creating new Smart Contract: %s", err)
// 	}
// }

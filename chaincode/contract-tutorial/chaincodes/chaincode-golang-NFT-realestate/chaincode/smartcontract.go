package chaincode

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	sc "github.com/hyperledger/fabric-protos-go/peer"
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

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	switch function {
	case "mintPropertyNFT":
		return s.mintPropertyNFT(APIstub, args)
	case "buyShares":
		return s.buyShares(APIstub, args)
	case "transferShares":
		return s.transferShares(APIstub, args)
	default:
		return shim.Error("Invalid function name.")
	}
}

func (s *SmartContract) mintPropertyNFT(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// args: propertyID, description
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	property := PropertyNFT{
		PropertyID:  args[0],
		Description: args[1],
		TotalShares: 1000, // defaulting to 1,000 shares for fractional ownership
	}

	propertyAsBytes, _ := json.Marshal(property)
	err := APIstub.PutState(args[0], propertyAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to mint property NFT: %s", args[0]))
	}

	return shim.Success(nil)
}

func (s *SmartContract) buyShares(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// args: shareID, propertyID, owner, numShares
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	propertyAsBytes, _ := APIstub.GetState(args[1])
	if propertyAsBytes == nil {
		return shim.Error("Property not found")
	}
	var property PropertyNFT
	json.Unmarshal(propertyAsBytes, &property)

	requestedShares, _ := strconv.Atoi(args[3])
	if requestedShares < 1 || requestedShares > property.TotalShares {
		return shim.Error("Invalid number of shares requested")
	}

	share := PropertyShare{
		ShareID:    args[0],
		PropertyID: args[1],
		Owner:      args[2],
		NumShares:  requestedShares,
	}

	shareAsBytes, _ := json.Marshal(share)
	err := APIstub.PutState(args[0], shareAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to buy shares: %s", args[0]))
	}

	property.TotalShares -= requestedShares
	propertyAsBytes, _ = json.Marshal(property)
	APIstub.PutState(args[1], propertyAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) transferShares(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// args: shareID, currentOwner, newOwner
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	shareAsBytes, _ := APIstub.GetState(args[0])
	if shareAsBytes == nil {
		return shim.Error("Shares not found")
	}

	var share PropertyShare
	json.Unmarshal(shareAsBytes, &share)

	// Check if the current owner is the one initiating the transfer
	if share.Owner != args[1] {
		return shim.Error("Only the current owner can transfer the shares")
	}

	share.Owner = args[2]

	shareAsBytes, _ = json.Marshal(share)
	err := APIstub.PutState(args[0], shareAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to transfer shares: %s", args[0]))
	}

	return shim.Success(nil)
}

// func main() {
// 	err := shim.Start(new(SmartContract))
// 	if err != nil {
// 		fmt.Printf("Error creating new Smart Contract: %s", err)
// 	}
// }

package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Asset struct {
	AssetID     string  `json:"assetID"`
	Name        string  `json:"name"`
	TotalSupply float64 `json:"totalSupply"`
}

type Order struct {
	OrderID   string  `json:"orderID"`
	AssetID   string  `json:"assetID"`
	UserID    string  `json:"userID"`
	OrderType string  `json:"orderType"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	Status    string  `json:"status"`
	Timestamp int64   `json:"timestamp"`
	Expiry    int64   `json:"expiry"`
}

type User struct {
	UserID  string             `json:"userID"`
	Balance map[string]float64 `json:"balance"`
	Role    string             `json:"role"`
}

// Event structures
type OrderMatchedEvent struct {
	BuyOrderID   string  `json:"buyOrderID"`
	SellOrderID  string  `json:"sellOrderID"`
	MatchedQty   float64 `json:"matchedQty"`
	MatchedPrice float64 `json:"matchedPrice"`
}

var ErrInsufficientBalance = fmt.Errorf("insufficient balance")
var ErrInvalidOrderType = fmt.Errorf("invalid order type")
var ErrUnauthorized = fmt.Errorf("unauthorized action")
var ErrOrderExpired = fmt.Errorf("order has expired")

func (s *SmartContract) RegisterUser(ctx contractapi.TransactionContextInterface, userID, role string) error {
	if role != "admin" && role != "user" {
		return fmt.Errorf("invalid role")
	}
	user := User{
		UserID:  userID,
		Balance: make(map[string]float64),
		Role:    role,
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(userID, userJSON)
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, assetID, name string, totalSupply float64) error {
	userID, _ := ctx.GetClientIdentity().GetID()
	userJSON, err := ctx.GetStub().GetState(userID)
	if err != nil {
		return err
	}
	var user User
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		return err
	}
	if user.Role != "admin" {
		return ErrUnauthorized
	}

	asset := Asset{
		AssetID:     assetID,
		Name:        name,
		TotalSupply: totalSupply,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetID, assetJSON)
}

func (s *SmartContract) PlaceOrder(ctx contractapi.TransactionContextInterface, orderID, assetID, orderType string, price, quantity float64, expiry int64) error {
	userID, _ := ctx.GetClientIdentity().GetID()
	userJSON, err := ctx.GetStub().GetState(userID)
	if err != nil {
		return err
	}
	var user User
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		return err
	}

	if orderType != "buy" && orderType != "sell" {
		return ErrInvalidOrderType
	}

	if orderType == "buy" {
		requiredBalance := price * quantity
		if user.Balance["ETH"] < requiredBalance {
			return ErrInsufficientBalance
		}
	} else if orderType == "sell" {
		if user.Balance[assetID] < quantity {
			return ErrInsufficientBalance
		}
	}

	order := Order{
		OrderID:   orderID,
		AssetID:   assetID,
		UserID:    userID,
		OrderType: orderType,
		Price:     price,
		Quantity:  quantity,
		Status:    "open",
		Timestamp: time.Now().Unix(),
		Expiry:    expiry,
	}
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return err
	}

	s.matchOrder(ctx, &order)

	if order.Quantity > 0 {
		return ctx.GetStub().PutState(orderID, orderJSON)
	}

	return nil
}

func (s *SmartContract) matchOrder(ctx contractapi.TransactionContextInterface, order *Order) error {
	// Placeholder for matching logic...

	// Sample matched event
	matchEvent := OrderMatchedEvent{
		BuyOrderID:   "sampleBuyOrderID",
		SellOrderID:  "sampleSellOrderID",
		MatchedQty:   10.0,
		MatchedPrice: 100.0,
	}

	matchEventBytes, _ := json.Marshal(matchEvent)

	err := ctx.GetStub().SetEvent("OrderMatched", matchEventBytes)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) CancelOrder(ctx contractapi.TransactionContextInterface, orderID string) error {
	orderJSON, err := ctx.GetStub().GetState(orderID)
	if err != nil {
		return err
	}
	var order Order
	err = json.Unmarshal(orderJSON, &order)
	if err != nil {
		return err
	}

	userID, _ := ctx.GetClientIdentity().GetID()
	if order.UserID != userID {
		return ErrUnauthorized
	}

	if order.Expiry < time.Now().Unix() {
		return ErrOrderExpired
	}

	order.Status = "cancelled"
	updatedOrderJSON, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(orderID, updatedOrderJSON)
}

// func main() {
// 	chaincode, err := contractapi.NewChaincode(new(SmartContract))
// 	if err != nil {
// 		fmt.Printf("Error creating chaincode: %s", err.Error())
// 		return
// 	}

// 	if err := chaincode.Start(); err != nil {
// 		fmt.Printf("Error starting chaincode: %s", err.Error())
// 	}
// }

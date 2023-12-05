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

func (s *SmartContract) PlaceOrder(ctx contractapi.TransactionContextInterface, orderID, assetID, userID, orderType string, price, quantity float64, expiry int64) error {
	// Validate order type
	if orderType != "buy" && orderType != "sell" {
		return ErrInvalidOrderType
	}

	// Check for valid quantity and price
	if quantity <= 0 || price <= 0 {
		return fmt.Errorf("invalid quantity or price")
	}

	// Create and store the order
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
		return fmt.Errorf("error marshalling order: %s", err)
	}

	err = ctx.GetStub().PutState(orderID, orderJSON)
	if err != nil {
		return fmt.Errorf("error storing order: %s", err)
	}

	// Attempt to match the order with existing orders
	return s.matchOrder(ctx, &order)
}

func (s *SmartContract) CancelOrder(ctx contractapi.TransactionContextInterface, orderID string) error {
	// Retrieve the order
	orderJSON, err := ctx.GetStub().GetState(orderID)
	if err != nil {
		return fmt.Errorf("error retrieving order: %s", err)
	}
	if orderJSON == nil {
		return fmt.Errorf("order not found: %s", orderID)
	}

	var order Order
	err = json.Unmarshal(orderJSON, &order)
	if err != nil {
		return fmt.Errorf("error unmarshalling order: %s", err)
	}

	// Check if the order can be cancelled
	if order.Status != "open" {
		return fmt.Errorf("only open orders can be cancelled")
	}
	if time.Now().Unix() > order.Expiry {
		return ErrOrderExpired
	}

	// Cancel the order
	order.Status = "cancelled"
	updatedOrderJSON, _ := json.Marshal(order)
	return ctx.GetStub().PutState(orderID, updatedOrderJSON)
}

func (s *SmartContract) matchOrder(ctx contractapi.TransactionContextInterface, order *Order) error {
	// Placeholder logic for matching orders
	// In a real implementation, this would involve:
	// - Querying the ledger for matching buy/sell orders
	// - Matching orders based on price and quantity
	// - Updating the ledger with the matched orders
	// - Emitting an OrderMatchedEvent

	// For demonstration purposes, let's emit a sample matched event
	matchEvent := OrderMatchedEvent{
		BuyOrderID:   "buyOrder123",
		SellOrderID:  "sellOrder123",
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

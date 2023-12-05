package chaincode

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type AuctionItem struct {
	ItemID       string    `json:"itemId"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	StartingBid  float64   `json:"startingBid"`
	EndTime      time.Time `json:"endTime"`
	Status       string    `json:"status"`
	SellerID     string    `json:"sellerId"`
	ReservePrice float64   `json:"reservePrice"`
}

type Bid struct {
	ItemID        string    `json:"itemId"`
	BidderID      string    `json:"bidderId"`
	ActualBid     float64   `json:"actualBid"`
	SealedBid     string    `json:"sealedBid"`
	AttachedValue float64   `json:"attachedValue"`
	Timestamp     time.Time `json:"timestamp"`
	IsWithdrawn   bool      `json:"isWithdrawn"`
}

// StartAuction initializes a new auction
func (cc *SmartContract) StartAuction(ctx contractapi.TransactionContextInterface, args []string) error {
	// Example auction item details
	item := AuctionItem{
		ItemID:       "AucItem123",
		Name:         "Vintage Watch",
		Description:  "A rare vintage watch from 1950",
		StartingBid:  100.0,
		EndTime:      time.Now().Add(48 * time.Hour), // 48 hours from now
		Status:       "ongoing",
		SellerID:     "Seller123",
		ReservePrice: 200.0,
	}

	itemBytes, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("error marshalling auction item: %s", err.Error())
	}

	return ctx.GetStub().PutState(item.ItemID, itemBytes)
}

// PlaceBid allows a bidder to place a bid on an auction item
func (cc *SmartContract) PlaceBid(ctx contractapi.TransactionContextInterface, args []string) error {
	// Example bid details
	bid := Bid{
		ItemID:        "AucItem123",
		BidderID:      "Bidder123",
		ActualBid:     150.0,
		SealedBid:     "encrypted-sealed-bid",
		AttachedValue: 200.0,
		Timestamp:     time.Now(),
		IsWithdrawn:   false,
	}

	bidBytes, err := json.Marshal(bid)
	if err != nil {
		return fmt.Errorf("error marshalling bid: %s", err.Error())
	}

	return ctx.GetStub().PutState("BID_"+bid.ItemID+"_"+bid.BidderID, bidBytes)
}

// EndAuction closes an ongoing auction
func (cc *SmartContract) EndAuction(ctx contractapi.TransactionContextInterface, args []string) error {
	itemID := "AucItem123" // Example item ID

	itemBytes, err := ctx.GetStub().GetState(itemID)
	if err != nil {
		return fmt.Errorf("error retrieving auction item: %s", err.Error())
	}

	var item AuctionItem
	err = json.Unmarshal(itemBytes, &item)
	if err != nil {
		return fmt.Errorf("error unmarshalling auction item: %s", err.Error())
	}

	if time.Now().Before(item.EndTime) {
		return fmt.Errorf("auction cannot be ended before its end time")
	}

	item.Status = "ended"
	updatedItemBytes, _ := json.Marshal(item)
	return ctx.GetStub().PutState(item.ItemID, updatedItemBytes)
}

// RevealWinner calculates the winner of the auction
func (cc *SmartContract) RevealWinner(ctx contractapi.TransactionContextInterface, args []string) (*Bid, error) {
	itemID := "AucItem123" // Example item ID

	// Fetch all bids for the item and sort them by ActualBid in descending order
	allBids := getAllBidsForItem(ctx, itemID)
	sort.Slice(allBids, func(i, j int) bool {
		return allBids[i].ActualBid > allBids[j].ActualBid
	})

	// Determine the winning bid based on Vickrey auction rules
	if len(allBids) < 2 {
		return nil, errors.New("not enough bids to determine a winner")
	}

	winningBid := allBids[0]
	secondHighestBid := allBids[1]

	// The winning price is the second-highest bid
	winningBid.ActualBid = secondHighestBid.ActualBid

	return &winningBid, nil
}

// ClaimExcessValue handles returning excess bid value to the winning bidder
func (cc *SmartContract) ClaimExcessValue(ctx contractapi.TransactionContextInterface, args []string) error {
	// Example item and bidder IDs
	itemID := "AucItem123"
	bidderID := "Bidder123"

	winningBid, err := cc.RevealWinner(ctx, []string{itemID})
	if err != nil {
		return err
	}

	if winningBid.BidderID != bidderID {
		return errors.New("only the winning bidder can claim excess value")
	}

	// Calculate excess value
	excessValue := winningBid.AttachedValue - winningBid.ActualBid
	fmt.Printf("Excess value of %f returned to bidder %s\n", excessValue, bidderID)

	return nil
}

// getAllBidsForItem fetches all bids for a given item
func getAllBidsForItem(ctx contractapi.TransactionContextInterface, itemID string) []Bid {
	// For simplicity, this example returns a hardcoded list of bids
	return []Bid{
		{ItemID: itemID, BidderID: "Bidder123", ActualBid: 150.0, AttachedValue: 200.0, Timestamp: time.Now()},
		{ItemID: itemID, BidderID: "Bidder456", ActualBid: 250.0, AttachedValue: 300.0, Timestamp: time.Now()},
	}
}

// func main() {
// 	chaincode, err := contractapi.NewChaincode(new(SmartContract))
// 	if err != nil {
// 		fmt.Printf("Error creating chaincode: %s", err)
// 		return
// 	}

// 	if err := chaincode.Start(); err != nil {
// 		fmt.Printf("Error starting chaincode: %s", err)
// 	}
// }

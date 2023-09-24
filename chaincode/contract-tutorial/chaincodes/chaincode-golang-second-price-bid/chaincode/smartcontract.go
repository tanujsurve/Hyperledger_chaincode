package chaincode

import (
	"encoding/json"
	"errors"
	"fmt"
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

type NFT struct {
	NFTID   string `json:"nftId"`
	ItemID  string `json:"itemId"`
	OwnerID string `json:"ownerId"`
}

func (cc *SmartContract) StartAuction(ctx contractapi.TransactionContextInterface, item AuctionItem) error {
	// Check if auction for the item already exists
	existingItemBytes, _ := ctx.GetStub().GetState(item.ItemID)
	if existingItemBytes != nil {
		return errors.New("auction item already exists")
	}

	item.Status = "ongoing"
	itemBytes, err := json.Marshal(item)
	if err != nil {
		return errors.New("failed to marshal auction item")
	}

	return ctx.GetStub().PutState(item.ItemID, itemBytes)
}

func (cc *SmartContract) PlaceBid(ctx contractapi.TransactionContextInterface, bid Bid) error {
	// Fetch the auction item
	itemBytes, err := ctx.GetStub().GetState(bid.ItemID)
	if err != nil || itemBytes == nil {
		return errors.New("auction item not found")
	}

	var item AuctionItem
	json.Unmarshal(itemBytes, &item)

	if item.Status != "ongoing" {
		return errors.New("auction is not ongoing")
	}

	// ... (rest of the logic)

	return nil
}

func (cc *SmartContract) EndAuction(ctx contractapi.TransactionContextInterface, itemID string) error {
	// Fetch the auction item
	itemBytes, err := ctx.GetStub().GetState(itemID)
	if err != nil || itemBytes == nil {
		return errors.New("auction item not found")
	}

	var item AuctionItem
	json.Unmarshal(itemBytes, &item)

	if item.Status != "ongoing" {
		return errors.New("auction is not ongoing")
	}

	if time.Now().Before(item.EndTime) {
		return errors.New("auction has not yet reached its end time")
	}

	item.Status = "ended"
	updatedItemBytes, err := json.Marshal(item)
	if err != nil {
		return errors.New("failed to marshal updated auction item")
	}

	return ctx.GetStub().PutState(item.ItemID, updatedItemBytes)
}

func (cc *SmartContract) RevealWinner(ctx contractapi.TransactionContextInterface, itemID string) (*Bid, error) {
	// Fetch the auction item
	itemBytes, err := ctx.GetStub().GetState(itemID)
	if err != nil || itemBytes == nil {
		return nil, errors.New("auction item not found")
	}

	var item AuctionItem
	json.Unmarshal(itemBytes, &item)

	if item.Status != "ended" {
		return nil, errors.New("auction hasn't ended yet")
	}

	highestBid := float64(0)
	var winningBid Bid

	// Fetch all bids for the item (this is a placeholder - a real implementation would query the ledger)
	allBids := getAllBidsForItem(ctx, itemID)

	for _, bid := range allBids {
		if !bid.IsWithdrawn && bid.ActualBid > highestBid {
			highestBid = bid.ActualBid
			winningBid = bid
		}
	}

	return &winningBid, nil
}

func (cc *SmartContract) ClaimNFTAndExcessValue(ctx contractapi.TransactionContextInterface, itemID string, bidderID string) error {
	winningBid, err := cc.RevealWinner(ctx, itemID)
	if err != nil {
		return err
	}

	if winningBid.BidderID != bidderID {
		return errors.New("only the winning bidder can claim the NFT and excess value")
	}

	// Issue the NFT
	nft := NFT{
		NFTID:   fmt.Sprintf("NFT_%s", itemID),
		ItemID:  itemID,
		OwnerID: bidderID,
	}
	nftBytes, _ := json.Marshal(nft)
	ctx.GetStub().PutState(nft.NFTID, nftBytes)

	// Calculate and handle returning excess value
	excessValue := winningBid.AttachedValue - winningBid.ActualBid

	// This is a placeholder. In a real-world scenario, you'd handle the return of this value through transactions or token transfers.
	fmt.Printf("Returning excess value of %f to bidder %s\n", excessValue, bidderID)

	return nil
}

// Placeholder function to fetch all bids for an item
func getAllBidsForItem(ctx contractapi.TransactionContextInterface, itemID string) []Bid {
	// This function should query the ledger to get all bids related to an item.
	// For this example, we'll just return an empty list.
	return []Bid{}
}

// func main() {
// 	// Entry point for the chaincode
// }

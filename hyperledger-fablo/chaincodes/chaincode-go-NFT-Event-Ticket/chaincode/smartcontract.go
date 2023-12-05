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

type Ticket struct {
	TicketID   string `json:"ticket_id"`
	EventID    string `json:"event_id"`
	Seat       string `json:"seat"`
	Owner      string `json:"owner"`
	IsTraded   bool   `json:"is_traded"`
	IsAttended bool   `json:"is_attended"`
}

type User struct {
	UserID       string `json:"user_id"`
	TicketsOwned int    `json:"tickets_owned"`
	RewardPoints int    `json:"reward_points"`
}

type Event struct {
	EventID   string `json:"event_id"`
	EventName string `json:"event_name"`
	Date      string `json:"date"`
	Location  string `json:"location"`
	IsSpecial bool   `json:"is_special"`
}

func (t *SmartContract) MintTicket(ctx contractapi.TransactionContextInterface, args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("expected 4 arguments: TicketID, EventID, Seat, Owner")
	}

	ticket := Ticket{
		TicketID:   args[0],
		EventID:    args[1],
		Seat:       args[2],
		Owner:      args[3],
		IsTraded:   false,
		IsAttended: false,
	}
	ticketAsBytes, _ := json.Marshal(ticket)
	return ctx.GetStub().PutState(ticket.TicketID, ticketAsBytes)
}

func (t *SmartContract) BuyTicket(ctx contractapi.TransactionContextInterface, args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("expected 4 arguments for buying a ticket")
	}

	ticketAsBytes, _ := ctx.GetStub().GetState(args[0])
	if ticketAsBytes == nil {
		return fmt.Errorf("ticket not found")
	}

	var ticket Ticket
	json.Unmarshal(ticketAsBytes, &ticket)
	ticket.Owner = args[3]
	ticket.IsTraded = true

	ticketAsBytes, _ = json.Marshal(ticket)
	return ctx.GetStub().PutState(ticket.TicketID, ticketAsBytes)
}

func (t *SmartContract) MarkAttendance(ctx contractapi.TransactionContextInterface, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument: TicketID")
	}

	ticketID := args[0]
	ticketAsBytes, _ := ctx.GetStub().GetState(ticketID)
	if ticketAsBytes == nil {
		return fmt.Errorf("ticket not found")
	}

	var ticket Ticket
	json.Unmarshal(ticketAsBytes, &ticket)
	ticket.IsAttended = true

	ticketAsBytes, _ = json.Marshal(ticket)
	return ctx.GetStub().PutState(ticket.TicketID, ticketAsBytes)
}

func (t *SmartContract) RegisterEvent(ctx contractapi.TransactionContextInterface, args []string) error {
	if len(args) != 5 {
		return fmt.Errorf("expected 5 arguments: EventID, EventName, Date, Location, IsSpecial")
	}

	isSpecial, err := strconv.ParseBool(args[4])
	if err != nil {
		return fmt.Errorf("invalid value for IsSpecial. Expected true or false")
	}

	event := Event{
		EventID:   args[0],
		EventName: args[1],
		Date:      args[2],
		Location:  args[3],
		IsSpecial: isSpecial,
	}
	eventAsBytes, _ := json.Marshal(event)
	return ctx.GetStub().PutState(event.EventID, eventAsBytes)
}

// RedeemRewardPoints, TradeInTicketsForSpecial, and other functions as needed

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

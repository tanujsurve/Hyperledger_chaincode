package chaincode

import (
	"encoding/json"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	pb "github.com/hyperledger/fabric-protos-go/peer"
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

func (t *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "MintTicket" {
		return t.MintTicket(stub, args)
	} else if function == "BuyTicket" {
		return t.BuyTicket(stub, args)
	} else if function == "RedeemRewardPoints" {
		return t.RedeemRewardPoints(stub, args)
	} else if function == "MarkAttendance" {
		return t.MarkAttendance(stub, args)
	} else if function == "TradeInTicketsForSpecial" {
		return t.TradeInTicketsForSpecial(stub, args)
	} else if function == "RegisterEvent" {
		return t.RegisterEvent(stub, args)
	}
	return shim.Error("Invalid function name.")
}

func (t *SmartContract) MintTicket(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 4 {
		return shim.Error("Expected 4 arguments: TicketID, EventID, Seat, Owner")
	}

	ticket := Ticket{
		TicketID: args[0],
		EventID:  args[1],
		Seat:     args[2],
		Owner:    args[3],
	}

	ticketAsBytes, _ := json.Marshal(ticket)
	stub.PutState(ticket.TicketID, ticketAsBytes)

	return shim.Success(nil)
}

func (t *SmartContract) BuyTicket(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// ... Implementation as shown previously ...

	return shim.Success(nil)
}

func (t *SmartContract) RedeemRewardPoints(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// ... Implementation as shown previously ...

	return shim.Success(nil)
}

func (t *SmartContract) MarkAttendance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Expected 1 argument: TicketID")
	}

	ticketID := args[0]
	ticketAsBytes, _ := stub.GetState(ticketID)
	if ticketAsBytes == nil {
		return shim.Error("Ticket not found.")
	}

	var ticket Ticket
	json.Unmarshal(ticketAsBytes, &ticket)
	ticket.IsAttended = true

	ticketAsBytes, _ = json.Marshal(ticket)
	stub.PutState(ticket.TicketID, ticketAsBytes)

	return shim.Success(nil)
}

func (t *SmartContract) TradeInTicketsForSpecial(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// ... Implementation as shown previously ...

	return shim.Success(nil)
}

func (t *SmartContract) RegisterEvent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 5 {
		return shim.Error("Expected 5 arguments: EventID, EventName, Date, Location, IsSpecial")
	}

	isSpecial, err := strconv.ParseBool(args[4])
	if err != nil {
		return shim.Error("Invalid value for IsSpecial. Expected true or false.")
	}

	event := Event{
		EventID:   args[0],
		EventName: args[1],
		Date:      args[2],
		Location:  args[3],
		IsSpecial: isSpecial,
	}

	eventAsBytes, _ := json.Marshal(event)
	stub.PutState(event.EventID, eventAsBytes)

	return shim.Success(nil)
}

// func main() {
// 	err := shim.Start(new(SmartContract))
// 	if err != nil {
// 		fmt.Printf("Error starting EventTicket chaincode: %s", err)
// 	}
// }

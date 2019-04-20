package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hyperledger/sawtooth-sdk-go/logging"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/processor_pb2"
)

var logger *logging.Logger = logging.Get()

// Hexdigest returns a sha512 hash of the input as a string
func Hexdigest(str string) string {
	hash := sha512.New()
	hash.Write([]byte(str))
	hashBytes := hash.Sum(nil)
	return strings.ToLower(hex.EncodeToString(hashBytes))
}

// CookiejarHandler is the handler for the cookiejar transaction family processor
// The handlers implements the SDK's TransactionHandler interface: https://github.com/hyperledger/sawtooth-sdk-go/blob/master/processor/handler.go
type CookiejarHandler struct {
	namespace string
}

// getAddress returns an address consisting of the name space prefix and 64 characters of the provided name
func (h *CookiejarHandler) getAddress(name string) string {
	hashedName := Hexdigest(name)
	return h.namespace + hashedName[:64]
}

// FamilyName returns the name of the transaction family this handler processes
func (h *CookiejarHandler) FamilyName() string {
	return familyName
}

// FamilyVersions return the versions of the transaction processor this handler can process
func (h *CookiejarHandler) FamilyVersions() []string {
	return []string{"1.0"}
}

// Namespaces returns all the handler's namespaces
func (h *CookiejarHandler) Namespaces() []string {
	return []string{h.namespace}
}

// Apply is the single method where all the business logic for a transaction family is defined. The method will be called by the
// transaction processor upon receiving a TpProcessRequest that the handler understands and will pass in the TpProcessRequest and an initialized
// instance of the Context type.
func (h *CookiejarHandler) Apply(r *processor_pb2.TpProcessRequest, ctx *processor.Context) error {
	// Get the sender's public key
	fromKey := r.GetHeader().GetSignerPublicKey()

	// Decode the payload, which is in csv format
	// Field 0 represents the requested action
	// Field 1 represents the amount
	payloadList := strings.Split(string(r.GetPayload()), ",")
	action := payloadList[0]
	amount, err := strconv.Atoi(payloadList[1]) // Convert to int
	if err != nil {
		return &processor.InternalError{Msg: fmt.Sprintf("Couldn't parse amount : %v", err)}
	}

	logger.Debugf("Action: %s, Amount: %d\n", action, amount)

	// Process action
	switch action {
	case "bake":
		if err := h.bake(ctx, amount, fromKey); err != nil {
			return err
		}
	case "eat":
		if err := h.eat(ctx, amount, fromKey); err != nil {
			return err
		}
	case "clear":
		if err := h.empty(ctx, fromKey); err != nil {
			return err
		}
	default:
		logger.Debugf("Invalid action")
		return &processor.InvalidTransactionError{Msg: fmt.Sprintf("Invalid Action: '%v'", action)}
	}

	return nil
}

// bake will register the provided amount of cookies in a cookiejar
func (h *CookiejarHandler) bake(ctx *processor.Context, amount int, fromKey string) error {
	// Get the composite address for the sender's public key
	address := h.getAddress(fromKey)
	logger.Debugf("Got the key %s and cookiejar address %s", fromKey, address)

	// Get the current state
	state, err := ctx.GetState([]string{address})
	if err != nil {
		return err
	}

	// Get the current amount of cookies if the address exists
	cookies := 0
	c, ok := state[address]
	if ok {
		cookies, _ = strconv.Atoi(string(c))
	}

	// Update the new state to current cookies + amount, as a slice of bytes
	state[address] = []byte(strconv.Itoa(cookies + amount))

	// Store the new state
	if _, err := ctx.SetState(state); err != nil {
		return err
	} // In a real app binary.LittleIndian might be a better way to deal with bytes <-> int

	// Launch an event
	if err := ctx.AddEvent(
		"cookiejar/bake",
		[]processor.Attribute{processor.Attribute{"cookies-baked", strconv.Itoa(amount)}},
		nil,
	); err != nil {
		return err
	}

	return nil
}

// eat updates a cookiejar by deducting the provided amount of cookies
func (h *CookiejarHandler) eat(ctx *processor.Context, amount int, fromKey string) error {
	// Get the composite address for the sender's public key
	address := h.getAddress(fromKey)
	logger.Debugf("Got the key %s and cookiejar address %s", fromKey, address)

	// Get the current state
	state, err := ctx.GetState([]string{address})
	if err != nil {
		return err
	}

	// Get the current state for the address
	c, ok := state[address]
	if !ok {
		// The address doesn't exist, so we'll return with an error
		logger.Errorf("No cookie jar with the key %s", address)
		return &processor.InternalError{Msg: "Invalid cookie jar"}
	}

	cookies, _ := strconv.Atoi(string(c)) // convert to int
	if cookies < amount {
		// Not enough of cookies, return an error
		logger.Error("Not enough of cookies in the jar")
		return &processor.InvalidTransactionError{Msg: "Not enough of cookies in the jar"}
	}

	// Update the state to current amount of cookies - amount
	state[address] = []byte(strconv.Itoa(cookies - amount))

	// Store the new state
	addresses, err := ctx.SetState(state)
	if err != nil {
		return &processor.InternalError{Msg: fmt.Sprintf("Couldn't update state: %v", err)}
	}

	// Check whether addresses is empty
	if len(addresses) == 0 {
		return &processor.InternalError{Msg: "No addresses in set response"}
	}

	// Launch an event
	if err := ctx.AddEvent(
		"cookiejar/eat",
		[]processor.Attribute{processor.Attribute{"cookies-ate", strconv.Itoa(amount)}},
		nil,
	); err != nil {
		return &processor.InternalError{Msg: fmt.Sprintf("Couldn't publish event: %v", err)}
	}

	return nil
}

// empty clears a cookiejar
func (h *CookiejarHandler) empty(ctx *processor.Context, fromKey string) error {
	// Get the composite address for the sender's public key
	address := h.getAddress(fromKey)
	logger.Debugf("Got the key %s and cookiejar address %s", fromKey, address)

	// Get the current state
	state, err := ctx.GetState([]string{address})
	if err != nil {
		return err
	}

	// Check if the address exists. We don't care about it's current value
	_, ok := state[address]
	if !ok {
		log.Printf("No cookie jar with the key %s", address)
		return &processor.InternalError{
			Msg: "Invalid cookie jar",
		}
	}

	// Set the new state to 0 and store the state
	state[address] = []byte(strconv.Itoa(0))
	if _, err := ctx.SetState(state); err != nil {
		return err
	}

	return nil
}

// NewCookiejarHandler returns an initialized CookiejarHandler
func NewCookiejarHandler() *CookiejarHandler {
	return &CookiejarHandler{
		namespace: Hexdigest(familyName)[:6],
	}
}

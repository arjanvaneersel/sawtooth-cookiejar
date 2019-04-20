package main

import (
	gobytes "bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/batch_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
	yaml "gopkg.in/yaml.v2"
)

const familyName = "cookiejar"

// hexdigist returns a string version of the sha512 hash of the input
func hexdigest(str string) string {
	hash := sha512.New()
	hash.Write([]byte(str))
	hashBytes := hash.Sum(nil)
	return strings.ToLower(hex.EncodeToString(hashBytes))
}

// CookiejarClient is the client object which allows communication with the sawtooth network
type CookiejarClient struct {
	url    string
	signer *signing.Signer
}

// getPrefix returns the 6 character prefix based upon the transaction family name
func (c *CookiejarClient) getPrefix() string {
	return hexdigest(familyName)[:6]
}

// getAddress returns a composite key based upon the namespace prefix and the user's address
func (c *CookiejarClient) getAddress() string {
	hashedName := hexdigest(c.signer.GetPublicKey().AsHex())
	return c.getPrefix() + hashedName[:64]
}

// sendRequest sends the request to the Sawtooth network
func (c *CookiejarClient) sendRequest(suffix, contentType string, data []byte) ([]byte, error) {
	// Create the url
	var url string
	if strings.HasPrefix(c.url, "http://") {
		url = fmt.Sprintf("%s/%s", c.url, suffix)
	} else {
		url = fmt.Sprintf("http://%s/%s", c.url, suffix)
	}

	// Send the request and get the response
	var response *http.Response
	var err error
	if len(data) > 0 {
		// If there is data, we'll send a POST request
		response, err = http.Post(url, contentType, gobytes.NewBuffer(data))
	} else {
		// Else we'll send a GET request
		response, err = http.Get(url)
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to REST API: %v", err)
	}

	// Check for potential errors
	if response.StatusCode == http.StatusNotFound {
		logger.Debug(err.Error())
		return nil, fmt.Errorf("Not found")
	} else if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("Error %d: %s", response.StatusCode, response.Status)
	}
	defer response.Body.Close() // Ensure the body will be closed once done

	// Get the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %v", err)
	}

	return body, nil
}

// getResponseMap returns a response as a map
func (c *CookiejarClient) getResponseMap(response []byte) (map[interface{}]interface{}, error) {
	responseMap := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(response, &responseMap); err != nil {
		return nil, err
	}

	return responseMap, nil
}

// getData returns the data element (from the response map) in a result
func (c *CookiejarClient) getData(response []byte, i uint) (map[interface{}]interface{}, error) {
	responseMap, err := c.getResponseMap(response)
	if err != nil {
		return nil, err
	}

	data := responseMap["data"].([]interface{})[i].(map[interface{}]interface{})
	return data, nil
}

// getStatus reads the status of a transaction from the Sawtooth network
func (c *CookiejarClient) getStatus(batchID string, wait uint) (string, error) {
	// Create the URL suffix
	suffix := fmt.Sprintf("%s?id=%s&wait=%d", "batch_statuses", batchID, wait)

	// Send the request
	response, err := c.sendRequest(suffix, "", []byte{})
	if err != nil {
		return "", err
	}

	// Get the data element from the respone's response map
	data, err := c.getData(response, 0)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %v", err)
	}

	// Return the status as a string
	return data["status"].(string), nil
}

// waitForStatus will wait and keep probing whether a transaction's status changed from PENDING or until timeout
func (c *CookiejarClient) waitForStatus(batchID string, timeout uint) (string, error) {
	// Create a go channel for a response and error
	resChan := make(chan map[interface{}]interface{}, 1)
	errChan := make(chan error, 1)

	// Launch a goroutine
	go func() {
		// Create a suffix
		suffix := fmt.Sprintf("batch_statuses?id=%s&wait=%d", batchID, timeout)

		// Get the status from the server
		b, err := c.sendRequest(suffix, "application/octet-stream", []byte{})
		if err != nil {
			errChan <- err
			return
		}

		// Store the whole response
		res, err := c.getData(b, 0)
		if err != nil {
			errChan <- err
			return
		}

		// Convert the status to a string
		status := res["status"].(string)

		// If a timeout is set and the status is PENDING, keep polling
		for timeout > 0 && status == "PENDING" {
			var err error
			status, err = c.getStatus(batchID, timeout)
			if err != nil {
				// Something went wrong, send the error via the channel and end the goroutine
				errChan <- err
				return
			}
		}

		// Send the response to the channel
		resChan <- res
	}()

	// Keep waiting until we got some response from the goroutine or a timeout
	select {
	case res := <-resChan:
		return fmt.Sprintf("%#v", res), nil
	case err := <-errChan:
		return "", err
	case <-time.After(time.Duration(timeout) * time.Second):
		return "", fmt.Errorf("timeout")
	}
}

// wrapAndSend will wrap a payload into a batchlist and sends it to the Sawtooth network
func (c *CookiejarClient) wrapAndSend(action string, amount int, timeout uint) (string, error) {
	rand.Seed(time.Now().UnixNano())

	// We're using CSV encoding
	payload := strings.Join([]string{action, strconv.Itoa(amount)}, ",")

	// Get the public key as a hex string
	pubKey := c.signer.GetPublicKey().AsHex()

	// Add the address to the address list
	addressList := []string{c.getAddress()}

	rawTransactionsHeader := transaction_pb2.TransactionHeader{
		SignerPublicKey:  pubKey,
		FamilyName:       familyName,
		FamilyVersion:    "1.0",
		Inputs:           addressList, // Important for parallel processing
		Outputs:          addressList, // Important for parallel processing
		PayloadSha512:    hexdigest(payload),
		BatcherPublicKey: pubKey,
		Nonce:            strconv.Itoa(rand.Int()),
	}

	// Serialize the raw transaction
	transactionHeader, err := proto.Marshal(&rawTransactionsHeader)
	if err != nil {
		return "", fmt.Errorf("Unable to serialize transaction header: %v", err)
	}

	// Create the signature for the transaction header
	transactionHeaderSignature := hex.EncodeToString(c.signer.Sign(transactionHeader))

	// Create a list of transactions
	transactions := []*transaction_pb2.Transaction{
		&transaction_pb2.Transaction{
			Header:          transactionHeader,
			HeaderSignature: transactionHeaderSignature,
			Payload:         []byte(payload),
		},
	}

	// Create the batch header
	rawBatchHeader := batch_pb2.BatchHeader{
		SignerPublicKey: pubKey,
		TransactionIds:  []string{transactions[0].HeaderSignature},
	}

	// Encode the batch header
	batchHeader, err := proto.Marshal(&rawBatchHeader)
	if err != nil {
		return "", fmt.Errorf("Unable to serialize batch header: %v", err)
	}

	// Create the batch header signature
	batchHeaderSignature := hex.EncodeToString(c.signer.Sign(batchHeader))

	// Create a batchlist
	rawBatchList := batch_pb2.BatchList{
		Batches: []*batch_pb2.Batch{
			&batch_pb2.Batch{
				Header:          batchHeader,
				Transactions:    transactions,
				HeaderSignature: batchHeaderSignature,
			},
		},
	}

	// Encode the batch list
	batchList, err := proto.Marshal(&rawBatchList)
	if err != nil {
		return "", fmt.Errorf("Unable to serialize batch list: %v", err)
	}

	// Send the request
	if _, err := c.sendRequest("batches", "application/octet-stream", batchList); err != nil {
		return "", err
	}

	// Wait for the status to change
	return c.waitForStatus(batchHeaderSignature, timeout)
}

// NewCookiejarClient returns an initialized cookiejar client
func NewCookiejarClient(url string, keyFile string) (*CookiejarClient, error) {
	// Get the locally stored private key
	var privateKey signing.PrivateKey
	if keyFile != "" {
		// Read private key file
		privateKeyStr, err := ioutil.ReadFile(keyFile)
		if err != nil {
			return nil, fmt.Errorf("Failed to read private key: %v", err)
		}
		// Get private key object
		privateKey = signing.NewSecp256k1PrivateKey(privateKeyStr)
	} else {
		privateKey = signing.NewSecp256k1Context().NewRandomPrivateKey()
	}

	// Initialize a new cryptoFactory
	cryptoFactory := signing.NewCryptoFactory(signing.NewSecp256k1Context())

	// Create a signer object via the cryptoFactory
	signer := cryptoFactory.NewSigner(privateKey)

	return &CookiejarClient{url, signer}, nil
}

package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/pem"
	"crypto/x509"
	"bytes"
	"errors"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}


func ParseCreatorCertificate(stub shim.ChaincodeStubInterface) (string, error) {
	creator, err := stub.GetCreator()
	if err != nil {
		return "", err
	}
	if creator == nil {
		fmt.Print("No certificate found")
		return "", errors.New("No certificate found")
	}
	certStart := bytes.IndexAny(creator, "----BEGIN CERTIFICATE-----")
	if certStart == -1 {
		fmt.Print("No certificate found")
		return "", errors.New("No certificate found")
	}
	certText := creator[certStart:]
	block, _ := pem.Decode(certText)
	if block == nil {
		fmt.Printf("Error received on pem.Decode of certificate %s",  certText)
		return "", errors.New("Error received on pem.Decode of certificate")
	}

	uCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		fmt.Printf("Error received on ParseCertificate %s", err)
		return "", err
	}

	fmt.Printf("Common Name %s ", uCert.Subject.CommonName)
	return uCert.Subject.CommonName, nil
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Init")
	err := stub.PutState("total_amount", []byte(strconv.Itoa(0)))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "create" {
		// Create wallet
		return t.create(stub, args)
	} else if function == "charge" {
		// Make payment of X units from A to B
		return t.charge(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return t.query(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

func (t *SimpleChaincode) create(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var accountName, cert string
	accountName = args[0]

	accountValBytes, err := stub.GetState(accountName)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if accountValBytes != nil {
		return shim.Error("Entity already exists")
	}

	cert, err = ParseCreatorCertificate(stub)

	err = stub.PutState(accountName, []byte(strconv.Itoa(100)))
	if err != nil {
		return shim.Error(err.Error())
	}

	totalBytes, err := stub.GetState("total_amount")
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if totalBytes == nil {
		return shim.Error("Entity not found")
	}
	Total, _ := strconv.Atoi(string(totalBytes))


	err = stub.PutState("total_amount", []byte(strconv.Itoa(Total + 100)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(accountName + "_owner", []byte(cert))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) charge(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var sourceAccount, destinationAccount, cert string    // Entities
	var sourceValue, destinationValue int // Asset holdings
	var chargeAmount int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	sourceAccount = args[0]
	destinationAccount = args[1]

	cert, err = ParseCreatorCertificate(stub)

	accountCertBytes, err := stub.GetState(sourceAccount + "_owner")

	if string(accountCertBytes) != cert {
		return shim.Error("Invalid certificate")
	}

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	sourceValBytes, err := stub.GetState(sourceAccount)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if sourceValBytes == nil {
		return shim.Error("Entity not found")
	}
	sourceValue, _ = strconv.Atoi(string(sourceValBytes))

	destinationValBytes, err := stub.GetState(destinationAccount)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if destinationValBytes == nil {
		return shim.Error("Entity not found")
	}
	destinationValue, _ = strconv.Atoi(string(destinationValBytes))

	// Perform the execution
	chargeAmount, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	sourceValue = sourceValue - chargeAmount
	destinationValue = destinationValue + chargeAmount
	fmt.Printf("Aval = %d, Bval = %d\n", sourceValue, destinationValue)

	// Write the state back to the ledger
	err = stub.PutState(sourceAccount, []byte(strconv.Itoa(sourceValue)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(destinationAccount, []byte(strconv.Itoa(destinationValue)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var accountName string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	accountName = args[0]

	// Get the state from the ledger
	accountValBytes, err := stub.GetState(accountName)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + accountName + "\"}"
		return shim.Error(jsonResp)
	}

	if accountValBytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + accountName + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + accountName + "\",\"Amount\":\"" + string(accountValBytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(accountValBytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

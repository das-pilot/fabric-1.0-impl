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
	"encoding/json"
	"time"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type HistoryResponse struct {
	Wallet  string
	Amount  float64
	Message string
	Time    string
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
	fmt.Println("Wallet Init")
	err := stub.PutState("total_amount", []byte("0"))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Wallet Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "create" {
		// Create wallet
		return t.create(stub, args)
	} else if function == "charge" {
		// Make payment of X units from A to B
		return t.charge(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return t.getBalance(stub, args)
	} else if function == "queryHistory" {
		return t.queryHistory(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

func (t *SimpleChaincode) create(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var accountName, cert string
	accountName = args[0]

	accountValBytes, err := stub.GetState(accountName + "_owner")
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if accountValBytes != nil {
		return shim.Error("Entity already exists")
	}

	cert, err = ParseCreatorCertificate(stub)

	err = stub.PutState(accountName + "_owner", []byte(cert))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// Transaction charges of X units from source to destination
func (t *SimpleChaincode) charge(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var sourceAccount, destinationAccount, cert string    // Entities
	var sourceValue, destinationValue float64 // Asset holdings
	var chargeAmount float64          // Charge value
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
	sourceValBytes, err := stub.GetState(sourceAccount + "." + destinationAccount)
	if err != nil {
		return shim.Error("Failed to get state")
	}

	if sourceValBytes == nil {
		sourceValue = 0
	} else {
		sourceValue, _ = strconv.ParseFloat(string(sourceValBytes), 64)
	}

	destinationValBytes, err := stub.GetState(destinationAccount + "." + sourceAccount)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if destinationValBytes == nil {
		destinationValue = 0
	} else {
		destinationValue, _ = strconv.ParseFloat(string(destinationValBytes), 64)
	}

	// Perform the execution
	chargeAmount, err = strconv.ParseFloat(args[2], 64)
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	sourceValue = sourceValue - chargeAmount
	destinationValue = destinationValue + chargeAmount
	fmt.Printf("Aval = %d, Bval = %d\n", sourceValue, destinationValue)

	// Write the state back to the ledger
	err = stub.PutState(sourceAccount  + "." + destinationAccount, []byte(strconv.FormatFloat(sourceValue, 'f', 2, 64)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(destinationAccount + "." + sourceAccount, []byte(strconv.FormatFloat(destinationValue, 'f', 2, 64)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *SimpleChaincode) getBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var sourceAccountName, destinationAccountName string // Entities
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	sourceAccountName = args[0]
	destinationAccountName = args[1]

	// Get the state from the ledger
	accountValBytes, err := stub.GetState(sourceAccountName + "." + destinationAccountName)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + sourceAccountName + "." + destinationAccountName + "\"}"
		return shim.Error(jsonResp)
	}

	if accountValBytes == nil {
		accountValBytes = []byte("0")
	}

	jsonResp := "{\"Source\":\"" + sourceAccountName + "\"," +
		"\"Destination\":\"" + destinationAccountName + "\",\"" +
		"Amount\":\"" + string(accountValBytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(accountValBytes)
}

func (t *SimpleChaincode) queryHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var sourceAccount, destinationAccount string
	sourceAccount = args[0]
	destinationAccount = args[1]
	iterator, err := stub.GetHistoryForKey(sourceAccount + "." + destinationAccount)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + sourceAccount + "." + destinationAccount + "\"}"
		fmt.Print(jsonResp)
		return shim.Error(jsonResp)
	}
	var prevVal, amount float64
	var flag bool
	flag = false
	prevVal = 0
	response := []HistoryResponse{}
	for ;iterator.HasNext(); {
		qResult, err := iterator.Next()
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to get state for " + sourceAccount + "." + destinationAccount + "\"}"
			fmt.Print(jsonResp)
			return shim.Error(jsonResp)
		}

		accountValBytes := qResult.Value
		value, err := strconv.ParseFloat(string(accountValBytes), 64)
		if err != nil {
			return shim.Error(err.Error())
		}
		if !flag {
			amount = value
			prevVal = amount
			flag = true
		} else {
			amount = value - prevVal
			prevVal = value
		}
		message := "INCOME"
		if amount < 0 {
			message = "SPENT"
		}
		response = append(response, HistoryResponse{
			Wallet:  destinationAccount,
			Amount:  amount,
			Message: message,
			Time:    time.Unix(qResult.Timestamp.GetSeconds(),
				int64(qResult.Timestamp.GetNanos())).Format("15.01.2006 15:04:05")})
	}
	jsonStr, err := json.Marshal(response)
	fmt.Print(jsonStr)
	return shim.Success(jsonStr)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

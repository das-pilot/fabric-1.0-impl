/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

//WARNING - this chaincode's ID is hard-coded in chaincode_example04 to illustrate one way of
//calling chaincode from a chaincode. If this example is modified, chaincode_example04.go has
//to be modified as well with the new ID of chaincode_example02.
//chaincode_example05 show's how chaincode ID can be passed in as a parameter instead of
//hard-coding.

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

	ucert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		fmt.Printf("Error received on ParseCertificate %s", err)
		return "", err
	}

	fmt.Printf("Common Name %s ", ucert.Subject.CommonName)
	return ucert.Subject.CommonName, nil
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
	var A, Cert string
	A = args[0]

	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes != nil {
		return shim.Error("Entity already exists")
	}

	Cert, err = ParseCreatorCertificate(stub)

	err = stub.PutState(A, []byte(strconv.Itoa(100)))
	if err != nil {
		return shim.Error(err.Error())
	}

	TotalBytes, err := stub.GetState("total_amount")
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if TotalBytes == nil {
		return shim.Error("Entity not found")
	}
	Total, _ := strconv.Atoi(string(TotalBytes))


	err = stub.PutState("total_amount", []byte(strconv.Itoa(Total + 100)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(A + "_owner", []byte(Cert))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) charge(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B, Cert string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	Cert, err = ParseCreatorCertificate(stub)

	Avalcert, err := stub.GetState(A + "_owner")

	if string(Avalcert) != Cert {
		return shim.Error("Invalid certificate")
	}

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

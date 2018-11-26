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

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("fabric-ocean")

type OceanChaincode struct {
}

type Token struct {
	Address     string `json:"address"`
	TokenName   string `json:"tokenName"`
	TotalNumber string `json:"totalNumber"`
}

const (
	TokenPrefix      = "TokenPrefix"
	WalletInfoPrefix = "WalletInfoPrefix"
)

func (t *OceanChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	var A, B string
	var Aval, Bval int
	var err error

	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		logger.Error(err)
		return shim.Error("Expecting integer value for asset holding")
	}

	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		logger.Error(err)
		return shim.Error("Expecting integer value for asset holding")
	}

	logger.Infof("Aval = %d, Bval = %d\n", Aval, Bval)

	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *OceanChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	logger.Info("function =", function)

	if function == "delete" {
		return t.delete(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	} else if function == "initValue" {
		return t.initValue(stub, args)
	} else if function == "move" {
		return t.move(stub, args)
	}

	logger.Error("func unknown : " + function)
	return shim.Error(function + " unknown")
}

func (t *OceanChaincode) issueToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 4 {
		return shim.Error("incorrect number of args")
	}

	tokenID := args[0]

	verify, err := Verify(args[1], args[2], args[3])
	if !verify {
		return shim.Error("verify fail: " + err.Error())
	}

	if tokenID == "" {
		return shim.Error("tokenID is null")
	}

	tokenJson, err := hex.DecodeString(args[2])
	if err != nil {
		return shim.Error(err.Error())
	}

	token := Token{}
	err = json.Unmarshal(tokenJson, &token)
	if err != nil {
		return shim.Error("json unmarshal fail")
	}

	if GetAddress(args[1]) != token.Address {
		return shim.Error("address and public key not match")
	}

	if len(token.TokenName) < 2 || len(token.TokenName) > 16 {
		return shim.Error("tokenName need have 2-16 char")
	}

	if !IsGtZeroInteger(token.TotalNumber) {
		return shim.Error("totalNumber need to be greater than 0 integer")
	}

	totalNumber := new(big.Int)
	totalNumber, success := totalNumber.SetString(token.TotalNumber, 10)
	if !success {
		return shim.Error("totalNumber not match: " + token.TotalNumber)
	}

	tokenIDBytes, err := stub.GetState(TokenPrefix + tokenID)
	if len(tokenIDBytes) != 0 {
		return shim.Error("token already existed")
	}

	err = stub.PutState(TokenPrefix+tokenID, tokenJson)
	if err != nil {
		return shim.Error(err.Error())
	}

	compositeKey, err := stub.CreateCompositeKey(WalletInfoPrefix+token.Address, []string{tokenID, "+", token.TotalNumber})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(compositeKey, []byte{0})
	if err != nil {
		return shim.Error(err.Error())
	}
	//GetStateByPartialCompositeKey
	//SplitCompositeKey

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(OceanChaincode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}

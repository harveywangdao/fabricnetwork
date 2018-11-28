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
	"errors"
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
	TokenPrefix    = "TokenPrefix"
	WalletPrefix   = "WalletPrefix"
	TransferPrefix = "TransferPrefix"
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
	} else if function == "queryToken" {
		return t.queryToken(stub, args)
	} else if function == "queryBalance" {
		return t.queryBalance(stub, args)
	} else if function == "issueToken" {
		return t.issueToken(stub, args)
	} else if function == "transfer" {
		return t.transfer(stub, args)
	} else if function == "queryTx" {
		return t.queryTx(stub, args)
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

	compositeKey, err := stub.CreateCompositeKey(WalletPrefix+token.Address, []string{tokenID, "+", token.TotalNumber, "issueToken"})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(compositeKey, []byte{0})
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *OceanChaincode) queryToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("incorrect number of args")
	}

	tokenID := args[0]

	tokenIDBytes, err := stub.GetState(TokenPrefix + tokenID)
	if len(tokenIDBytes) == 0 || err != nil {
		return shim.Error("token not exist")
	}

	return shim.Success(tokenIDBytes)
}

type TokenBalance struct {
	TokenID        string   `json:"tokenID"`
	Balance        string   `json:"balance"`
	BalanceNumeric *big.Int `json:"-"`
}

type BalanceInfo struct {
	Address       string          `json:"address"`
	TokenBalances []*TokenBalance `json:"tokenBalances"`
}

func (t *OceanChaincode) queryBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("incorrect number of args")
	}

	address := args[0]

	iterator, err := stub.GetStateByPartialCompositeKey(WalletPrefix+address, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer iterator.Close()

	balanceInfo := BalanceInfo{
		Address: address,
	}

	for iterator.HasNext() {
		responseRange, err := iterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		tokenID := compositeKeyParts[0]
		operation := compositeKeyParts[1]
		num := compositeKeyParts[2]

		numBigInt := new(big.Int)
		numBigInt, success := numBigInt.SetString(num, 10)
		if !success {
			return shim.Error("number not match: " + num)
		}

		exist := false
		for _, tokenBalance := range balanceInfo.TokenBalances {
			if tokenBalance.TokenID == tokenID {
				if operation == "+" {
					tokenBalance.BalanceNumeric = new(big.Int).Add(tokenBalance.BalanceNumeric, numBigInt)
				} else {
					tokenBalance.BalanceNumeric = new(big.Int).Sub(tokenBalance.BalanceNumeric, numBigInt)
				}
				exist = true
				break
			}
		}

		if !exist {
			tokenBalance := &TokenBalance{
				TokenID:        tokenID,
				BalanceNumeric: big.NewInt(0),
			}

			if operation == "+" {
				tokenBalance.BalanceNumeric = new(big.Int).Add(tokenBalance.BalanceNumeric, numBigInt)
			} else {
				tokenBalance.BalanceNumeric = new(big.Int).Sub(tokenBalance.BalanceNumeric, numBigInt)
			}

			balanceInfo.TokenBalances = append(balanceInfo.TokenBalances, tokenBalance)
		}
	}

	for i := 0; i < len(balanceInfo.TokenBalances); i++ {
		balanceInfo.TokenBalances[i].Balance = balanceInfo.TokenBalances[i].BalanceNumeric.String()
	}

	balanceData, err := json.Marshal(&balanceInfo)
	if err != nil {
		return shim.Error("Json marshal fail: " + err.Error())
	}

	return shim.Success(balanceData)
}

func (t *OceanChaincode) getBalance(stub shim.ChaincodeStubInterface, address string) (*BalanceInfo, error) {
	iterator, err := stub.GetStateByPartialCompositeKey(WalletPrefix+address, []string{})
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	balanceInfo := BalanceInfo{
		Address: address,
	}

	for iterator.HasNext() {
		responseRange, err := iterator.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		tokenID := compositeKeyParts[0]
		operation := compositeKeyParts[1]
		num := compositeKeyParts[2]

		numBigInt := new(big.Int)
		numBigInt, success := numBigInt.SetString(num, 10)
		if !success {
			return nil, errors.New("number not match: " + num)
		}

		exist := false
		for _, tokenBalance := range balanceInfo.TokenBalances {
			if tokenBalance.TokenID == tokenID {
				if operation == "+" {
					tokenBalance.BalanceNumeric = new(big.Int).Add(tokenBalance.BalanceNumeric, numBigInt)
				} else {
					tokenBalance.BalanceNumeric = new(big.Int).Sub(tokenBalance.BalanceNumeric, numBigInt)
				}
				exist = true
				break
			}
		}

		if !exist {
			tokenBalance := &TokenBalance{
				TokenID:        tokenID,
				BalanceNumeric: big.NewInt(0),
			}

			if operation == "+" {
				tokenBalance.BalanceNumeric = new(big.Int).Add(tokenBalance.BalanceNumeric, numBigInt)
			} else {
				tokenBalance.BalanceNumeric = new(big.Int).Sub(tokenBalance.BalanceNumeric, numBigInt)
			}

			balanceInfo.TokenBalances = append(balanceInfo.TokenBalances, tokenBalance)
		}
	}

	for i := 0; i < len(balanceInfo.TokenBalances); i++ {
		balanceInfo.TokenBalances[i].Balance = balanceInfo.TokenBalances[i].BalanceNumeric.String()
	}

	return &balanceInfo, nil
}

type Transfer struct {
	FromAddress string `json:"fromAddress"`
	ToAddress   string `json:"toAddress"`
	TokenID     string `json:"tokenID"`
	Number      string `json:"number"`
	TxID        string `json:"txID"`
}

func (t *OceanChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 4 {
		return shim.Error("incorrect number of args")
	}

	txID := args[0]
	if txID == "" {
		return shim.Error("txID is null")
	}

	verify, err := Verify(args[1], args[2], args[3])
	if !verify {
		return shim.Error("verify fail: " + err.Error())
	}

	transferJson, err := hex.DecodeString(args[2])
	if err != nil {
		return shim.Error(err.Error())
	}

	tx := Transfer{}
	err = json.Unmarshal(transferJson, &tx)
	if err != nil {
		return shim.Error("json unmarshal fail")
	}

	if GetAddress(args[1]) != tx.FromAddress {
		return shim.Error("address and public key not match")
	}

	if !IsValidAddress(tx.ToAddress) {
		return shim.Error("toAddress is invalid")
	}

	if tx.FromAddress == tx.ToAddress {
		return shim.Error("fromAddress and toAddress can not be same")
	}

	if tx.Number == "" || tx.TokenID == "" {
		return shim.Error("number or tokenID is null string")
	}

	if !IsGtZeroInteger(tx.Number) {
		return shim.Error("number need to be greater than 0 integer")
	}

	tokenIDBytes, err := stub.GetState(TokenPrefix + tx.TokenID)
	if len(tokenIDBytes) == 0 {
		return shim.Error("token not exist")
	}

	transferBytes, err := stub.GetState(TransferPrefix + txID)
	if len(transferBytes) != 0 {
		return shim.Error("transfer already existed")
	}

	fromAddrBalanceInfo, err := t.getBalance(stub, tx.FromAddress)
	if err != nil {
		return shim.Error(err.Error())
	}

	number := new(big.Int)
	number, success := number.SetString(tx.Number, 10)
	if !success {
		return shim.Error("number not match: " + tx.Number)
	}

	fromAddrBalance := big.NewInt(0)
	for _, tokenBalace := range fromAddrBalanceInfo.TokenBalances {
		if tokenBalace.TokenID == tx.TokenID {
			fromAddrBalance = tokenBalace.BalanceNumeric
			break
		}
	}

	if fromAddrBalance.Cmp(number) < 0 {
		return shim.Error("balance of fromAddress " + fromAddrBalance.String() + " less than " + "number " + tx.Number)
	}

	tx.TxID = txID
	txJson, err := json.Marshal(&tx)
	if err != nil {
		return shim.Error("Json marshal fail: " + err.Error())
	}

	err = stub.PutState(TransferPrefix+txID, txJson)
	if err != nil {
		return shim.Error(err.Error())
	}

	compositeKey, err := stub.CreateCompositeKey(WalletPrefix+tx.FromAddress, []string{tx.TokenID, "-", tx.Number, txID})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(compositeKey, []byte{0})
	if err != nil {
		return shim.Error(err.Error())
	}

	compositeKey, err = stub.CreateCompositeKey(WalletPrefix+tx.ToAddress, []string{tx.TokenID, "+", tx.Number, txID})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(compositeKey, []byte{0})
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *OceanChaincode) queryTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("incorrect number of args")
	}

	txID := args[0]

	txBytes, err := stub.GetState(TransferPrefix + txID)
	if len(txBytes) == 0 || err != nil {
		return shim.Error("transfer not exist")
	}

	return shim.Success(txBytes)
}

func main() {
	err := shim.Start(new(OceanChaincode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}

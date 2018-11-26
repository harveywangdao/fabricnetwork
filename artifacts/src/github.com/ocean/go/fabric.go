package main

import (
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func (t *OceanChaincode) initValue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("initValue enter")

	if len(args) != 2 {
		logger.Error("args error, args :", args)
		return shim.Error("Incorrect number of arguments. Expecting 4, function followed by 2 names and 1 value")
	}

	toAddr := args[0]
	logger.Info(toAddr)

	// Perform the execution
	toAddrVal, err := strconv.Atoi(args[1])
	if err != nil {
		logger.Error(err)
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	logger.Info(toAddrVal)

	err = stub.PutState(toAddr, []byte(strconv.Itoa(toAddrVal)))
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}
	logger.Info("initValue end")

	return shim.Success(nil)
}

func (t *OceanChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		logger.Error("args error, args :", args)
		return shim.Error("Incorrect number of arguments. Expecting 4, function followed by 2 names and 1 value")
	}

	logger.Info("move args :", args)

	fromAddr := args[0]
	toAddr := args[1]

	fromAddrValueBytes, err := stub.GetState(fromAddr)
	if err != nil {
		logger.Error(err)
		return shim.Error("Failed to get state")
	}

	if fromAddrValueBytes == nil {
		logger.Error("fromAddrValueBytes is nil")
		return shim.Error("Entity not found")
	}
	fromAddrVal, _ := strconv.Atoi(string(fromAddrValueBytes))

	toAddrValueBytes, err := stub.GetState(toAddr)
	if err != nil {
		logger.Error(err)
		return shim.Error("Failed to get state")
	}

	var toAddrVal int
	if toAddrValueBytes == nil {
		logger.Error("toAddrValueBytes is nil")
		toAddrVal = 0
		//return shim.Error("Entity not found")
	} else {
		toAddrVal, _ = strconv.Atoi(string(toAddrValueBytes))
	}

	num, err := strconv.Atoi(args[2])
	if err != nil {
		logger.Error(err)
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}

	fromAddrVal = fromAddrVal - num
	toAddrVal = toAddrVal + num

	err = stub.PutState(fromAddr, []byte(strconv.Itoa(fromAddrVal)))
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	err = stub.PutState(toAddr, []byte(strconv.Itoa(toAddrVal)))
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *OceanChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		logger.Error("args error, args :", args)
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	logger.Info("delete :", A)

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		logger.Error(err)
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// Query callback representing the query of a chaincode
func (t *OceanChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		logger.Error("args error, args :", args)
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]
	logger.Info("query :", A)

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		logger.Error(err)
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}
	logger.Info("Avalbytes =", string(Avalbytes))

	if Avalbytes == nil {
		logger.Error("Avalbytes is nil")
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	logger.Info("Query Response:", string(Avalbytes))
	return shim.Success(Avalbytes)
}

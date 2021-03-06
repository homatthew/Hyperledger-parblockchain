/**
* Copyright 2017 HUAWEI. All Rights Reserved.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const ERROR_SYSTEM = "{\"code\":300, \"reason\": \"system error: %s\"}"
const ERROR_WRONG_FORMAT = "{\"code\":301, \"reason\": \"command format is wrong\"}"
const ERROR_ACCOUNT_EXISTING = "{\"code\":302, \"reason\": \"account already exists\"}"
const ERROR_ACCOUT_ABNORMAL = "{\"code\":303, \"reason\": \"abnormal account\"}"
const ERROR_MONEY_NOT_ENOUGH = "{\"code\":304, \"reason\": \"account's money is not enough\"}"

type NewChaincode struct {
}

// Initialize the accounts
func (t *NewChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// nothing to do
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments")
	}

	nacc, err := strconv.Atoi(args[0])
	money, err2 := strconv.Atoi(args[1])

	if err != nil || err2 != nil {
		return shim.Error("Expecting integer value for number of accounts")
	}

	for i := 0; i < nacc; {
		acc := fmt.Sprintf("acc%d", i)
		err = stub.PutState(acc, []byte(strconv.Itoa(money)))
		if err != nil {
			fmt.Println("error putting state in Init")
		} else {
			i += 1
		}
	}

	fmt.Println("Initialized accounts")

	return shim.Success(nil)
}

func (t *NewChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "open" {
		return t.Open(stub, args)
	}
	if function == "delete" {
		return t.Delete(stub, args)
	}
	if function == "query" {
		return t.Query(stub, args)
	}
	if function == "transfer" {
		return t.Transfer(stub, args)
	}
	if function == "readwrite" {
		return t.ReadWrite(stub, args)
	}

	return shim.Error("Error in Invoke function")
}

// open an account, should be [open account money]
func (t *NewChaincode) Open(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error(ERROR_WRONG_FORMAT)
	}

	account := args[0]
	money, err := stub.GetState(account)
	if money != nil {
		return shim.Error(ERROR_ACCOUNT_EXISTING)
	}

	_, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error(ERROR_WRONG_FORMAT)
	}

	err = stub.PutState(account, []byte(args[1]))
	if err != nil {
		s := fmt.Sprintf(ERROR_SYSTEM, err.Error())
		return shim.Error(s)
	}

	return shim.Success(nil)
}

// delete an account, should be [delete account]
func (t *NewChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error(ERROR_WRONG_FORMAT)
	}

	err := stub.DelState(args[0])
	if err != nil {
		s := fmt.Sprintf(ERROR_SYSTEM, err.Error())
		return shim.Error(s)
	}

	return shim.Success(nil)
}

// query current money of the account,should be [query accout]
func (t *NewChaincode) Query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error(ERROR_WRONG_FORMAT)
	}

	money, err := stub.GetState(args[0])
	if err != nil {
		s := fmt.Sprintf(ERROR_SYSTEM, err.Error())
		return shim.Error(s)
	}

	if money == nil {
		return shim.Error(ERROR_ACCOUT_ABNORMAL)
	}

	return shim.Success(money)
}

// transfer money from account1 to account2, should be [transfer accout1 accout2 money]
func (t *NewChaincode) Transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Number of args error")
	}
	money, err := strconv.Atoi(args[2])
	fmt.Println(args[1])
	if err != nil {
		return shim.Error("Money format error")
	}

	moneyBytes1, err1 := stub.GetState(args[0])
	moneyBytes2, err2 := stub.GetState(args[1])
	if err1 != nil || err2 != nil {
		s := fmt.Sprintf(ERROR_SYSTEM, err.Error())
		return shim.Error(s)
	}
	if moneyBytes1 == nil || moneyBytes2 == nil {
		return shim.Error(ERROR_ACCOUT_ABNORMAL)
	}

	money1, _ := strconv.Atoi(string(moneyBytes1))
	money2, _ := strconv.Atoi(string(moneyBytes1))
	if money1 < money {
		return shim.Error(ERROR_MONEY_NOT_ENOUGH)
	}

	money1 -= money
	money2 += money

	err = stub.PutState(args[0], []byte(strconv.Itoa(money1)))
	if err != nil {
		s := fmt.Sprintf(ERROR_SYSTEM, err.Error())
		return shim.Error(s)
	}

	err = stub.PutState(args[1], []byte(strconv.Itoa(money2)))
	if err != nil {
		stub.PutState(args[0], []byte(strconv.Itoa(money1+money)))
		s := fmt.Sprintf(ERROR_SYSTEM, err.Error())
		return shim.Error(s)
	}

	return shim.Success(nil)
}

// [ReadWritex 5 acc1 acc2 acc3 acc4 acc5 2 acc5 acc6]
func (t *NewChaincode) ReadWrite(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	ind := 0
	readCount, err := strconv.Atoi(args[ind])
	if err != nil {
		return shim.Error("Format error in reads")
	}
	ind++
	for i := 0; i < readCount; i++ {
		moneyByte, err := stub.GetState(args[ind])
		if err != nil {
			s := fmt.Sprintf(ERROR_SYSTEM, err.Error())
			return shim.Error(s)
		}
		if moneyByte == nil {
			return shim.Error("ERROR_ACCOUNT_ABNORMAL")
		}
		ind++
	}
	writeCount, err := strconv.Atoi(args[ind])
	ind++
	if err != nil {
		return shim.Error("Number of writes error")
	}
	for i := 0; i < writeCount; i++ {
		err := stub.PutState(args[ind], []byte("100"))
		if err != nil {
			s := fmt.Sprintf(ERROR_SYSTEM, err.Error())
			return shim.Error(s)
		}
		ind++
	}
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(NewChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %v \n", err)
	}

}

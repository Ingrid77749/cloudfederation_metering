/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	"reflect"
	"strings"

	//"google.golang.org/genproto/googleapis/appengine/v1"
	"strconv"
)

//通过shim包中不同链码API操作， 实现不同的业务逻辑（对交易账本进行查询或更新）

type MyChaincode struct {
}

const (
	InitialQuotaSuffix = "_InitialQuota"
	ECSResourceSuffix  = "_ECSResource"
	ECSBillSuffix      = "_ECSBill"
	OSSResourceSuffix  = "_OSSResource"
	OSSBillSuffix      = "_OSSBill"
	PriceSuffix        = "_Price"
)

type ObjectType string

const (
	InitialQuotaObjectType ObjectType = "InitialQuota"
	ECSResourceObjectType             = "ECSResource"
	ECSBillObjectType                 = "ECSBill"
	OSSResourceObjectType             = "OSSResource"
	OSSBillObjectType                 = "OSSBill"
	PriceObjectType                   = "Price"
)

func validObjectType() []string {
	return strings.Split("InitialQuota,ECSResource,ECSBill,OSSResource,OSSBill,Price", ",")
}

func (o ObjectType) String() string {
	switch o {
	case InitialQuotaObjectType:
		return "InitialQuota"
	case ECSResourceObjectType:
		return "ECSResource"
	case ECSBillObjectType:
		return "ECSBill"
	case OSSResourceObjectType:
		return "OSSResource"
	case OSSBillObjectType:
		return "OSSBill"
	case PriceObjectType:
		return "Price"
	default:
		return ""
	}
}

//初始配额情况
type InitialQuota struct {
	ObjectType       ObjectType `json:"objectType"`
	QuotaID          string     `json:"quotaID"`
	ResourceProvider string     `json:"resourceProvider"`
	ProviderID       string     `json:"providerID"`
	ResourceUser     string     `json:"resourceUser"`
	ResourceUserID   string     `json:"resource_userID"` //每个组织都有自己的ID号 不等同于下面业务中的用户个人ID号
	HoldingTime      uint64     `json:"holdingtime"`
	CPU              uint64     `json:"CPU"`
	RAM              uint64     `json:"RAM"`
	HardDisk_ecs     uint64     `json:"hardDisk_ecs"`
	HardDisk_oss     uint64     `json:"hardDisk_oss"`
	NetworkBandwidth uint64     `json:"networkBandwidth"`
	PublicIP         uint64     `json:"publicIP"`
	PrivateIP        uint64     `json:"privateIP"`
}

//ECS资源申请情况
type ECSResource struct {
	ObjectType            ObjectType          `json:"objectType"` //区分状态数据库中的各类对象
	InstanceID            string              `json:"instanceID"`
	ResourceProvider      string              `json:"resourceProvider"`
	ProviderID            string              `json:"providerID"`
	ResourceUser          string              `json:"resourceUser"`
	UserID                string              `json:"userID"`
	CloudService          string              `json:"cloudService"`
	BeginTime             timestamp.Timestamp `json:"beginTime"`
	NetworkType           string              `json:"networkType"`
	NetworkBandwidth      uint64              `json:"networkBandwidth"`
	CPU                   uint64              `json:"CPU"`
	RAM                   uint64              `json:"RAM"`
	HardDisk              uint64              `json:"hardDisk"`
	CommitmentServicetime uint64              `json:"commitment_servicetime"`
}

//ECS资源使用情况
type ECSBill struct {
	//Record *Resource
	ObjectType  ObjectType          `json:"objectType"` //区分状态数据库中的各类对象
	ECSResource ECSResource         `json:"ECSResource"`
	ProviderID  string              `json:"providerID"`
	UserID      string              `json:"userID"`
	InstanceID  string              `json:"instanceID"`
	EndTime     timestamp.Timestamp `json:"endTime"`
	UsageTime   uint64              `json:"usageTime"`
	TotalCost   uint64              `json:"totalCost"`
}

//OSS资源申请情况
type OSSResource struct {
	ObjectType       ObjectType          `json:"objectType"` //区分状态数据库中的各类对象
	InstanceID       string              `json:"instanceID"`
	ResourceProvider string              `json:"resourceProvider"`
	ProviderID       string              `json:"providerID"`
	ResourceUser     string              `json:"resourceUser"`
	UserID           string              `json:"userID"`
	CloudService     string              `json:"cloudService"`
	BeginTime        timestamp.Timestamp `json:"beginTime"`
	HardDisk         uint64              `json:"hardDisk"`
	NetworkType      string              `json:"network_type"`
}

//OSS资源使用情况
type OSSBill struct {
	//Record *Resource
	ObjectType  ObjectType          `json:"objectType"` //区分状态数据库中的各类对象
	OSSresource OSSResource         `json:"OSSResource"`
	ProviderID  string              `json:"providerID"`
	UserID      string              `json:"userID"`
	InstanceID  string              `json:"instanceID"`
	OSSRequest  uint64              `json:"ossRequest"`
	EndTime     timestamp.Timestamp `json:"endTime"`
	UsageTime   uint64              `json:"usageTime"`
	FlowOut     uint64              `json:"flowout"`
	TotalCost   uint64              `json:"totalCost"`
}

//资源定价情况
type Price struct {
	ObjectType               ObjectType `json:"objectType"`
	ResourceProvider         string     `json:"resouce_provider"`
	ProviderID               string     `json:"provider_id"` //主键
	CpuPrice                 uint64     `json:"cpu_price"`   //CPU 单价 分/核/天
	RamPrice                 uint64     `json:"ram_price"`
	HarddiskEcs              uint64     `json:"hard_disk_ecs"`
	HarddiskOss              uint64     `json:"hard_disk_oss"`
	RequestPrice             uint64     `json:"request_price"`
	NetworkBandwidth_public  uint64     `json:"network_bandwidth_public"`
	NetworkBandwidth_private uint64     `json:"network_bandwidth_private"`
}

type QueryResult struct {
	Key string `json:"Key"`
	//Record *Bill
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
//对数据初始化
func (t *MyChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

//调用链码的入口
func (t *MyChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()
	fmt.Println("Invoke is running " + fn)

	//var result string
	//var err error
	switch fn {
	case "initECSResource":
		return t.initECSResource(stub, args)
	case "initOSSResource":
		return t.initOSSResource(stub, args)
	case "initInitialQuota":
		return t.initInitialQuota(stub, args)
	case "initPrice":
		return t.initPrice(stub, args)
	case "readResource":
		return t.readResource(stub, args)
	case "updateQuota":
		return t.updateQuota(stub, args)
	case "updatePrice":
		return t.updatePrice(stub, args)
	case "billedECS":
		return t.billedECS(stub, args)
	case "billedOSS":
		return t.billedOSS(stub, args)
	case "queryProviderBill":
		return t.queryProviderBill(stub, args)
	case "queryProviderTotalCost":
		return t.queryProviderTotalCost(stub, args)
	case "queryUserBill":
		return t.queryUserBill(stub, args)
	case "queryUserTotalCost":
		return t.queryUserTotalCost(stub, args)
	default:
		return shim.Error(fmt.Sprintf("unsupported function: %s", fn))
	}
	// Return the result as success payload
	//return shim.Success([]byte(result))
}

//init ecs-resource
func (t *MyChaincode) initECSResource(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error
	if s := checkArgumentsNum(12, args); len(s) != 0 {
		return shim.Error(s)
	}
	fmt.Println("- start init ECS resource")
	InstanceID := args[0]
	ResourceProvider := args[1]
	ProviderID := args[2]
	ResourceUser := args[3]
	UserID := args[4]
	CloudService := args[5]
	NetworkType := args[6]
	NetworkBandwidth, err := strconv.ParseUint(args[7], 10, 64)
	CPU, err := strconv.ParseUint(args[8], 10, 64)
	RAM, err := strconv.ParseUint(args[9], 10, 64)
	HardDisk, err := strconv.ParseUint(args[10], 10, 64)
	CommitmentServiceTime, err := strconv.ParseUint(args[11], 10, 64)
	// check if resource already exists
	r, err := getECSResource(InstanceID, stub)
	if err != nil {
		return shim.Error("Failed to get ECSResource: " + err.Error())
	} else if r != nil {
		fmt.Println("This ECSResource already exists: " + InstanceID)
		return shim.Error("This ECSResource already exists: " + InstanceID)
	}

	//create resource object and marshal to JSON
	now, err := stub.GetTxTimestamp()
	if err != nil {
		return shim.Error(err.Error())
	}
	resource := &ECSResource{
		ObjectType:            ECSResourceObjectType,
		InstanceID:            InstanceID,
		ResourceProvider:      ResourceProvider,
		ProviderID:            ProviderID,
		ResourceUser:          ResourceUser,
		UserID:                UserID,
		CloudService:          CloudService,
		BeginTime:             *now,
		NetworkType:           NetworkType,
		NetworkBandwidth:      NetworkBandwidth,
		CPU:                   CPU,
		RAM:                   RAM,
		HardDisk:              HardDisk,
		CommitmentServicetime: CommitmentServiceTime,
	}

	err = putState(resource, InstanceID, stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("ECSResource put state failed.\n%+v", resource))
	}
	fmt.Println("-end init ECSResource")
	return shim.Success(nil)
}

func (t *MyChaincode) initOSSResource(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error
	if s := checkArgumentsNum(8, args); len(s) != 0 {
		return shim.Error(s)
	}
	fmt.Println("- start init OSS resource")
	InstanceID := args[0]
	ResourceProvider := args[1]
	ProviderID := args[2]
	ResourceUser := args[3]
	UserID := args[4]
	CloudService := args[5]
	NetworkType := args[6]
	HardDisk, err := strconv.ParseUint(args[7], 10, 64)

	// check if resource already exists
	r, err := getOSSResource(InstanceID, stub)
	if err != nil {
		return shim.Error("Failed to get resource: " + err.Error())
	} else if r != nil {
		fmt.Println("This OSS resource already exists: " + InstanceID)
		return shim.Error("This OSS resource already exists: " + InstanceID)
	}
	//create resource object and marshal to JSON
	now, err := stub.GetTxTimestamp()
	if err != nil {
		return shim.Error(err.Error())
	}
	resource := &OSSResource{
		ObjectType:       OSSResourceObjectType,
		InstanceID:       InstanceID,
		ResourceProvider: ResourceProvider,
		ProviderID:       ProviderID,
		ResourceUser:     ResourceUser,
		UserID:           UserID,
		CloudService:     CloudService,
		BeginTime:        *now,
		NetworkType:      NetworkType,
		HardDisk:         HardDisk,
	}
	err = putState(resource, InstanceID, stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("OSSResource put state failed.\n%+v", resource))
	}
	fmt.Println("-end init OSS resource")
	return shim.Success(nil)
}
func (t *MyChaincode) initInitialQuota(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error
	if s := checkArgumentsNum(13, args); len(s) != 0 {
		return shim.Error(s)
	}
	fmt.Println("- start init initial quota")
	QuotaID := args[0]
	ResourceProvider := args[1]
	ProviderID := args[2]
	ResourceUser := args[3]
	ResourceUserID := args[4]
	HoldingTime, err := strconv.ParseUint(args[5], 10, 64)
	CPU, err := strconv.ParseUint(args[6], 10, 64)
	RAM, err := strconv.ParseUint(args[7], 10, 64)
	HardDisk_ecs, err := strconv.ParseUint(args[8], 10, 64)
	HardDisk_oss, err := strconv.ParseUint(args[9], 10, 64)
	NetworkBandwidth, err := strconv.ParseUint(args[10], 10, 64)
	PublicIP, err := strconv.ParseUint(args[11], 10, 64)
	PrivateIP, err := strconv.ParseUint(args[12], 10, 64)

	// check if resource already exists
	r, err := getInitialQuota(QuotaID, stub)
	if err != nil {
		return shim.Error("Failed to get InitialQuota: " + err.Error())
	} else if r != nil {
		fmt.Println("This InitialQuota already exists: " + QuotaID)
		return shim.Error("This InitialQuota already exists: " + QuotaID)
	}

	//create quota object and marshal to JSON
	quota := &InitialQuota{
		ObjectType:       InitialQuotaObjectType,
		QuotaID:          QuotaID,
		ResourceProvider: ResourceProvider,
		ProviderID:       ProviderID,
		ResourceUser:     ResourceUser,
		ResourceUserID:   ResourceUserID,
		HoldingTime:      HoldingTime,
		CPU:              CPU,
		RAM:              RAM,
		HardDisk_ecs:     HardDisk_ecs,
		HardDisk_oss:     HardDisk_oss,
		NetworkBandwidth: NetworkBandwidth,
		PublicIP:         PublicIP,
		PrivateIP:        PrivateIP,
	}
	err = putState(quota, QuotaID, stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("InitialQuota put state failed.\n%+v", quota))
	}
	fmt.Println("-end init initial quota")
	return shim.Success(nil)
}

func (t *MyChaincode) initPrice(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error
	if s := checkArgumentsNum(9, args); len(s) != 0 {
		return shim.Error(s)
	}
	fmt.Println("- start init price table")
	ResourceProvider := args[0]
	ProviderID := args[1]
	CPU_price, err := strconv.ParseUint(args[2], 10, 64)
	RAM_price, err := strconv.ParseUint(args[3], 10, 64)
	harddiskEcs, err := strconv.ParseUint(args[4], 10, 64)
	harddiskOss, err := strconv.ParseUint(args[5], 10, 64)
	requestPrice, err := strconv.ParseUint(args[6], 10, 64)
	networkbandwidthPublic, err := strconv.ParseUint(args[7], 10, 64)
	networkbandwidthPrivate, err := strconv.ParseUint(args[8], 10, 64)

	// check if resource already exists
	r, err := getPrice(ProviderID, stub)
	if err != nil {
		return shim.Error("Failed to get Price: " + err.Error())
	} else if r != nil {
		fmt.Println("This Price already exists: " + ProviderID)
		return shim.Error("This Price already exists: " + ProviderID)
	}
	//create quota object and marshal to JSON

	if err != nil {
		return shim.Error(err.Error())
	}
	price := &Price{
		ObjectType:               PriceObjectType,
		ResourceProvider:         ResourceProvider,
		ProviderID:               ProviderID,
		CpuPrice:                 CPU_price,
		RamPrice:                 RAM_price,
		HarddiskEcs:              harddiskEcs,
		HarddiskOss:              harddiskOss,
		RequestPrice:             requestPrice,
		NetworkBandwidth_public:  networkbandwidthPublic,
		NetworkBandwidth_private: networkbandwidthPrivate,
	}
	//把对象变成字符串（字节串）
	err = putState(price, ProviderID, stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("Price put state failed.\n%+v", price))
	}
	fmt.Println("-end init price table")
	return shim.Success(nil)
}

//read resource
func (t *MyChaincode) readResource(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var keyID, queryType, jsonResp string
	var err error

	if s := checkArgumentsNum(2, args); len(s) != 0 {
		return shim.Error(s)
	}
	queryType = args[0]
	keyID = args[1]
	if queryType == "price" {
		keyID = keyID + PriceSuffix
	} else if queryType == "quota" {
		keyID = keyID + InitialQuotaSuffix
	} else if queryType == "ECSResource" {
		keyID = keyID + ECSResourceSuffix
	} else if queryType == "OSSResource" {
		keyID = keyID + OSSResourceSuffix
	} else if queryType == "ECSBill" {
		keyID = keyID + ECSBillSuffix
	} else if queryType == "OSSBill" {
		keyID = keyID + OSSBillSuffix
	} else {
		return shim.Error(fmt.Sprintf("invalid param: %v", queryType))
	}
	valAsbytes, err := stub.GetState(keyID) //get the instance from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + keyID + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Your input ID does not exist: " + keyID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

//update quota修改配额
func (t *MyChaincode) updateQuota(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// TODO:check if quota exists
	if s := checkArgumentsNum(13, args); len(s) != 0 {
		return shim.Error(s)
	}

	QuotaID := args[0]
	ResourceProvider := args[1]
	ProviderID := args[2]
	ResourceUser := args[3]
	ResourceUserID := args[4]
	HoldingTime, err := strconv.ParseUint(args[5], 10, 64)
	CPU, err := strconv.ParseUint(args[6], 10, 64)
	RAM, err := strconv.ParseUint(args[7], 10, 64)
	HardDisk_ecs, err := strconv.ParseUint(args[8], 10, 64)
	HardDisk_oss, err := strconv.ParseUint(args[9], 10, 64)
	NetworkBandwidth, err := strconv.ParseUint(args[10], 10, 64)
	PublicIP, err := strconv.ParseUint(args[11], 10, 64)
	PrivateIP, err := strconv.ParseUint(args[12], 10, 64)
	quota := &InitialQuota{
		ObjectType:       InitialQuotaObjectType,
		QuotaID:          QuotaID,
		ResourceProvider: ResourceProvider,
		ProviderID:       ProviderID,
		ResourceUser:     ResourceUser,
		ResourceUserID:   ResourceUserID,
		HoldingTime:      HoldingTime,
		CPU:              CPU,
		RAM:              RAM,
		HardDisk_ecs:     HardDisk_ecs,
		HardDisk_oss:     HardDisk_oss,
		NetworkBandwidth: NetworkBandwidth,
		PublicIP:         PublicIP,
		PrivateIP:        PrivateIP,
	}
	//把对象变成字符串（字节串）
	quotaID := QuotaID
	quotaJSONasByte, err := json.Marshal(quota)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(quotaID, quotaJSONasByte)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//修改价格
func (t *MyChaincode) updatePrice(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// TODO: check if price exists
	if s := checkArgumentsNum(9, args); len(s) != 0 {
		return shim.Error(s)
	}
	ResourceProvider := args[0]
	ProviderID := args[1]
	CPU_price, err := strconv.ParseUint(args[2], 10, 64)
	RAM_price, err := strconv.ParseUint(args[3], 10, 64)
	harddiskEcs, err := strconv.ParseUint(args[4], 10, 64)
	harddiskOss, err := strconv.ParseUint(args[5], 10, 64)
	requestPrice, err := strconv.ParseUint(args[6], 10, 64)
	networkbandwidthPublic, err := strconv.ParseUint(args[7], 10, 64)
	networkbandwidthPrivate, err := strconv.ParseUint(args[8], 10, 64)
	price := &Price{
		ObjectType:               PriceObjectType,
		ResourceProvider:         ResourceProvider,
		ProviderID:               ProviderID,
		CpuPrice:                 CPU_price,
		RamPrice:                 RAM_price,
		HarddiskEcs:              harddiskEcs,
		HarddiskOss:              harddiskOss,
		RequestPrice:             requestPrice,
		NetworkBandwidth_public:  networkbandwidthPublic,
		NetworkBandwidth_private: networkbandwidthPrivate,
	}
	//把对象变成字符串（字节串）
	providerID := ProviderID
	quotaJSONasByte, err := json.Marshal(price)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(providerID+PriceSuffix, quotaJSONasByte)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//billed记账ecs oss
func (t *MyChaincode) billedECS(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var (
		err          error
		ECSResource  *ECSResource
		price        *Price
		priceNetwork uint64
	)
	if s := checkArgumentsNum(3, args); len(s) != 0 {
		return shim.Error(s)
	}
	fmt.Println("- start accounting ECS bill")
	InstanceID := args[0]
	ProviderID := args[1]
	UserID := args[2]
	ECSResource, err = getECSResource(InstanceID, stub)
	if err != nil {
		return shim.Error("Failed to get ECSResource: " + err.Error())
	} else if ECSResource == nil {
		fmt.Println("This ECSResource doesn't exist: " + InstanceID)
		return shim.Error("This ECSResource doesn't exist: " + InstanceID)
	}
	now, err := stub.GetTxTimestamp()
	if err != nil {
		return shim.Error(err.Error())
	}
	usedTime := uint64(now.Seconds-ECSResource.BeginTime.Seconds) / 60 / 60

	price, err = getPrice(ProviderID, stub)
	if err != nil {
		return shim.Error("Failed to get Price: " + err.Error())
	} else if price == nil {
		fmt.Println("This Price doesn't exist: " + InstanceID)
		return shim.Error("This Price doesn't exist: " + InstanceID)
	}
	if ECSResource.NetworkType == "public" {
		priceNetwork = price.NetworkBandwidth_public
	} else {
		priceNetwork = price.NetworkBandwidth_private
	}
	TotalCost := usedTime * (price.CpuPrice*ECSResource.CPU + price.HarddiskEcs*ECSResource.HardDisk + price.RamPrice*ECSResource.RAM + priceNetwork*ECSResource.NetworkBandwidth)

	r, err := getECSBill(InstanceID, stub)
	if err != nil {
		return shim.Error("Failed to get ECSBill: " + err.Error())
	} else if r != nil {
		fmt.Println("This ECSBill already exists: " + InstanceID)
		return shim.Error("This ECSBill already exists: " + InstanceID)
	}
	//create bill object and marshal to JSON

	bill := &ECSBill{
		ObjectType:  ECSBillObjectType,
		ProviderID:  ProviderID,
		UserID:      UserID,
		InstanceID:  InstanceID,
		ECSResource: *ECSResource,
		EndTime:     *now,
		UsageTime:   usedTime,
		TotalCost:   TotalCost,
	}
	err = putState(bill, InstanceID, stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("ECSbill put state failed.\n%+v", bill))
	}
	fmt.Println("-end init ECS bill")
	return shim.Success(nil)
}

func (t *MyChaincode) billedOSS(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var (
		err          error
		OSSResource  *OSSResource
		price        *Price
		priceNetwork uint64
	)
	if s := checkArgumentsNum(5, args); len(s) != 0 {
		return shim.Error(s)
	}
	fmt.Println("- start accounting OSS bill")
	InstanceID := args[0]
	ProviderID := args[1]
	UserID := args[2]
	OSSRequest, err := strconv.ParseUint(args[3], 10, 64)
	FlowOut, err := strconv.ParseUint(args[4], 10, 64)

	OSSResource, err = getOSSResource(InstanceID, stub)
	if err != nil {
		return shim.Error("Failed to get OSSResource: " + err.Error())
	} else if OSSResource == nil {
		fmt.Println("This OSSResource doesn't exist: " + InstanceID)
		return shim.Error("This OSSResource doesn't exist: " + InstanceID)
	}
	now, err := stub.GetTxTimestamp()
	if err != nil {
		return shim.Error(err.Error())
	}
	usedTime := uint64(now.Seconds-OSSResource.BeginTime.Seconds) / 60 / 60

	price, err = getPrice(ProviderID, stub)
	if err != nil {
		return shim.Error("Failed to get Price: " + err.Error())
	} else if price == nil {
		fmt.Println("This Price doesn't exist: " + InstanceID)
		return shim.Error("This Price doesn't exist: " + InstanceID)
	}
	if OSSResource.NetworkType == "public" {
		priceNetwork = price.NetworkBandwidth_public
	} else {
		priceNetwork = price.NetworkBandwidth_private
	}
	TotalCost := usedTime * (price.HarddiskOss*OSSResource.HardDisk + price.RequestPrice*OSSRequest + priceNetwork*FlowOut)

	r, err := getOSSBill(InstanceID, stub)
	if err != nil {
		return shim.Error("Failed to get OSSBill: " + err.Error())
	} else if r != nil {
		fmt.Println("This OSSBill already exists: " + InstanceID)
		return shim.Error("This OSSBill already exists: " + InstanceID)
	}
	//create bill object and marshal to JSON

	bill := &OSSBill{
		ObjectType:  OSSBillObjectType,
		ProviderID:  ProviderID,
		UserID:      UserID,
		InstanceID:  InstanceID,
		OSSresource: *OSSResource,
		EndTime:     *now,
		FlowOut:     FlowOut,
		OSSRequest:  OSSRequest,
		UsageTime:   usedTime,
		TotalCost:   TotalCost,
	}
	err = putState(bill, InstanceID, stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("OSSBill put state failed.\n%+v", bill))
	}
	fmt.Println("-end init OSS bill")
	return shim.Success(nil)
}

//查询某provider的所有账单【未完成、已完成、合计金额】
func (t *MyChaincode) queryProviderBill(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//查询某provider所有已完成的账单[云主机、对象存储分开查询]
	if s := checkArgumentsNum(2, args); len(s) != 0 {
		return shim.Error(s)
	}
	providerID := args[0]
	objectType := args[1]
	if objectType != "ECSBill" && objectType != "OSSBill" {
		return shim.Error(fmt.Sprintf("invalid object type: %v, valid object types: ECSBill, OSSBill", objectType))
	}
	queryString := fmt.Sprintf(`{"selector":{"providerID":"%s","objectType":"%s"}}`, providerID, objectType)
	resultsIterator, err := stub.GetQueryResult(queryString)
	// 富查询的返回结果可能为多条 所以这里返回的是一个迭代器 需要我们进一步的处理来获取需要的结果
	if err != nil {
		return shim.Error(fmt.Sprintf("Rich query failed: %v", err))
	}
	defer resultsIterator.Close()

	results := make([]string, 0)

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Fail")
		}
		results = append(results, string(queryResponse.Value))
	}
	result := "{\"result\":[" + strings.Join(results, ",") + "]}"
	fmt.Printf("Query result: %s", result)

	return shim.Success([]byte(result))
}
func (t *MyChaincode) queryProviderTotalCost(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//查询某provider所有已完成的账单的总金额[ECS OSS分开]
	if s := checkArgumentsNum(2, args); len(s) != 0 {
		return shim.Error(s)
	}
	providerID := args[0]
	objectType := args[1]
	if objectType != "ECSBill" && objectType != "OSSBill" {
		return shim.Error(fmt.Sprintf("invalid object type: %v, valid object types: ECSBill, OSSBill", objectType))
	}
	queryString := fmt.Sprintf(`{"selector":{"providerID":"%s","objectType":"%s"}}`, providerID, objectType)
	resultsIterator, err := stub.GetQueryResult(queryString)
	// 富查询的返回结果可能为多条 所以这里返回的是一个迭代器 需要我们进一步的处理来获取需要的结果
	if err != nil {
		return shim.Error(fmt.Sprintf("Rich query failed: %v", err))
	}
	defer resultsIterator.Close()
	var totalCost uint64

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Fail")
		}
		if objectType == "OSSBill" {
			var ossBill OSSBill
			err = json.Unmarshal(queryResponse.Value, &ossBill)
			if err != nil {
				return shim.Error(err.Error())
			}
			totalCost += ossBill.TotalCost
		} else if objectType == "ECSBill" {
			var ecsBill ECSBill
			err = json.Unmarshal(queryResponse.Value, &ecsBill)
			if err != nil {
				return shim.Error(err.Error())
			}
			totalCost += ecsBill.TotalCost
		}
	}
	return shim.Success([]byte(fmt.Sprintf("%v", totalCost)))
}

//查询provider里所有未完成的账单

//查询某用户所有账单【未完成、已完成、合计金额】
//某用户已完成账单
func (t *MyChaincode) queryUserBill(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//查询某provider所有已完成的账单[云主机、对象存储分开查询]
	if s := checkArgumentsNum(2, args); len(s) != 0 {
		return shim.Error(s)
	}
	userID := args[0]
	objectType := args[1]
	if objectType != "ECSBill" && objectType != "OSSBill"{
		return shim.Error(fmt.Sprintf("invalid object type: %v, valid object types: ECSBill, OSSBill",objectType))
	}
	queryString := fmt.Sprintf(`{"selector":{"userID":"%s","objectType":"%s"}}`, userID, objectType)
	resultsIterator, err := stub.GetQueryResult(queryString)
	// 富查询的返回结果可能为多条 所以这里返回的是一个迭代器 需要我们进一步的处理来获取需要的结果
	if err != nil {
		return shim.Error(fmt.Sprintf("Rich query failed: %v",err))
	}
	defer resultsIterator.Close()

	results := make([]string,0)

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Fail")
		}
		results = append(results,string(queryResponse.Value))
	}
	result := "{\"result\":[" + strings.Join(results, ",") + "]}"
	fmt.Printf("Query result: %s", result)

	return shim.Success([]byte(result))
}

//某用户已完成账单合计金额
func (t *MyChaincode) queryUserTotalCost(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//查询某provider所有已完成的账单的总金额[ECS OSS分开]
	if s := checkArgumentsNum(2, args); len(s) != 0 {
		return shim.Error(s)
	}
	userID := args[0]
	objectType := args[1]
	if objectType != "ECSBill" && objectType != "OSSBill" {
		return shim.Error(fmt.Sprintf("invalid object type: %v, valid object types: ECSBill, OSSBill", objectType))
	}
	queryString := fmt.Sprintf(`{"selector":{"userID":"%s","objectType":"%s"}}`, userID, objectType)
	resultsIterator, err := stub.GetQueryResult(queryString)
	// 富查询的返回结果可能为多条 所以这里返回的是一个迭代器 需要我们进一步的处理来获取需要的结果
	if err != nil {
		return shim.Error(fmt.Sprintf("Rich query failed: %v", err))
	}
	defer resultsIterator.Close()
	var totalCost uint64

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Fail")
		}
		if objectType == "OSSBill"{
			var ossBill OSSBill
			err = json.Unmarshal(queryResponse.Value, &ossBill)
			if err != nil {
				return shim.Error(err.Error())
			}
			totalCost += ossBill.TotalCost
		}else if objectType == "ECSBill"{
			var ecsBill ECSBill
			err = json.Unmarshal(queryResponse.Value,&ecsBill)
			if err != nil{
				return shim.Error(err.Error())
			}
			totalCost += ecsBill.TotalCost
		}

	}
	return shim.Success([]byte(fmt.Sprintf("%v", totalCost)))
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(MyChaincode)); err != nil {
		fmt.Printf("Error starting MyChaincode chaincode: %s", err)
	}
}

func checkArgumentsNum(expect int, args []string) string {
	if len(args) != expect {
		return fmt.Sprintf("Incorrect arguments, expecting %v arguments, get %v arguments: %v", expect, len(args), args)
	}
	return ""
}

func getState(id string, stub shim.ChaincodeStubInterface) ([]byte, error) {
	bytes, err := stub.GetState(id)
	if err != nil {
		return nil, errors.Wrapf(err, "get id: %v failed", id)
	}
	return bytes, nil
}

func getInitialQuota(id string, stub shim.ChaincodeStubInterface) (*InitialQuota, error) {
	var s InitialQuota
	bytes, err := getState(id+InitialQuotaSuffix, stub)
	if err != nil {
		return nil, err
	} else if len(bytes) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(bytes, &s)
	if err != nil {
		return nil, errors.Wrapf(err, "json unmarshal failed: %v", string(bytes))
	}
	return &s, nil
}

func getECSResource(id string, stub shim.ChaincodeStubInterface) (*ECSResource, error) {
	var s ECSResource
	bytes, err := getState(id+ECSResourceSuffix, stub)
	if err != nil {
		return nil, err
	} else if len(bytes) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(bytes, &s)
	if err != nil {
		return nil, errors.Wrapf(err, "json unmarshal failed: %v", string(bytes))
	}
	return &s, nil
}
func getECSBill(id string, stub shim.ChaincodeStubInterface) (*ECSBill, error) {
	var s ECSBill
	bytes, err := getState(id+ECSBillSuffix, stub)
	if err != nil {
		return nil, err
	} else if len(bytes) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(bytes, &s)
	if err != nil {
		return nil, errors.Wrapf(err, "json unmarshal failed: %v", string(bytes))
	}
	return &s, nil
}
func getOSSResource(id string, stub shim.ChaincodeStubInterface) (*OSSResource, error) {
	var s OSSResource
	bytes, err := getState(id+OSSResourceSuffix, stub)
	if err != nil {
		return nil, err
	} else if len(bytes) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(bytes, &s)
	if err != nil {
		return nil, errors.Wrapf(err, "json unmarshal failed: %v", string(bytes))
	}
	return &s, nil
}
func getOSSBill(id string, stub shim.ChaincodeStubInterface) (*OSSBill, error) {
	var s OSSBill
	bytes, err := getState(id+OSSBillSuffix, stub)
	if err != nil {
		return nil, err
	} else if len(bytes) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(bytes, &s)
	if err != nil {
		return nil, errors.Wrapf(err, "json unmarshal failed: %v", string(bytes))
	}
	return &s, nil
}
func getPrice(id string, stub shim.ChaincodeStubInterface) (*Price, error) {
	var s Price
	bytes, err := getState(id+PriceSuffix, stub)
	if err != nil {
		return nil, err
	} else if len(bytes) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(bytes, &s)
	if err != nil {
		return nil, errors.Wrapf(err, "json unmarshal failed: %v", string(bytes))
	}
	return &s, nil
}

func putState(s interface{}, id string, stub shim.ChaincodeStubInterface) error {
	bytes, err := json.Marshal(s)
	if err != nil {
		return errors.Wrapf(err, "json marshal failed: %v", reflect.TypeOf(s).String())
	}
	typeName := getType(s)
	return stub.PutState(id+"_"+typeName, bytes)
}

func getType(s interface{}) string {
	if t := reflect.TypeOf(s); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

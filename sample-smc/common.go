package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

//MSPList ...
type MSPList struct {
	OrgType string `json:"orgType"`
	OrgCode string `json:"orgCode"`
	MSP     string `json:"MSP"`
	ID      string `json:"ID"`
}

type errCode struct {
	ErrorCode string   `json:"errorCode"`
	Options   []string `json:"options"`
}

func GetTxID(stub *hypConnect) (string, string) {
	str := stub.Connection.GetTxID()
	_, args := stub.Connection.GetFunctionAndParameters()
	fmt.Println("\n\nARG >> ", args)
	fmt.Println("\n\n\n\n ARGS[req] >> ", args[len(args)-1])
	fmt.Println("\n\n")
	str2 := args[len(args)-1]
	return str, str2
}

func getCustomsMSP() string {
	return "org2MSP"
}

func insertData(stub *hypConnect, key string, privateCollection string, data []byte) error {

	err := stub.Connection.PutPrivateData(privateCollection, key, data)
	if err != nil {
		return err
	}

	event := eventDataFormat{}
	event.Key = key
	event.Collection = privateCollection
	stub.EventList = stub.AddEvent(event)

	fmt.Println("Successfully Put State for Key: " + key + " and Private Collection " + privateCollection)
	return nil
}

func insertDataEP(stub *hypConnect, key string, privateCollection string, data []byte, ep []byte) error {

	err := stub.Connection.SetPrivateDataValidationParameter(key, privateCollection, ep)
	if err != nil {
		return err
	}

	err = stub.Connection.PutPrivateData(privateCollection, key, ep)
	if err != nil {
		return err
	}
	event := eventDataFormat{}
	event.Key = key
	event.Collection = privateCollection
	stub.EventList = stub.AddEvent(event)

	fmt.Println("Successfully Put State for Key: " + key + " and Private Collection " + privateCollection)
	return nil
}

func fetchDataEP(stub hypConnect, key string, privateCollection string) ([]byte, error) {

	_, err := stub.Connection.GetPrivateDataValidationParameter(privateCollection, key)
	if err != nil {
		return nil, err
	}

	bytes, err := stub.Connection.GetPrivateData(privateCollection, key)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func fetchDataPublic(stub hypConnect, key string) ([]byte, error) {
	bytes, err := stub.Connection.GetState(key)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func fetchData(stub hypConnect, key string, privateCollection string) ([]byte, error) {
	bytes, err := stub.Connection.GetPrivateData(privateCollection, key)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func getArguments(stub shim.ChaincodeStubInterface) ([]string, error) {
	transMap, err := stub.GetTransient()
	if err != nil {
		return nil, err
	}

	fmt.Println("\n\ntransMap =========== ", transMap)

	if _, ok := transMap["PrivateArgs"]; !ok {
		return nil, errors.New("PrivateArgs must be a key in the transient map")
	}
	fmt.Printf("Arguments: %v", transMap)
	generalInput := string(transMap["PrivateArgs"])
	retVal := strings.Split(generalInput, "|")
	fmt.Printf("retVal: %v", retVal)
	return retVal, nil
}

func deleteData(stub *hypConnect, key string, privateCollection string) error {

	err := stub.Connection.DelPrivateData(privateCollection, key)
	if err != nil {
		return err
	}
	fmt.Println("Successfully Delete for Key: " + key + " and Private Collection " + privateCollection)

	event := eventDataFormat{}
	event.Key = key
	event.Collection = privateCollection
	stub.EventList = stub.AddEvent(event)

	return nil
}

func getOrgTypeByJWTOrgCode(orgCode string) (string, error) {
	switch orgCode {
	case "Aramex":
		return "F", nil
	case "DHL":
		return "F", nil

	default:
		return "E", nil

	}
}
func getOrgTypeByMSP(stub shim.ChaincodeStubInterface, MSP string) (string, error) {

	MSPMappingAsBytes, err := stub.GetState("MSPMapping")
	if err != nil {
		return "", err
	}

	if err != nil {
		fmt.Println("MSPMapping - Failed to get state MSP mapping information." + err.Error())
		return "", err
	} else if MSPMappingAsBytes != nil {
		fmt.Println("MSPMapping - This data Fetched from Transactions.")
		var MSPListUnmarshaled []MSPList

		err := json.Unmarshal(MSPMappingAsBytes, &MSPListUnmarshaled)
		if err != nil {
			fmt.Println("MSPMapping-Failed to UnMarshal state.")
			return "", err
		}
		fmt.Printf("Unmarshaled: %v", MSPListUnmarshaled)
		for i := 0; i < len(MSPListUnmarshaled); i++ {
			if MSPListUnmarshaled[i].MSP == MSP {
				fmt.Println("OrgType for MSP " + MSP + " is " + MSPListUnmarshaled[i].OrgType)
				return MSPListUnmarshaled[i].OrgType, nil
			}
		}
	}
	return "", nil
}
func (t *Customs) ValidateOrgCode(stub shim.ChaincodeStubInterface, orgCode string) bool {

	validOrgCode := false
	//getting MSP
	certOrgType, err := cid.GetMSPID(stub)
	if err != nil {
		fmt.Println("Error occurred while fetching msp id")
		return validOrgCode
	}

	fmt.Println("MSP : " + certOrgType)
	orgCodeList, err := getOrgCodeList(stub, string(certOrgType))
	if err != nil {
		fmt.Println("Error occurred while fetching orgcode list")
		return validOrgCode
	}
	for i := 0; i < len(orgCodeList); i++ {
		if orgCode == orgCodeList[i] {
			validOrgCode = true
			break
		}
	}
	return validOrgCode
}
func getOrgCodeList(stub shim.ChaincodeStubInterface, MSP string) ([]string, error) {

	var orgCodeList []string
	MSPMappingAsBytes, err := stub.GetState("MSPMapping")
	if err != nil {
		fmt.Println("MSPMapping - Failed to get state MSP mapping information." + err.Error())
		return orgCodeList, err
	} else if MSPMappingAsBytes != nil {
		fmt.Println("MSPMapping - This data Fetched from Transactions.")
		var MSPListUnmarshaled []MSPList

		err := json.Unmarshal(MSPMappingAsBytes, &MSPListUnmarshaled)
		if err != nil {
			fmt.Println("MSPMapping-Failed to UnMarshal state.")
			return orgCodeList, err
		}
		fmt.Println("\n Unmarshaled MSPMappingAsBytes: ", MSPListUnmarshaled)
		for i := 0; i < len(MSPListUnmarshaled); i++ {
			orgCodeList = append(orgCodeList, MSPListUnmarshaled[i].OrgCode)
		}
	}
	return orgCodeList, nil
}
func getOrgTypeByOrgCode(stub shim.ChaincodeStubInterface, orgCode string) (string, error) {

	MSPMappingAsBytes, err := stub.GetState("MSPMapping")
	if err != nil {
		return "", err
	}

	if err != nil {
		fmt.Println("MSPMapping - Failed to get state MSP mapping information." + err.Error())
		return "", err
	} else if MSPMappingAsBytes != nil {
		fmt.Println("MSPMapping - This data Fetched from Transactions.")
		var MSPListUnmarshaled []MSPList

		err := json.Unmarshal(MSPMappingAsBytes, &MSPListUnmarshaled)
		if err != nil {
			fmt.Println("MSPMapping-Failed to UnMarshal state.")
			return "", err
		}
		fmt.Printf("Unmarshaled: %v", MSPListUnmarshaled)
		for i := 0; i < len(MSPListUnmarshaled); i++ {
			if MSPListUnmarshaled[i].OrgCode == orgCode {
				fmt.Println("OrgType for OrgCode " + orgCode + " is " + MSPListUnmarshaled[i].OrgType)
				return MSPListUnmarshaled[i].OrgType, nil
			}
		}
	}
	return "", nil
}

//RaiseEventData
func RaiseEventData(stub hypConnect, eventName string, args ...interface{}) (string, error) {

	var eventList generalEventStruct
	eventList.EventName = eventName
	eventList.EventList = stub.EventList
	eventList.AdditionalData = args
	eventJSONasBytes, err2 := json.Marshal(eventList)
	if err2 != nil {
		return "", err2
	}
	fmt.Println("Event raised: " + eventName)
	//fmt.Println("\neventJSONasBytes : ", eventList.EventName+"\n")
	mEventName := eventList.EventName
	err3 := stub.Connection.SetEvent("chainCodeEvent", []byte(eventJSONasBytes))
	if err3 != nil {
		return "", err3
	}
	var err4 error
	err4 = nil
	return mEventName, err4

}

func orgCodeVerify(stub hypConnect, orgCode string) error {

	return nil
	m := make(map[string]string)

	m["GEMS"] = "gemsMSP"
	m["ETISALAT"] = "etisalatMSP"

	mspID, err := cid.GetMSPID(stub.Connection)

	if err != nil {

		return errors.New("orgCodeVerify getting msp error " + err.Error())
	} else if m[orgCode] == mspID || "WaslMSP" == mspID {

		fmt.Println("Your MSP ID is >>>" + mspID)
		return nil
	}
	return errors.New("you are not allowed to performed this action using this msp " + mspID)
}

//GetDataByKey ...
func (t *Customs) GetDataByKey(stub hypConnect, args []string, controller string) pb.Response {
	fmt.Println("GetDataByKey: ", args)

	if len(args[0]) <= 0 {
		return shim.Error("Invalid Argument")
	}
	key := sanitize(args[0], "string").(string)
	collection := sanitize(args[1], "string").(string)
	trnxAsBytes, err := fetchData(stub, key, collection)
	if err != nil {
		fmt.Println("No Data found with Key: " + key)
		return shim.Error("No Data found with key: " + err.Error())
	}
	if trnxAsBytes == nil {
		return shim.Error("No data found with key: ")
	}
	return shim.Success(trnxAsBytes)
}

//GetDataByKey ... For Public Collection
func (t *Customs) GetDataByKeyPublic(stub hypConnect, args []string, controller string) pb.Response {
	fmt.Println("GetDataByKeyPublic: ", args)

	if len(args[0]) <= 0 {
		return shim.Error("Invalid Argument")
	}
	key := sanitize(args[0], "string").(string)
	trnxAsBytes, err := fetchDataPublic(stub, key)
	if err != nil {
		fmt.Println("No Data found with Key: " + key)
		return shim.Error("No Data found with key: " + err.Error())
	}
	if trnxAsBytes == nil {
		return shim.Error("No data found with key: ")
	}
	return shim.Success(trnxAsBytes)
}

//getRecordByID ...
func (t *Customs) getRecordByID(stub hypConnect, ID string, collection string, docName string) ([]byte, error) {

	fmt.Println("getRecordByID : ", ID)
	trnxAsBytes, err := fetchData(stub, ID, collection)
	if err != nil {
		fmt.Println("Error occurred while fetching  " + docName + " " + err.Error())
		return nil, errors.New("Error occurred while fetching  " + docName + " " + err.Error())
	}
	if trnxAsBytes == nil {
		return nil, errors.New(docName + " not found")
	}
	return trnxAsBytes, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func GetTxIDCommon(stub *hypConnect) (string, string) {
	//return "21789798213", "21789798213"
	return GetTxID(stub)
}

func genericInsertData(stub *hypConnect, key string, collection string, v interface{}) error {
	dataAsBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

	insertErr := insertData(stub, key, collection, dataAsBytes)
	if insertErr != nil {
		return insertErr
	}
	return nil

}
func convertToChainCodeArgs(args []string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

func prepareErrorCode(errorCode string, options []string) ([]byte, error) {
	errCode := errCode{}
	errCode.ErrorCode = errorCode
	errCode.Options = options

	errCodeasBytes, errCodeMarshalError := json.Marshal(errCode)
	if errCodeMarshalError != nil {
		return nil, errCodeMarshalError
	}
	//return shim.Success(errCodeasBytes)
	return errCodeasBytes, nil
}

func genericFetchData(stub *hypConnect, key string, collection string, v interface{}) (error, bool) {

	fmt.Println("\n\n key > ", key)

	dataBytes, err := getDataByKey(*stub, key, collection)

	if err != nil {
		return err, false
	}

	if dataBytes == nil {
		fmt.Println("\n dataAsBytes nil")
		return nil, false
	}

	errUnMarsh := json.Unmarshal(dataBytes, v)
	if errUnMarsh != nil {
		fmt.Println("\n\n\n >> ", errUnMarsh)
		return errUnMarsh, false
	}

	return nil, true
}

func getDataByKey(stub hypConnect, key string, collectionName string) ([]byte, error) {

	dataAsBytes, errorMsg := fetchData(stub, key, collectionName)
	if errorMsg != nil {
		return nil, errorMsg
	}

	return dataAsBytes, nil
}

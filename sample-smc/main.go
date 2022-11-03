package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type Customs struct {
}

// ///Standard Functions
func main() {

	fmt.Println("Dubai Customs ChainCode Started")
	err := shim.Start(new(Customs))
	if err != nil {
		fmt.Printf("Error starting DC chaincode: %s", err)
	}

}

// Init is called during chaincode instantiation to initialize any data.
func (t *Customs) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("DC ChainCode Initiated")

	_, args := stub.GetFunctionAndParameters()
	fmt.Printf("Init: %v", args)
	if len(args[0]) <= 0 {
		return shim.Error("MSP Mapping information is required for initiating the chain code")
	}

	var MSPListUnmarshaled []MSPList
	err := json.Unmarshal([]byte(args[0]), &MSPListUnmarshaled)

	if err != nil {
		return shim.Error("An error occurred while Unmarshiling MSPMapping: " + err.Error())
	}
	MSPMappingJSONasBytes, err := json.Marshal(MSPListUnmarshaled)
	if err != nil {
		return shim.Error("An error occurred while Marshiling MSPMapping :" + err.Error())
	}
	_Key := "MSPMapping"
	err = stub.PutState(_Key, []byte(MSPMappingJSONasBytes))
	if err != nil {
		return shim.Error("An error occurred while inserting MSPMapping:" + err.Error())
	}
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode
func (t *Customs) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	//getting MSP =====================///
	certOrgType, err := cid.GetMSPID(stub)
	if err != nil {
		return shim.Error("Enrollment mspid Type invalid !!! " + err.Error())
	}
	fmt.Println("MSP:" + certOrgType)

	orgType, err := getOrgTypeByMSP(stub, string(certOrgType))

	if err != nil {
		return shim.Error(err.Error())
	}

	function, tempArgs := stub.GetFunctionAndParameters()
	fmt.Println("Invoke is running for function: " + function)
	fmt.Println("\n\n TEMP ARGS >> ", tempArgs)

	args, errArgs := getArguments(stub)
	if errArgs != nil {
		return shim.Error(errArgs.Error())
	}

	connection := hypConnect{}
	connection.Connection = stub

	if allowed, checkForSignature := ValidateAccessControl(function, orgType); allowed {
		fmt.Println("checkForSignature: ", checkForSignature)
		fmt.Println("allowed: ", allowed)

		switch functionName := function; functionName {
		case "createAsset":
			return t.createAsset(connection, args, "createAsset", certOrgType)
		case "transferAsset":
			return t.transferAsset(connection, args, "transferAsset", certOrgType)
		case "getAssetData":
			return t.getAssetData(connection, args, "getAssetData")
		default:
			fmt.Printf("Invoke did not find function: " + function)
			return shim.Error("Received unknown function invocation: " + function)
		}
	} else {
		return shim.Error("Invalid MSP: " + orgType)
	}
}

func (t *Customs) createAsset(stub hypConnect, args []string, functionName string, MSPID string) pb.Response {
	fmt.Println("createData_args-----------> ", args)

	if len(args[0]) <= 0 {
		return shim.Error("Invalid Argument")
	}

	var data Asset
	var assetId = sanitize(args[0], "string").(string)

	data.AssetId = assetId
	data.AssetName = sanitize(args[1], "string").(string)
	data.OwnedByUser = sanitize(args[2], "string").(string)
	data.AssetType = sanitize(args[3], "string").(string)
	data.OrgCode = sanitize(args[4], "string").(string)

	data.DocumentName = "osamaasset"
	data.Key = assetId

	dataAsBytes, errorMarshal := json.Marshal(data)
	if errorMarshal != nil {
		return shim.Error("Error while Marshalling Data -----> " + errorMarshal.Error())
	}
	errorInsert := insertData(&stub, data.Key, data.DocumentName, []byte(dataAsBytes))
	if errorInsert != nil {
		fmt.Println("Insertion failed of keysMarshalled", errorInsert)
		return shim.Error(errorInsert.Error())
	}

	RaiseEventData(stub, "Training")

	return shim.Success(nil)
}

func (t *Customs) transferAsset(stub hypConnect, args []string, functionName string, MSPID string) pb.Response {
	fmt.Println("transferData_args-----------> ", args)

	if len(args[0]) <= 0 {
		return shim.Error("Invalid Argument")
	}
	fmt.Println("MSPID>>>>>>>>", MSPID)

	var data Transaction
	var asset Asset

	var assetId = sanitize(args[0], "string").(string)
	var timeStamp = sanitize(args[1], "string").(string)
	var fromUser = sanitize(args[2], "string").(string)
	var toUser = sanitize(args[3], "string").(string)
	var fromUserOrgCode = sanitize(args[4], "string").(string)
	var toUserOrgCode = sanitize(args[5], "string").(string)

	trnxAsBytes, err := fetchData(stub, assetId, "osamaasset")
	if err != nil {
		fmt.Println("No Data found with Key: " + assetId)
		return shim.Error("No Data found with key: " + err.Error())
	}
	if trnxAsBytes == nil {
		return shim.Error("No data found with key: ")
	}

	//unmarshal Schedule Data
	err = json.Unmarshal([]byte(trnxAsBytes), &asset)
	if err != nil {
		return shim.Error("Error while unmarshal Data " + err.Error())
	}
	//Snder cannot be receiver
	if asset.OwnedByUser == toUser {
		return shim.Error("Sender and receiver cannot be same person please choose another address!")
	}
	//sender cannot be receiver by arguments
	if fromUser == toUser {
		return shim.Error("Sender and receiver cannot be same person please choose another address!")
	}

	fmt.Println("Condition >>>", asset, asset.OwnedByUser, fromUser)
	if asset.OwnedByUser != fromUser {
		fmt.Println("Owner>>>", asset.OwnedByUser)
		fmt.Println("fromUser", fromUser)
		return shim.Error("User is not the owner!")
	}

	if fromUserOrgCode == toUserOrgCode {
		orgCode := fromUserOrgCode

		data.Key = assetId

		if orgCode == "ORG1" {
			data.DocumentName = "osamatransactionorg1"
		} else {
			data.DocumentName = "osamatransactionorg2"
		}

		data.AssetId = assetId
		data.Timestamp = timeStamp
		data.FromUser = fromUser
		data.ToUser = toUser

		dataAsBytes, errorMarshal := json.Marshal(data)
		if errorMarshal != nil {
			return shim.Error("Error while Marshalling Data -----> " + errorMarshal.Error())
		}
		errorInsert := insertData(&stub, data.Key, data.DocumentName, []byte(dataAsBytes))
		if errorInsert != nil {
			fmt.Println("Insertion failed of keysMarshalled", errorInsert)
			return shim.Error(errorInsert.Error())
		}
	} else {

		data.Key = assetId
		data.AssetId = assetId
		data.Timestamp = timeStamp
		data.FromUser = fromUser
		data.ToUser = toUser

		data.DocumentName = "osamatransactionorg1"

		dataAsBytes, errorMarshal := json.Marshal(data)
		if errorMarshal != nil {
			return shim.Error("Error while Marshalling Data -----> " + errorMarshal.Error())
		}
		errorInsert := insertData(&stub, data.Key, data.DocumentName, []byte(dataAsBytes))
		if errorInsert != nil {
			fmt.Println("Insertion failed of keysMarshalled", errorInsert)
			return shim.Error(errorInsert.Error())
		}

		//FOR 2 ORG
		data.DocumentName = "osamatransactionorg2"

		dataAsBytes, errorMarshal = json.Marshal(data)
		if errorMarshal != nil {
			return shim.Error("Error while Marshalling Data -----> " + errorMarshal.Error())
		}
		errorInsert = insertData(&stub, data.Key, data.DocumentName, []byte(dataAsBytes))
		if errorInsert != nil {
			fmt.Println("Insertion failed of keysMarshalled", errorInsert)
			return shim.Error(errorInsert.Error())
		}

	}

	asset.Key = assetId
	asset.OwnedByUser = toUser
	asset.DocumentName = "osamaasset"

	dataAsBytes, errorMarshal := json.Marshal(asset)
	if errorMarshal != nil {
		return shim.Error("Error while Marshalling Asset Data -----> " + errorMarshal.Error())
	}
	errorInsert := insertData(&stub, asset.Key, asset.DocumentName, []byte(dataAsBytes))
	if errorInsert != nil {
		fmt.Println("Insertion failed of keysMarshalled", errorInsert)
		return shim.Error(errorInsert.Error())
	}

	RaiseEventData(stub, "Training")

	return shim.Success(nil)
}

func (t *Customs) getAssetData(stub hypConnect, args []string, controller string) pb.Response {
	fmt.Println(" getAssetData----->>> ", args)

	if len(args[0]) <= 0 {
		return shim.Error("Invalid Argument")
	}
	key := sanitize(args[0], "string").(string)
	collection := "osamaasset"
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

func (t *Customs) postData(stub hypConnect, args []string, functionName string, MSPID string) pb.Response {

	fmt.Println("postData_args-----------> ", args)

	var data twoorgsstructure

	data.Feild1 = sanitize(args[0], "string").(string)
	data.Feild2 = sanitize(args[1], "string").(string)
	data.Feild3 = sanitize(args[2], "string").(string)
	data.Feild4 = sanitize(args[3], "string").(string)
	data.DocumentName = "twoorgsdata"
	data.Key = data.Feild1

	dataAsBytes, errorMarshal := json.Marshal(data)
	if errorMarshal != nil {
		return shim.Error("Error while Marshalling Data -----> " + errorMarshal.Error())
	}

	errorInsert := insertData(&stub, data.Key, data.DocumentName, []byte(dataAsBytes))
	if errorInsert != nil {
		fmt.Println("Insertion failed of keysMarshalled", errorInsert)
		return shim.Error(errorInsert.Error())
	}

	RaiseEventData(stub, "SampleEvent")

	return shim.Success(nil)
}

func (t *Customs) getData(stub hypConnect, args []string, controller string) pb.Response {
	fmt.Println("getData----->>> ", args)

	if len(args[0]) <= 0 {
		return shim.Error("Invalid Argument")
	}
	key := sanitize(args[0], "string").(string)
	collection := "twoorgsdata"
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

type twoorgsstructure struct {
	DocumentName string `json:"documentName"`
	Key          string `json:"key"`
	Feild1       string `json:"Feild1"`
	Feild2       string `json:"Feild2"`
	Feild3       string `json:"Feild3"`
	Feild4       string `json:"Feild4"`
}

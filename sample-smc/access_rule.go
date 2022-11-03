package main

import (
	"fmt"
)

const (
	channelName           = "avanzachannel"
	validationContract    = "validationContract"
	configurationContract = "configurationContract"
	typedataContract      = "typedatatest"
)

type functionInfo struct {
	functionName        string
	isSignatureRequired bool
}

func InitializeMap() map[string][]functionInfo {
	accessControlMap := make(map[string][]functionInfo)

	trainingFuncArrForOrg1 := []functionInfo{

		{"createAsset", false},
		{"transferAsset", false},
		{"getAssetData", false},
	}

	trainingFuncArrForOrg2 := []functionInfo{

		{"createAsset", false},
		{"transferAsset", false},
		{"getAssetData", false},
	}

	//MAP
	accessControlMap["Org1"] = append(accessControlMap["Org1"], trainingFuncArrForOrg1...)
	accessControlMap["Org2"] = append(accessControlMap["Org2"], trainingFuncArrForOrg2...)
	// accessControlMap["Customs"] = append(accessControlMap["General"], generalFuncArr...)

	return accessControlMap
}

func ValidateAccessControl(functionName string, orgType string) (bool, bool) {

	fmt.Println("\n\n ValidateAccessControl orgType >> ", orgType)
	accessControlMap := InitializeMap()
	if functionInfoArr, ok := accessControlMap[orgType]; ok {
		for _, funcInfo := range functionInfoArr {
			if funcInfo.functionName == functionName {
				return true, funcInfo.isSignatureRequired // comment if want to bypass signature
			}
		}
		return false, false
	} else {
		return false, false
	}
}

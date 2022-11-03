package main

// // Document associations
// type Document struct {
// 	Hash string `json:"hash"`
// 	Name string `json:"name"`
// 	Path string `json:"path"`
// }

// // Exemption /s applied
// type Exemption struct {
// 	ExemptionType  string `json:"exemptionType"`
// 	ExemptionRefNo string `json:"exemptionRefNo"`
// }

// // SoapPayload detail
// type SoapPayload struct {
// 	Hash string `json:"hash"`
// 	Name string `json:"name"`
// 	Path string `json:"path"`
// }

// Asset
type Asset struct {
	DocumentName string `json:"documentName"`
	Key          string `json:"key"`
	AssetId      string `json:"assetId"`
	AssetName    string `json:"assetName"`
	OwnedByUser  string `json:"ownedByUser"`
	AssetType    string `json:"assetType"`
	OrgCode      string `json:"orgCode"`
}

// Transaction
type Transaction struct {
	DocumentName    string `json:"documentName"`
	Key             string `json:"key"`
	AssetId         string `json:"assetId"`
	Timestamp       string `json:"time"`
	FromUser        string `json:"fromUser"`
	ToUser          string `json:"toUser"`
	TxnIdBlockchain string `json:"txnIdBlockchain"`
}

// // Sku code & unit
// type Sku struct {
// 	ProductCode string `json:"productCode"`
// 	QuantityUOM string `json:"quantityUOM"`
// }

// // ErrorReturnStruct
// type ErrorReturnStruct struct {
// 	Error   string      `json:"error"`
// 	Options interface{} `json:"options"`
// }

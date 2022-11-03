package main

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
//	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	// "time"

	// "github.com/hyperledger/fabric/core/chaincode/lib/cid"
	// pb "github.com/hyperledger/fabric/protos/peer"
	// "github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	// pb "github.com/hyperledger/fabric-protos-go/peer"
//	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type AuditLogs struct {
	Key             string `json:"key"` //TransactionId - blockchain. //orgcode msp is also to be added and type of document.
	DocumentName    string `json:"documentName"`
	ReferenceId     string `json:"referenceId"`
	CipherMessageId string `json:"cipherMessageId"`
	Json            string `json:"jsonPayload"`
	Status          string `json:"status"`
	Timestamp       int64  `json:"timestamp"`
	Signature       string `json:"signature"`
	OrgCode         string `json:"orgCode"`
}

type rsaPublicKey struct {
	*rsa.PublicKey
}

// using StrAtegy Pattern
type Unsigner interface {
	Unsign(data []byte, sig string, sharedSecret []byte) error
}

func newUnsignerFromKey(k interface{}) (Unsigner, error) {
	var sshKey Unsigner
	switch t := k.(type) {
	case *rsa.PublicKey:
		sshKey = &rsaPublicKey{t}
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %T", k)
	}
	return sshKey, nil
}

func parsePublicKey(pemBytes []byte) (Unsigner, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("ssh: no key found")
	}

	var rawkey interface{}
	switch block.Type {
	case "PUBLIC KEY":
		rsa, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rawkey = rsa
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %q", block.Type)
	}

	return newUnsignerFromKey(rawkey)
}

func loadPublicKey(stub *hypConnect, path string) (Unsigner, error) {

	var publicKey = []byte(`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDD7eZUIfsi9iFxM8iXDYQMoybqA0leNzWGftpExaW/7yhEv7CAHxW64p9DDYzzuhmykZTAsC+Q/1+iRNPba9Pzvq+Xz0Om1W3hbN89Qn83ZcJ6wCeiw4GK3Z8AHDTjwBFBWkQxzZ7de3MqDDyeGRUfXZtLcfqx40StqaMW7SVqBQIDAQAB
-----END PUBLIC KEY-----`)

	// var strArgs = []string{"getAssociatedProviders", path}
	/////

	// fmt.Println("\n\nloadPublicKey ======== ", strArgs)

	// dataInvoke := fetchDataFromOtherContract(stub, strArgs, configurationContract, channelName)

	// fmt.Println("\n\n loadPublicKey RESP >> ", dataInvoke.Message)

	// if dataInvoke.Status != shim.OK {
	// 	return nil, errors.New(dataInvoke.Message)
	// }

	// assocEcomDataObj := ECommOnboarding{}
	// err := json.Unmarshal([]byte(dataInvoke.Payload), &assocEcomDataObj)

	// if err != nil {
	// 	return nil, err
	// }

	// fmt.Println("\n\nloadPublicKey Certificate ======== ", assocEcomDataObj.Certificate)
	return parsePublicKey([]byte(publicKey))
}

func (r *rsaPublicKey) Unsign(message []byte, sig string, sharedSecret []byte) error {
	//h := sha512.New()
	//h.Write(message)
	//d := h.Sum(nil)

	decodedSign, err := hex.DecodeString(sig)
	if err != nil {
		return err
	}

	h := hmac.New(sha512.New, sharedSecret)
	h.Write([]byte(message))

	d := h.Sum(nil)
	fmt.Println("\n\n\nd: ", d)
	fmt.Println("\ndecodedSign: ", decodedSign)

	// fmt.Printf("\n\n\n>>>> d: %x , sig: %x", d, decodedSign)

	return rsa.VerifyPKCS1v15(r.PublicKey, crypto.SHA512, d, decodedSign)
}

func verifySignature(stub *hypConnect, org string, msg string, signature string, sharedSecret string, callerMSP string, customMSP string) error {
	fmt.Println("\n\n\n msg >>> ", msg)
	fmt.Println("\n\n\n signature >>> ", signature)
	fmt.Println("\n\n\n sharedSecret >>> ", sharedSecret)

	fmt.Println("\n\n\n")

	parser, perr := loadPublicKey(stub, org)
	if perr != nil {
		return errors.New("Error Parsing Key")
	}

	var logs AuditLogs

	//timestamp computation
	// now := time.Now()
	// nanos := now.UnixNano()
	// milis := nanos / 1000000

	// logs.Timestamp = milis
	logs.Signature = signature
	err := parser.Unsign([]byte(msg), signature, []byte(sharedSecret))
	transactionId, cipherMessageId := GetTxID(stub)
	fmt.Println("\n\n\n transactionId >> ", transactionId)
	fmt.Println("\n\n\n cipherMessageId >> ", cipherMessageId)
	logs.Key = transactionId
	logs.DocumentName = "signature_audit_logs"
	logs.Json = msg
	logs.OrgCode = org
	logs.CipherMessageId = cipherMessageId
	logs.Status = "SUCCESS"
	if err != nil {
		logs.Status = "FAILED"

		if err := genericInsertData(stub, logs.Key, "_implicit_org_"+callerMSP, logs); err != nil {
			return err
		}
		if err := genericInsertData(stub, logs.Key, "_implicit_org_"+customMSP, logs); err != nil {
			return err
		}
		return err
	}
	if err := genericInsertData(stub, logs.Key, "_implicit_org_"+callerMSP, logs); err != nil {
		return err
	}
	if err := genericInsertData(stub, logs.Key, "_implicit_org_"+customMSP, logs); err != nil {
		return err
	}

	return nil
}

package transactoins

import (
	"errors"
	"fmt"
	"os"
	"starcoin-ns-demo/contract"
	"starcoin-ns-demo/db"
	"starcoin-ns-demo/tools"
	"testing"
	"time"

	"github.com/celestiaorg/smt"
	"github.com/starcoinorg/starcoin-go/client"
)

const (
	DEV_CONTRACT_ADDRESS string = "0x18351d311d32201149a4df2a9fc2db8a"
	ONE_YEAR_MILLS       uint64 = 1000 * 60 * 60 * 24 * 365
)

var SMT_PLACEHOLDER []byte

func init() {
	SMT_PLACEHOLDER = make([]byte, 32)
}

func TestDomainNameInitGenesis(t *testing.T) {
	starcoinClient := localDevStarcoinClient()
	privateKeyConfig, err := localDevAdminPrivateKeyConfig()
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	init_genesis_tx_payload := EncodeEmptyArgsTxPaylaod(DEV_CONTRACT_ADDRESS+"::DomainNameScripts", "init_genesis")
	txHash, err := tools.SubmitStarcoinTransaction(&starcoinClient, privateKeyConfig, &init_genesis_tx_payload)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	ok, err := tools.WaitTransactionConfirm(&starcoinClient, txHash, time.Minute*2)
	fmt.Println(ok, err)
	if !ok {
		t.FailNow()
	}
}

func TestDomainNameRegisterFirstDomain(t *testing.T) {
	starcoinClient := localDevStarcoinClient()
	privateKeyConfig, err := localDevAlicePrivateKeyConfig()
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	tld := "stc"
	sld := "a"
	testSubmitDomainNameRegisterTransaction(&starcoinClient, privateKeyConfig, tld, sld, SMT_PLACEHOLDER, []byte{}, []byte{}, t)
}

func TestDomainNameRegisterDomains(t *testing.T) {
	starcoinClient := localDevStarcoinClient()
	privateKeyConfig, err := localDevAlicePrivateKeyConfig()
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	nodeStore, valueStore, _ := testGetDBDomainNameSmtMapStores()
	smTree := smt.NewSparseMerkleTree(nodeStore, valueStore, db.New256Hasher())

	// testRegisterDomainName(&starcoinClient, smTree, privateKeyConfig, "stc", "b", t)
	// time.Sleep(time.Second * 5)
	// testRegisterDomainName(&starcoinClient, smTree, privateKeyConfig, "stc", "c", t)
	// time.Sleep(time.Second * 5)
	// testRegisterDomainName(&starcoinClient, smTree, privateKeyConfig, "stc", "d", t)
	// time.Sleep(time.Second * 5)
	// testRegisterDomainName(&starcoinClient, smTree, privateKeyConfig, "stc", "e", t)
	// time.Sleep(time.Second * 5)
	// testRegisterDomainName(&starcoinClient, smTree, privateKeyConfig, "stc", "f", t)
	// time.Sleep(time.Second * 5)
	// testRegisterDomainName(&starcoinClient, smTree, privateKeyConfig, "stc", "g", t)
	// time.Sleep(time.Second * 5)

	// testRegisterDomainName(&starcoinClient, smTree, privateKeyConfig, "stc", "h", t)
	// time.Sleep(time.Second * 5)
	// testRegisterDomainName(&starcoinClient, smTree, privateKeyConfig, "stc", "i", t)
	// time.Sleep(time.Second * 5)

}

func testRegisterDomainName(starcoinClient *client.StarcoinClient, smTree *smt.SparseMerkleTree, privateKeyConfig map[string]string, tld string, sld string, t *testing.T) {
	var err error
	var domainNameId *db.DomainNameId
	var smtRoot []byte
	var key []byte
	var proof smt.SparseMerkleProof

	domainNameId = db.NewDomainNameId(tld, sld)
	key, err = domainNameId.BcsSerialize()
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	smtRootStr, err := contract.GetDomainNameSmtRoot(starcoinClient, DEV_CONTRACT_ADDRESS)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	smtRoot = testDecodeSmtRootHex(smtRootStr, t)
	proof, err = smTree.ProveForRoot(key, smtRoot)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	testSubmitDomainNameRegisterTransaction(starcoinClient, privateKeyConfig, domainNameId.TopLevelDomain, domainNameId.SecondLevelDomain, smtRoot, proof.NonMembershipLeafData, concatSideNodes(proof.SideNodes), t)

}

func concatSideNodes(nodes [][]byte) []byte {
	r := make([]byte, 0)
	for _, n := range nodes {
		r = append(r, n...)
	}
	return r
}

func testDecodeSmtRootHex(h string, t *testing.T) []byte {
	bytes, err := tools.HexToBytes(h)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	return bytes
}

func testSubmitDomainNameRegisterTransaction(starcoinClient *client.StarcoinClient, privateKeyConfig map[string]string, tld string, sld string, smtRoot []byte, nonMemberLeaf []byte, sideNodes []byte, t *testing.T) { //(bool, error) {
	register_tx_payload := EncodeDomainNameRegisterTxPaylaod(DEV_CONTRACT_ADDRESS, tld, sld, ONE_YEAR_MILLS, smtRoot, nonMemberLeaf, sideNodes)
	txHash, err := tools.SubmitStarcoinTransaction(starcoinClient, privateKeyConfig, &register_tx_payload)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	ok, err := tools.WaitTransactionConfirm(starcoinClient, txHash, time.Minute*2)
	fmt.Println(ok, err)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	if !ok {
		t.FailNow()
	}
}

func localDevStarcoinClient() client.StarcoinClient {
	starcoinClient := client.NewStarcoinClient("http://localhost:9850")
	return starcoinClient
}

func localDevAdminPrivateKeyConfig() (map[string]string, error) {
	privateKeyConfig := make(map[string]string)
	account, privateKey, err := localDevAdminAccountAddressAndPrivateKey()
	if err != nil {
		return nil, err
	}
	privateKeyConfig[account] = privateKey
	return privateKeyConfig, nil
}

func localDevAlicePrivateKeyConfig() (map[string]string, error) {
	privateKeyConfig := make(map[string]string)
	account, privateKey, err := localDevAliceAccountAddressAndPrivateKey()
	if err != nil {
		return nil, err
	}
	privateKeyConfig[account] = privateKey
	return privateKeyConfig, nil
}

func localDevAdminAccountAddressAndPrivateKey() (string, string, error) {
	account := "0x18351d311d32201149a4df2a9fc2db8a"
	if account == "" {
		return "", "", errors.New("Plz. provide account address.")
	}
	privateKey := os.Getenv("PRIVATE_KEY_18351d3")
	if privateKey == "" {
		return "", "", errors.New("Plz. privide private key.")
	}
	return account, privateKey, nil
}

func localDevAliceAccountAddressAndPrivateKey() (string, string, error) {
	account := "0xb6d69dd935edf7f2054acf12eb884df8" //os.Getenv("0xb6d69dd935edf7f2054acf12eb884df8")
	if account == "" {
		return "", "", errors.New("Plz. provide account address.")
	}
	privateKey := os.Getenv("PRIVATE_KEY_b6d69dd")
	if privateKey == "" {
		return "", "", errors.New("Plz. privide private key.")
	}
	return account, privateKey, nil
}

func testGetDBDomainNameSmtMapStores() (smt.MapStore, smt.MapStore, error) {
	// //////////// New MySQL node MapStore /////////////////
	//nodeStore := smt.NewSimpleMap()
	db, err := localDevDB()
	if err != nil {
		return nil, nil, err
	}
	nodeStore, err := db.NewDomainNameSmtNodeMapStore()
	if err != nil {
		return nil, nil, err
	}
	// //////////// New MySQL value MapStore /////////////////
	//valueStore := smt.NewSimpleMap()
	valueStore := db.NewDomainNameSmtValueMapStore()
	return nodeStore, valueStore, nil
}

func localDevDB() (*db.MySqlDB, error) {
	//CREATE SCHEMA `starcoin_ns_demo` DEFAULT CHARACTER SET utf8mb4 ;
	var dsn string = "root:123456@tcp(127.0.0.1:3306)/starcoin_ns_demo?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := db.NewMySqlDB(dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

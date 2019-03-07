package common

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	sig "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	"sort"
	"time"
)

//NewDeployCodeTransaction return a smart contract deploy transaction instance
func NewDeployCodeTransaction(
	gasPrice, gasLimit uint64,
	code []byte,
	needStorage bool,
	cname, cversion, cauthor, cemail, cdesc string) *types.Transaction {

	deployPayload := &payload.DeployCode{
		Code:        code,
		NeedStorage: needStorage,
		Name:        cname,
		Version:     cversion,
		Author:      cauthor,
		Email:       cemail,
		Description: cdesc,
	}
	tx := &types.Transaction{
		Version:  VERSION_TRANSACTION,
		TxType:   types.Deploy,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  deployPayload,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Sigs:     make([]*types.Sig, 0, 0),
	}
	return tx
}

//NewInvokeTransaction return smart contract invoke transaction
func NewInvokeTransaction(gasPrice, gasLimit uint64, code []byte) *types.Transaction {
	invokePayload := &payload.InvokeCode{
		Code: code,
	}
	tx := &types.Transaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		TxType:   types.Invoke,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     make([]*types.Sig, 0, 0),
	}
	return tx
}

//Sign to a transaction
func SignToTransaction(tx *types.Transaction, signer *account.Account) error {
	tx.Payer = signer.Address
	txHash := tx.Hash()
	sigData, err := SignToData(txHash.ToArray(), signer)
	if err != nil {
		return fmt.Errorf("SignToData error:%s", err)
	}
	sig := &types.Sig{
		PubKeys: []keypair.PublicKey{signer.PublicKey},
		M:       1,
		SigData: [][]byte{sigData},
	}
	tx.Sigs = []*types.Sig{sig}

	return nil
}

func MultiSignToTransaction(tx *types.Transaction, m uint16, pubKeys []keypair.PublicKey, signer *account.Account) error {
	pkSize := len(pubKeys)
	if m == 0 || int(m) > pkSize || pkSize > constants.MULTI_SIG_MAX_PUBKEY_SIZE {
		return fmt.Errorf("both m and number of pub key must larger than 0, and small than %d, and m must smaller than pub key number", constants.MULTI_SIG_MAX_PUBKEY_SIZE)
	}
	validPubKey := false
	for _, pk := range pubKeys {
		if keypair.ComparePublicKey(pk, signer.PublicKey) {
			validPubKey = true
			break
		}
	}
	if !validPubKey {
		return fmt.Errorf("Invalid signer")
	}
	var emptyAddress = common.Address{}
	if tx.Payer == emptyAddress {
		payer, err := types.AddressFromMultiPubKeys(pubKeys, int(m))
		if err != nil {
			return fmt.Errorf("AddressFromMultiPubKeys error:%s", err)
		}
		tx.Payer = payer
	}
	txHash := tx.Hash()
	if len(tx.Sigs) == 0 {
		tx.Sigs = make([]*types.Sig, 0)
	}
	sigData, err := SignToData(txHash.ToArray(), signer)
	if err != nil {
		return fmt.Errorf("SignToData error:%s", err)
	}
	hasMutilSig := false
	for i, sigs := range tx.Sigs {
		if pubKeysEqual(sigs.PubKeys, pubKeys) {
			hasMutilSig = true
			if hasAlreadySig(txHash.ToArray(), signer.PublicKey, sigs.SigData) {
				break
			}
			sigs.SigData = append(sigs.SigData, sigData)
			tx.Sigs[i] = sigs
			break
		}
	}
	if !hasMutilSig {
		tx.Sigs = append(tx.Sigs, &types.Sig{
			PubKeys: pubKeys,
			M:       m,
			SigData: [][]byte{sigData},
		})
	}
	return nil
}

func hasAlreadySig(data []byte, pk keypair.PublicKey, sigDatas [][]byte) bool {
	for _, sigData := range sigDatas {
		err := signature.Verify(pk, data, sigData)
		if err == nil {
			return true
		}
	}
	return false
}

func pubKeysEqual(pks1, pks2 []keypair.PublicKey) bool {
	if len(pks1) != len(pks2) {
		return false
	}
	size := len(pks1)
	if size == 0 {
		return true
	}
	pkstr1 := make([]string, 0, size)
	for _, pk := range pks1 {
		pkstr1 = append(pkstr1, hex.EncodeToString(keypair.SerializePublicKey(pk)))
	}
	pkstr2 := make([]string, 0, size)
	for _, pk := range pks2 {
		pkstr2 = append(pkstr2, hex.EncodeToString(keypair.SerializePublicKey(pk)))
	}
	sort.Strings(pkstr1)
	sort.Strings(pkstr2)
	for i := 0; i < size; i++ {
		if pkstr1[i] != pkstr2[i] {
			return false
		}
	}
	return true
}

//Sign SignToData return the signature to the data of private key
func SignToData(data []byte, signer *account.Account) ([]byte, error) {
	s, err := sig.Sign(signer.SigScheme, signer.PrivateKey, data, nil)
	if err != nil {
		return nil, err
	}
	sigData, err := sig.Serialize(s)
	if err != nil {
		return nil, fmt.Errorf("sig.Serialize error:%s", err)
	}
	return sigData, nil
}

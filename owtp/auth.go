/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package owtp

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/go-owcrypt"
	"github.com/mr-tron/base58/base58"
)

//KeyAgreement 协商密码
type KeyAgreement struct {
	EncryptType            string //协商密码类型
	PublicKeyInitiator     string //发送方：本地公钥
	PrivateKeyInitiator    string //发送方：本地私钥
	TmpPublicKeyInitiator  string //发送方：临时公钥
	TmpPrivateKeyInitiator string //发送方：临时私钥
	PublicKeyResponder     string //响应方：本地公钥
	PrivateKeyResponder    string //响应方：本地私钥
	TmpPublicKeyResponder  string //响应方：临时公钥
	TmpPrivateKeyResponder string //响应方：临时私钥
	S2                     string //响应方：本地验证码，RequestKeyAgreement生成
	SB                     string //响应方：生成协商密码的必要验证码，RequestKeyAgreement生成
	SA                     string //发送方：本地验证码，ResponseKeyAgreement生成
	Key                    string //协商的密钥
}

//Authorization 授权
type Authorization interface {

	//EnableAuth 开启授权
	EnableAuth() bool

	//GenerateSignature 生成签名，并把签名加入到DataPacket中
	GenerateSignature(data *DataPacket) bool
	//VerifySignature 校验签名，若验证错误，可更新错误信息到DataPacket中
	VerifySignature(data *DataPacket) bool

	//EnableKeyAgreement 开启密码协商
	EnableKeyAgreement() bool
	//InitKeyAgreement 发起协商
	InitKeyAgreement(keyAgreement *KeyAgreement) error
	//RequestKeyAgreement 请求协商，计算密钥
	RequestKeyAgreement(keyAgreement *KeyAgreement) error
	//ResponseKeyAgreement 响应协商，计算密钥
	ResponseKeyAgreement(keyAgreement *KeyAgreement) error
	//VerifyKeyAgreement 验证协商结果
	VerifyKeyAgreement(keyAgreement *KeyAgreement) bool

	//EncryptData 加密数据
	EncryptData(data []byte, key []byte) ([]byte, error)
	//DecryptData 解密数据
	DecryptData(data []byte, key []byte) ([]byte, error)
}

type AuthorizationBase struct{}

//EnableAuth 开启授权
func (base *AuthorizationBase) EnableAuth() bool {
	return false
}

//GenerateSignature 生成签名，并把签名加入到DataPacket中
func (base *AuthorizationBase) GenerateSignature(data *DataPacket) bool {
	return true
}

//VerifySignature 校验签名，若验证错误，可更新错误信息到DataPacket中
func (base *AuthorizationBase) VerifySignature(data *DataPacket) bool {
	return true
}

//EnableKeyAgreement 开启密码协商
func (base *AuthorizationBase) EnableKeyAgreement() bool {
	return false
}

//InitKeyAgreement 发起协商
func (base *AuthorizationBase) InitKeyAgreement(keyAgreement *KeyAgreement) error {
	return fmt.Errorf("InitKeyAgreement is not implemented")
}

//RequestKeyAgreement 请求协商，计算密钥
func (base *AuthorizationBase) RequestKeyAgreement(keyAgreement *KeyAgreement) error {
	return fmt.Errorf("RequestKeyAgreement is not implemented")
}

//ResponseKeyAgreement 响应协商，计算密钥
func (base *AuthorizationBase) ResponseKeyAgreement(keyAgreement *KeyAgreement) error {
	return fmt.Errorf("ResponseKeyAgreement is not implemented")
}

//VerifyKeyAgreement 是否完成密码协商，验证协商结果
func (base *AuthorizationBase) VerifyKeyAgreement(keyAgreement *KeyAgreement) bool {
	return false
}

//EncryptData 加密数据
func (base *AuthorizationBase) EncryptData(data []byte, key []byte) ([]byte, error) {
	return nil, fmt.Errorf("EncryptData is not implemented")
}

//DecryptData 解密数据
func (base *AuthorizationBase) DecryptData(data []byte, key []byte) ([]byte, error) {
	return nil, fmt.Errorf("DecryptData is not implemented")
}

type Certificate struct {
	privateKeyBytes []byte
	publicKeyBytes  []byte
	//consultType     string //协商类型
}

//RandomPrivateKey 生成随机私钥
func NewRandomCertificate() Certificate {
	priKey := make([]byte, 32)
	_, err := rand.Read(priKey)
	if err != nil {
		return Certificate{}
	}

	pubkey, ret := owcrypt.GenPubkey(priKey, owcrypt.ECC_CURVE_SM2_STANDARD)
	if ret != owcrypt.SUCCESS {
		return Certificate{}
	}

	cert := Certificate{
		privateKeyBytes: priKey,
		publicKeyBytes:  pubkey,
	}

	return cert
}

func NewCertificate(privateKey string, consultType ...string) (Certificate, error) {

	if len(privateKey) == 0 {
		return Certificate{}, nil
	}

	priKey, err := base58.Decode(privateKey)
	if err != nil {
		return Certificate{}, err
	}

	pubkey, ret := owcrypt.GenPubkey(priKey, owcrypt.ECC_CURVE_SM2_STANDARD)
	if ret != owcrypt.SUCCESS {
		return Certificate{}, err
	}

	//log.Debug("SME PUB:", hex.EncodeToString(pubkey))

	return Certificate{
		privateKeyBytes: priKey,
		publicKeyBytes:  pubkey,
	}, nil
}

func (cert *Certificate) KeyPair() (priv string, pub string) {
	return base58.Encode(cert.privateKeyBytes), base58.Encode(cert.publicKeyBytes)
}

func (cert *Certificate) ID() string {
	nodeID := owcrypt.Hash(cert.publicKeyBytes, 0, owcrypt.HASH_ALG_SHA256)
	return base58.Encode(nodeID)
}

func (cert *Certificate) PublicKeyBytes() []byte {
	return cert.publicKeyBytes
}

func (cert *Certificate) PrivateKeyBytes() []byte {
	return cert.privateKeyBytes
}

//Authorization 授权
type OWTPAuth struct {
	AuthorizationBase

	//本地公钥用于验证数据包合法性
	localPublicKey []byte
	//本地私钥用户签名数据包，生成协商密钥
	localPrivateKey []byte
	//远程节点的公钥
	remotePublicKey []byte
	//是否开启
	enable bool
	//是否协商
	isConsult bool
}

func NewOWTPAuthWithCertificate(cert Certificate, enable bool) (*OWTPAuth, error) {

	//var (
	//	isConsult bool
	//)
	//
	//if len(cert.consultType) > 0 {
	//	isConsult = true
	//} else {
	//	isConsult = false
	//}

	auth := &OWTPAuth{
		localPrivateKey: cert.PrivateKeyBytes(),
		localPublicKey:  cert.PublicKeyBytes(),
		//consultType:     cert.consultType,
		//isConsult:       isConsult,
		enable: enable,
	}

	return auth, nil
}

// RemotePID 远程节点ID
func (auth *OWTPAuth) RemotePID() string {
	//nodeID := crypto.SHA256(auth.remotePublicKey)
	nodeID := owcrypt.Hash(auth.remotePublicKey, 0, owcrypt.HASH_ALG_SHA256)
	return base58.Encode(nodeID)
}

// LocalPID 远程节点ID
func (auth *OWTPAuth) LocalPID() string {
	//log.Debug("input bytes:", auth.localPublicKey)
	//nodeID1 := crypto.SHA256(auth.localPublicKey)
	nodeID := owcrypt.Hash(auth.localPublicKey, 0, owcrypt.HASH_ALG_SHA256)
	//log.Debug("output1 bytes:", nodeID1)
	//log.Debug("output2 bytes:", nodeID)
	return base58.Encode(nodeID)
}

//GenerateSignature 生成签名，并把签名加入到DataPacket中
func (auth *OWTPAuth) GenerateSignature(data *DataPacket) bool {
	if auth.EnableAuth() {
		pub := base58.Encode(auth.localPublicKey)
		//给数据包生成签名
		dataString := common.NewString(data.Data)
		plainText := fmt.Sprintf("%d%s%d%d%s", data.Req, data.Method, data.Nonce, data.Timestamp, dataString)
		hash := owcrypt.Hash([]byte(plainText), 0, owcrypt.HASh_ALG_DOUBLE_SHA256)
		nodeID := owcrypt.Hash(auth.localPublicKey, 0, owcrypt.HASH_ALG_SHA256)
		signature, ret := owcrypt.Signature(auth.localPrivateKey, nodeID, 32, hash, 32, owcrypt.ECC_CURVE_SM2_STANDARD)
		if ret != owcrypt.SUCCESS {
			return false
		}
		data.SecretData.PublicKeyInitiator = pub
		data.Signature = base58.Encode(signature)
		//log.Debug("GenerateSignature packet.Signature: ", data.Signature)
	}
	return true
}

//VerifySignature 校验签名，若验证错误，可更新错误信息到DataPacket中
func (auth *OWTPAuth) VerifySignature(data *DataPacket) bool {
	//验证数据包签名是否合法
	if auth.EnableAuth() {
		var publickey []byte
		if len(data.SecretData.PublicKeyInitiator) > 0 {
			publickey, _ = base58.Decode(data.SecretData.PublicKeyInitiator)
		} else {
			publickey = auth.remotePublicKey
		}

		if len(publickey) == 0 {
			return false
		}

		//log.Debug("VerifySignature packet.Req: ", data.Req)
		//log.Debug("VerifySignature packet.Signature: ", data.Signature)

		dataString := data.Data.(string)
		plainText := fmt.Sprintf("%d%s%d%d%s", data.Req, data.Method, data.Nonce, data.Timestamp, dataString)
		//log.Debug("VerifySignature plainText: ", plainText)
		hash := owcrypt.Hash([]byte(plainText), 0, owcrypt.HASh_ALG_DOUBLE_SHA256)
		//log.Debug("VerifySignature hash: ", hex.EncodeToString(hash))
		nodeID := owcrypt.Hash(publickey, 0, owcrypt.HASH_ALG_SHA256)
		//log.Debug("VerifySignature remotePublicKey: ", hex.EncodeToString(auth.remotePublicKey))
		//log.Debug("VerifySignature nodeID: ", hex.EncodeToString(nodeID))
		signature, err := base58.Decode(data.Signature)
		if err != nil {
			return false
		}
		ret := owcrypt.Verify(publickey, nodeID, 32, hash, 32, signature, owcrypt.ECC_CURVE_SM2_STANDARD)
		if ret != owcrypt.SUCCESS {
			return false
		}
	}
	return true
}

//EnableAuth 开启授权
func (auth *OWTPAuth) EnableAuth() bool {
	return auth.enable
}

//EncryptData 加密数据
func (auth *OWTPAuth) EncryptData(data []byte, key []byte) ([]byte, error) {
	//使用协商密钥加密数据
	if auth.EnableKeyAgreement() && len(key) > 0 && len(data) > 0 {
		encD, err := crypto.AESEncrypt(data, key)
		if err != nil {
			return data, err
		}
		encS := base64.StdEncoding.EncodeToString(encD)
		return []byte(encS), nil
	}

	return data, nil
}

//DecryptData 解密数据
func (auth *OWTPAuth) DecryptData(data []byte, key []byte) ([]byte, error) {
	//使用协商密钥解密数据
	if auth.EnableKeyAgreement() && len(key) > 0 && len(data) > 0 {
		encD, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return data, err
		}
		decD, err := crypto.AESDecrypt(encD, key)
		if err != nil {
			return data, err
		}
		return decD, nil
	}

	return data, nil
}

//AuthHeader 返回授权头
func (auth *OWTPAuth) HTTPAuthHeader() map[string]string {

	if len(auth.localPublicKey) == 0 {
		return nil
	}

	a := base58.Encode(auth.localPublicKey)

	return map[string]string{
		"a": a,
	}
}

//EnableKeyAgreement 开启密码协商
func (auth *OWTPAuth) EnableKeyAgreement() bool {
	return auth.isConsult
}

//RandomPrivateKey 生成随机私钥
func RandomPrivateKey() string {
	buf := make([]byte, 32)
	_, err := rand.Read(buf)
	if err != nil {
		return ""
	}

	key := base58.Encode(buf)
	return key
}

func (auth *OWTPAuth) VerifyKeyAgreement(keyAgreement *KeyAgreement) bool {

	//没有完成密码协商
	if len(keyAgreement.SA) == 0 {
		//请求协商密码参数存在
		if len(keyAgreement.PublicKeyInitiator) > 0 &&
			len(keyAgreement.TmpPublicKeyInitiator) > 0 {
			auth.isConsult = false
			return false
		} else {
			//没有请求协商密码参数
			auth.isConsult = false
			return true
		}
	}

	//auth.isConsult = true

	if len(keyAgreement.S2) == 0 {
		return false
	}

	s2, err := base58.Decode(keyAgreement.S2)
	if err != nil {
		return false
	}

	sa, err := base58.Decode(keyAgreement.SA)
	if err != nil {
		return false
	}

	ret := owcrypt.KeyAgreement_responder_step2(sa, s2, owcrypt.ECC_CURVE_SM2_STANDARD)
	if ret != owcrypt.SUCCESS {
		return false
	}

	if Debug {
		log.Debug("VerifyKeyAgreement passed")
	}

	auth.isConsult = true

	return true
}

//InitKeyAgreement 发起协商
func (auth *OWTPAuth) InitKeyAgreement(keyAgreement *KeyAgreement) error {
	tmpPrikeyInitiator, tmpPubkeyInitiator := owcrypt.KeyAgreement_initiator_step1(owcrypt.ECC_CURVE_SM2_STANDARD)
	//auth.tmpPrivateKey = tmpPrikeyInitiator
	//auth.tmpPublicKey = tmpPubkeyInitiator
	//auth.consultType = keyAgreement.EncryptType

	keyAgreement.TmpPrivateKeyInitiator = base58.Encode(tmpPrikeyInitiator)
	keyAgreement.TmpPublicKeyInitiator = base58.Encode(tmpPubkeyInitiator)
	keyAgreement.PublicKeyInitiator = base58.Encode(auth.localPublicKey)

	auth.isConsult = true

	return nil
}

//RequestKeyAgreement 请求协商
func (auth *OWTPAuth) RequestKeyAgreement(keyAgreement *KeyAgreement) error {

	pubkeyBytes, err := base58.Decode(keyAgreement.PublicKeyInitiator)
	if err != nil {
		return err
	}

	tmpPubkeyBytes, err := base58.Decode(keyAgreement.TmpPublicKeyInitiator)
	if err != nil {
		return err
	}

	localPubkey, err := base58.Decode(keyAgreement.PublicKeyResponder)
	if err != nil {
		return err
	}

	localPrivkey, err := base58.Decode(keyAgreement.PrivateKeyResponder)
	if err != nil {
		return err
	}

	auth.localPublicKey = localPubkey
	auth.localPrivateKey = localPrivkey

	key, tmpPubkeyResponder, s2, sb, err := auth.requestKeyAgreement(pubkeyBytes, tmpPubkeyBytes, localPubkey, localPrivkey)
	if err != nil {
		return err
	}

	//auth.secretKey = key
	//auth.localChecksum = s2
	auth.isConsult = true

	keyAgreement.SB = base58.Encode(sb)
	keyAgreement.Key = base58.Encode(key)
	keyAgreement.S2 = base58.Encode(s2)
	keyAgreement.TmpPublicKeyResponder = base58.Encode(tmpPubkeyResponder)

	//result := map[string]interface{}{
	//	"pubkeyOther":    base58.Encode(auth.localPublicKey),
	//	"tmpPubkeyOther": base58.Encode(tmpPubkeyResponder),
	//	"sb":             base58.Encode(sb),
	//	"secretKey":      base58.Encode(key),
	//	"localChecksum":  base58.Encode(s2),
	//}

	return nil
}

//ResponseKeyAgreement 响应协商
func (auth *OWTPAuth) ResponseKeyAgreement(keyAgreement *KeyAgreement) error {

	remotePublicKeyBytes, err := base58.Decode(keyAgreement.PublicKeyResponder)
	if err != nil {
		return err
	}

	remoteTmpPublicKeyBytes, err := base58.Decode(keyAgreement.TmpPublicKeyResponder)
	if err != nil {
		return err
	}

	sbBytes, err := base58.Decode(keyAgreement.SB)
	if err != nil {
		return err
	}

	tmpPublicKeyBytes, err := base58.Decode(keyAgreement.TmpPublicKeyInitiator)
	if err != nil {
		return err
	}

	tmpPrivateKeyBytes, err := base58.Decode(keyAgreement.TmpPrivateKeyInitiator)
	if err != nil {
		return err
	}

	key, sa, err := auth.responseKeyAgreement(remotePublicKeyBytes, remoteTmpPublicKeyBytes, sbBytes, tmpPublicKeyBytes, tmpPrivateKeyBytes)
	if err != nil {
		return err
	}

	//auth.secretKey = key
	//auth.localChecksum = sa
	auth.isConsult = true

	//result := map[string]interface{}{
	//	"secretKey":     base58.Encode(key),
	//	"localChecksum": base58.Encode(sa),
	//}

	keyAgreement.Key = base58.Encode(key)
	keyAgreement.SA = base58.Encode(sa)

	return nil
}

//requestKeyAgreement 请求协商
func (auth *OWTPAuth) requestKeyAgreement(
	pubkeyInitiatorBytes, tmpPubkeyInitiatorBytes []byte,
	pubkeyResponderBytes, privkeyResponderBytes []byte) ([]byte, []byte, []byte, []byte, error) {

	IDinitiator := owcrypt.Hash(pubkeyInitiatorBytes, 0, owcrypt.HASH_ALG_SHA256)
	IDresponder := owcrypt.Hash(pubkeyResponderBytes, 0, owcrypt.HASH_ALG_SHA256)

	//key, tmpPubkeyResponder, S2, SB, ret := owcrypt.KeyAgreement_responder_step1(
	//	IDinitiator,
	//	32,
	//	IDresponder,
	//	32,
	//	auth.localPrivateKey,
	//	auth.localPublicKey,
	//	pubkeyBytes,
	//	tmpPubkeyBytes,
	//	32,
	//	owcrypt.ECC_CURVE_SM2_STANDARD)

	key, tmpPubkeyResponder, S2, SB, ret := owcrypt.KeyAgreement_responder_ElGamal_step1(
		IDinitiator,
		32,
		IDresponder,
		32,
		privkeyResponderBytes,
		pubkeyResponderBytes,
		pubkeyInitiatorBytes,
		tmpPubkeyInitiatorBytes,
		32,
		privkeyResponderBytes,
		owcrypt.ECC_CURVE_SM2_STANDARD)

	if ret != owcrypt.SUCCESS {
		return nil, nil, nil, nil, fmt.Errorf("KeyAgreement_responder_ElGamal_step1 failed")
	}

	return key, tmpPubkeyResponder, S2, SB, nil
}

//responseKeyAgreement 响应协商
func (auth *OWTPAuth) responseKeyAgreement(pubkeyResponder, tmpPubkeyResponder, sb, tmpPublicKey, tmpPrivateKey []byte) ([]byte, []byte, error) {

	IDinitiator := owcrypt.Hash(auth.localPublicKey, 0, owcrypt.HASH_ALG_SHA256)
	IDresponder := owcrypt.Hash(pubkeyResponder, 0, owcrypt.HASH_ALG_SHA256)

	retA, SA, ret := owcrypt.KeyAgreement_initiator_step2(
		IDinitiator,
		32,
		IDresponder,
		32,
		auth.localPrivateKey,
		auth.localPublicKey,
		pubkeyResponder,
		tmpPrivateKey,
		tmpPublicKey,
		tmpPubkeyResponder,
		sb,
		32,
		owcrypt.ECC_CURVE_SM2_STANDARD)

	if ret != owcrypt.SUCCESS {
		return nil, nil, fmt.Errorf("KeyAgreement_initiator_step2 failed")
	}

	return retA, SA, nil
}

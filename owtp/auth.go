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
	"fmt"
	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/go-owcrypt"
	"github.com/mr-tron/base58/base58"
	"net/http"
	"strconv"
	"time"
)

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
	InitKeyAgreement(consultType string) (map[string]interface{}, error)
	//RequestKeyAgreement 请求协商，计算密钥
	RequestKeyAgreement(params map[string]interface{}) (map[string]interface{}, error)
	//ResponseKeyAgreement 响应协商，计算密钥
	ResponseKeyAgreement(params map[string]interface{}) (map[string]interface{}, error)
	//VerifyKeyAgreement 验证协商结果
	VerifyKeyAgreement(data *DataPacket, checkCode []byte) bool

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
func (base *AuthorizationBase) InitKeyAgreement(consultType string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("InitKeyAgreement is not implemented")
}

//RequestKeyAgreement 请求协商，计算密钥
func (base *AuthorizationBase) RequestKeyAgreement(params map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("RequestKeyAgreement is not implemented")
}

//ResponseKeyAgreement 响应协商，计算密钥
func (base *AuthorizationBase) ResponseKeyAgreement(params map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("ResponseKeyAgreement is not implemented")
}

//VerifyKeyAgreement 是否完成密码协商，验证协商结果
func (base *AuthorizationBase) VerifyKeyAgreement(data *DataPacket, checkCode []byte) bool {
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

func NewCertificate(privateKey string, consultType string) (Certificate, error) {

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
	//本地协商校验值 base58编码
	localChecksum []byte
	//远程节点的公钥
	remotePublicKey []byte
	//远程节点协商校验值 base58编码
	//remoteChecksum []byte
	//用于协商密码的临时公钥
	tmpPublicKey []byte
	//用于协商密码的临时私钥
	tmpPrivateKey []byte
	//协商密钥
	secretKey []byte
	//是否开启
	enable bool
	//是否协商
	isConsult bool
	//协商类型
	consultType string
}

func NewOWTPAuthWithHTTPHeader(header http.Header, def ...Certificate) (*OWTPAuth, error) {

	/*
			| 参数名称 | 类型   | 是否可空   | 描述                                                                              |
		|----------|--------|------------|-----------------------------------------------------------------------------------|
		| a        | string | 是         | 节点公钥，base58                                                                 |
		| p        | string | 是         | 计算协商密码的临时公钥，base58                                                                 |
		| n        | uint32 | 123        | 请求序号。为了保证请求对应响应按序执行，并防御重放攻击，序号可以为随机数，但不可重复。 |
		| t        | uint32 | 1528520843 | 时间戳。限制请求在特定时间范围内有效，如10分钟。                                     |
		| c        | string | 是         | 协商密码，生成的密钥类别，aes-128-ctr，aes-128-cbc，aes-256-ecb等，为空不进行协商      |
		| s        | string | 是         | 组合[a+n+t]并sha256两次，使用钱包工具配置的本地私钥签名，最后base58编码             |
	*/

	var (
		//enable bool
		isConsult       bool
		remotePublicKey []byte
		err             error
	)

	//log.Debug("header:", header)

	a := header.Get("a")
	//p := header.Get("p")
	c := header.Get("c")
	//n := header.Get("n")
	//t := header.Get("t")
	//s := header.Get("s")

	if len(c) > 0 {
		isConsult = true
	} else {
		isConsult = false
	}

	if len(a) == 0 {

		if len(def) > 0 {
			remotePublicKey = def[0].publicKeyBytes
		}
	} else {
		remotePublicKey, err = base58.Decode(a)
		if err != nil {
			return nil, err
		}
	}

	auth := &OWTPAuth{
		remotePublicKey: remotePublicKey,
		consultType:     c,
		isConsult:       isConsult,
	}

	err = auth.VerifyHeader()
	if err != nil {
		return nil, err
	}

	return auth, nil
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
		//给数据包生成签名
		dataString := common.NewString(data.Data)
		plainText := fmt.Sprintf("%d%s%d%d%s", data.Req, data.Method, data.Nonce, data.Timestamp, dataString)
		hash := owcrypt.Hash([]byte(plainText), 0, owcrypt.HASh_ALG_DOUBLE_SHA256)
		nodeID := owcrypt.Hash(auth.localPublicKey, 0, owcrypt.HASH_ALG_SHA256)
		signature, ret := owcrypt.Signature(auth.localPrivateKey, nodeID, 32, hash, 32, owcrypt.ECC_CURVE_SM2_STANDARD)
		if ret != owcrypt.SUCCESS {
			return false
		}
		data.Signature = base58.Encode(signature)
		//log.Debug("GenerateSignature packet.Signature: ", data.Signature)
	}
	return true
}

//VerifySignature 校验签名，若验证错误，可更新错误信息到DataPacket中
func (auth *OWTPAuth) VerifySignature(data *DataPacket) bool {
	//验证数据包签名是否合法
	if auth.EnableAuth() {
		//log.Debug("VerifySignature packet.Req: ", data.Req)
		//log.Debug("VerifySignature packet.Signature: ", data.Signature)
		dataString := data.Data.(string)
		plainText := fmt.Sprintf("%d%s%d%d%s", data.Req, data.Method, data.Nonce, data.Timestamp, dataString)
		hash := owcrypt.Hash([]byte(plainText), 0, owcrypt.HASh_ALG_DOUBLE_SHA256)
		nodeID := owcrypt.Hash(auth.remotePublicKey, 0, owcrypt.HASH_ALG_SHA256)
		signature, err := base58.Decode(data.Signature)
		if err != nil {
			return false
		}
		ret := owcrypt.Verify(auth.remotePublicKey, nodeID, 32, hash, 32, signature, owcrypt.ECC_CURVE_SM2_STANDARD)
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
	//TODO:使用协商密钥加密数据
	if auth.EnableKeyAgreement() && len(key) > 0 && len(data) > 0 {
		encD, err := crypto.AESEncrypt(data, key)
		if err != nil {
			return data, err
		}
		encS := base58.Encode(encD)
		return []byte(encS), nil
	}

	return data, nil
}

//DecryptData 解密数据
func (auth *OWTPAuth) DecryptData(data []byte, key []byte) ([]byte, error) {
	//TODO:使用协商密钥解密数据
	if auth.EnableKeyAgreement() && len(key) > 0 && len(data) > 0 {
		encD, err := base58.Decode(string(data))
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

//VerifyHeader 验证授权头
func (auth *OWTPAuth) VerifyHeader() error {
	//TODO:检查HTTP头，是否进行授权
	return nil
}

//AuthHeader 返回授权头
func (auth *OWTPAuth) WSAuthHeader() map[string]string {

	if len(auth.localPublicKey) == 0 {
		return nil
	}

	a := base58.Encode(auth.localPublicKey)
	//p := base58.Encode(auth.tmpPublicKey)
	n := strconv.FormatInt(time.Now().Unix()+1, 10)
	t := strconv.FormatInt(time.Now().Unix(), 10)
	c := auth.consultType

	//组合[a+p+n+t+c]并sha256两次，使用钱包工具配置的本地私钥签名，最后base58编码
	msg := owcrypt.Hash([]byte(fmt.Sprintf("%s%s%s", a, n, t)), 0, owcrypt.HASh_ALG_DOUBLE_SHA256)

	//计算公钥的ID值
	nodeID := owcrypt.Hash(auth.localPublicKey, 0, owcrypt.HASH_ALG_SHA256)

	//SM2签名
	signature, ret := owcrypt.Signature(auth.localPrivateKey, nodeID, 4, msg, 32, owcrypt.ECC_CURVE_SM2_STANDARD)
	if ret != owcrypt.SUCCESS {
		return nil
	}

	s := base58.Encode(signature)
	return map[string]string{
		"a": a,
		//"p": p,
		"n": n,
		"t": t,
		"c": c,
		"s": s,
	}
}

//AuthHeader 返回授权头
func (auth *OWTPAuth) HTTPAuthHeader() map[string]string {

	if len(auth.localPublicKey) == 0 {
		return nil
	}

	a := base58.Encode(auth.localPublicKey)
	c := auth.consultType

	return map[string]string{
		"a": a,
		"c": c,
	}
}

//EnableKeyAgreement 开启密码协商
func (auth *OWTPAuth) EnableKeyAgreement() bool {
	return auth.isConsult
}

//InitKeyAgreement 发起协商
func (auth *OWTPAuth) InitKeyAgreement(consultType string) (map[string]interface{}, error) {
	tmpPrikeyInitiator, tmpPubkeyInitiator := owcrypt.KeyAgreement_initiator_step1(owcrypt.ECC_CURVE_SM2_STANDARD)
	auth.tmpPrivateKey = tmpPrikeyInitiator
	auth.tmpPublicKey = tmpPubkeyInitiator
	auth.consultType = consultType

	result := map[string]interface{}{
		"pubkey":      base58.Encode(auth.localPublicKey),
		"tmpPubkey":   base58.Encode(tmpPubkeyInitiator),
		"consultType": consultType,
	}

	return result, nil
}

//RequestKeyAgreement 请求协商
func (auth *OWTPAuth) RequestKeyAgreement(params map[string]interface{}) (map[string]interface{}, error) {

	pubkey, ok := params["pubkey"].(string)
	if !ok {
		return nil, fmt.Errorf("request key agreement failed")
	}

	pubkeyBytes, err := base58.Decode(pubkey)
	if err != nil {
		return nil, err
	}

	tmpPubkey := params["tmpPubkey"].(string)
	if !ok {
		return nil, fmt.Errorf("request key agreement failed")
	}

	tmpPubkeyBytes, err := base58.Decode(tmpPubkey)
	if err != nil {
		return nil, err
	}

	localPubkey, ok := params["localPubkey"].([]byte)
	if !ok {
		return nil, fmt.Errorf("request key agreement failed")
	}

	localPrivkey := params["localPrivkey"].([]byte)
	if !ok {
		return nil, fmt.Errorf("request key agreement failed")
	}

	auth.localPublicKey = localPubkey
	auth.localPrivateKey = localPrivkey

	key, tmpPubkeyResponder, s2, sb, err := auth.requestKeyAgreement(pubkeyBytes, tmpPubkeyBytes)
	if err != nil {
		return nil, err
	}

	auth.secretKey = key
	auth.localChecksum = s2
	auth.isConsult = true

	result := map[string]interface{}{
		"pubkeyOther":    base58.Encode(auth.localPublicKey),
		"tmpPubkeyOther": base58.Encode(tmpPubkeyResponder),
		"sb":             base58.Encode(sb),
		"secretKey":      base58.Encode(key),
		"localChecksum":  base58.Encode(s2),
	}

	return result, nil
}

//ResponseKeyAgreement 响应协商
func (auth *OWTPAuth) ResponseKeyAgreement(params map[string]interface{}) (map[string]interface{}, error) {

	remotePublicKey, ok := params["remotePublicKey"].(string)
	if !ok {
		return nil, fmt.Errorf("response key agreement failed")
	}

	remotePublicKeyBytes, err := base58.Decode(remotePublicKey)
	if err != nil {
		return nil, err
	}

	remoteTmpPublicKey := params["remoteTmpPublicKey"].(string)
	if !ok {
		return nil, fmt.Errorf("response key agreement failed")
	}

	remoteTmpPublicKeyBytes, err := base58.Decode(remoteTmpPublicKey)
	if err != nil {
		return nil, err
	}

	sb := params["sb"].(string)
	if !ok {
		return nil, fmt.Errorf("response key agreement failed")
	}

	sbBytes, err := base58.Decode(sb)
	if err != nil {
		return nil, err
	}

	key, sa, err := auth.responseKeyAgreement(remotePublicKeyBytes, remoteTmpPublicKeyBytes, sbBytes)
	if err != nil {
		return nil, err
	}

	auth.secretKey = key
	auth.localChecksum = sa
	auth.isConsult = true

	result := map[string]interface{}{
		"secretKey":     base58.Encode(key),
		"localChecksum": base58.Encode(sa),
	}

	return result, nil
}

//VerifyKeyAgreement 是否完成密码协商，验证协商结果
func (auth *OWTPAuth) VerifyKeyAgreement(data *DataPacket, localChecksum []byte) bool {

	//如果发起协商则不验证
	if data.Method == KeyAgreementMethod {
		auth.isConsult = false
		return true
	}

	//没有密码协商
	if len(data.CheckCode) == 0 {
		auth.isConsult = false
		return true
	}

	auth.isConsult = true

	if len(localChecksum) == 0 {
		return false
	}

	checkCode, err := base58.Decode(data.CheckCode)
	if err != nil {
		return false
	}

	ret := owcrypt.KeyAgreement_responder_step2(checkCode, localChecksum, owcrypt.ECC_CURVE_SM2_STANDARD)
	if ret != owcrypt.SUCCESS {
		return false
	}

	if Debug {
		log.Debug("VerifyKeyAgreement passed")
	}

	auth.isConsult = true

	return true
}

//requestKeyAgreement 请求协商
func (auth *OWTPAuth) requestKeyAgreement(pubkeyBytes, tmpPubkeyBytes []byte) ([]byte, []byte, []byte, []byte, error) {

	IDinitiator := owcrypt.Hash(pubkeyBytes, 0, owcrypt.HASH_ALG_HASH160)
	IDresponder := owcrypt.Hash(auth.localPublicKey, 0, owcrypt.HASH_ALG_HASH160)

	key, tmpPubkeyResponder, S2, SB, ret := owcrypt.KeyAgreement_responder_step1(
		IDinitiator,
		20,
		IDresponder,
		20,
		auth.localPrivateKey,
		auth.localPublicKey,
		pubkeyBytes,
		tmpPubkeyBytes,
		32,
		owcrypt.ECC_CURVE_SM2_STANDARD)

	if ret != owcrypt.SUCCESS {
		return nil, nil, nil, nil, fmt.Errorf("KeyAgreement_responder_step1 failed")
	}

	return key, tmpPubkeyResponder, S2, SB, nil
}

//responseKeyAgreement 响应协商
func (auth *OWTPAuth) responseKeyAgreement(pubkeyResponder, tmpPubkeyResponder, sb []byte) ([]byte, []byte, error) {

	IDinitiator := owcrypt.Hash(auth.localPublicKey, 0, owcrypt.HASH_ALG_HASH160)
	IDresponder := owcrypt.Hash(pubkeyResponder, 0, owcrypt.HASH_ALG_HASH160)

	retA, SA, ret := owcrypt.KeyAgreement_initiator_step2(
		IDinitiator,
		20,
		IDresponder,
		20,
		auth.localPrivateKey,
		auth.localPublicKey,
		pubkeyResponder,
		auth.tmpPrivateKey,
		auth.tmpPublicKey,
		tmpPubkeyResponder,
		sb,
		32,
		owcrypt.ECC_CURVE_SM2_STANDARD)

	if ret != owcrypt.SUCCESS {
		return nil, nil, fmt.Errorf("KeyAgreement_responder_step1 failed")
	}

	return retA, SA, nil
}

//verifySignature 钱包签名
func (auth *OWTPAuth) verifySignature(packet DataPacket) (uint64, string) {
	//TODO:处理验证签名
	return StatusSuccess, ""
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

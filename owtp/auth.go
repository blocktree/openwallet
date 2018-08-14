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
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/go-OWCrypt"
	"github.com/mr-tron/base58/base58"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

var (
	//默认的缓存文件路径
	defaultCachePath = filepath.Join(".", "owptp_cahce")
	defaultCacheFile = "openw.db"
	//重放限制时长，数据包的时间戳超过后，这个间隔，可以重复nonce
	replayLimit = 7 * 24 * time.Hour
)

//Authorization 授权
type Authorization interface {

	//EnableAuth 开启授权
	EnableAuth() bool

	//GenerateSignature 生成签名，并把签名加入到DataPacket中
	GenerateSignature(data *DataPacket) bool
	//VerifySignature 校验签名，若验证错误，可更新错误信息到DataPacket中
	VerifySignature(data *DataPacket) bool

	//InitKeyAgreement 发起协商
	InitKeyAgreement() error
	//RequestKeyAgreement 请求协商，计算密钥
	RequestKeyAgreement(params map[string]interface{}) (map[string]interface{}, error)
	//ResponseKeyAgreement 响应协商，计算密钥
	ResponseKeyAgreement(params map[string]interface{}) (map[string]interface{}, error)
	//VerifyKeyAgreement 验证协商结果
	VerifyKeyAgreement(sa string) error

	//EncryptData 加密数据
	EncryptData(data []byte) ([]byte, error)
	//DecryptData 解密数据
	DecryptData(data []byte) ([]byte, error)
}

type Certificate struct {
	privateKeyBytes []byte
	publicKeyBytes  []byte
	consultType     string //协商类型
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
		consultType:     consultType,
	}, nil
}

func (cert *Certificate) KeyPair() (string, string) {
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
	//本地公钥用于验证数据包合法性
	localPublicKey []byte
	//本地私钥用户签名数据包，生成协商密钥
	localPrivateKey []byte
	//本地协商校验值
	localChecksum string
	//远程节点的公钥
	remotePublicKey []byte
	//用于协商密码的临时公钥
	tmpPublicKey []byte
	//用于协商密码的临时私钥
	tmpPrivateKey []byte
	//协商密钥
	secretKey []byte
	//是否开启
	Enable bool
	//是否协商
	IsConsult bool
	//协商类型
	consultType string
}

func NewOWTPAuthWithHTTPHeader(header http.Header) (*OWTPAuth, error) {

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
		enable bool
	)

	log.Debug("header:", header)

	a := header.Get("a")
	p := header.Get("p")
	c := header.Get("c")
	//n := header.Get("n")
	//t := header.Get("t")
	//s := header.Get("s")

	if len(a) == 0 {
		enable = false
		return &OWTPAuth{
			Enable: enable,
		}, nil
	}

	if len(c) > 0 {
		enable = true
	} else {
		enable = false
	}

	pubkey, err := base58.Decode(a)
	if err != nil {
		return nil, err
	}

	tmpPubkey, err := base58.Decode(p)
	if err != nil {
		return nil, err
	}

	auth := &OWTPAuth{
		remotePublicKey: pubkey,
		tmpPublicKey:    tmpPubkey,
		consultType:     c,
		Enable:          enable,
	}

	err = auth.VerifyHeader()
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func NewOWTPAuthWithCertificate(cert Certificate) (*OWTPAuth, error) {

	var (
		enable bool
	)

	if len(cert.consultType) > 0 {
		enable = true
	} else {
		enable = false
	}

	tmpPrikeyInitiator, tmpPubkeyInitiator := owcrypt.KeyAgreement_initiator_step1(owcrypt.ECC_CURVE_SM2_STANDARD)

	auth := &OWTPAuth{
		localPrivateKey: cert.PrivateKeyBytes(),
		localPublicKey:  cert.PublicKeyBytes(),
		tmpPrivateKey:   tmpPrikeyInitiator,
		tmpPublicKey:    tmpPubkeyInitiator,
		consultType:     cert.consultType,
		Enable:          enable,
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

	//TODO:添加授权后的数据包到缓存数据库

	return true
}

//VerifySignature 校验签名，若验证错误，可更新错误信息到DataPacket中
func (auth *OWTPAuth) VerifySignature(data *DataPacket) bool {

	return true
}

//EnableAuth 开启授权
func (auth *OWTPAuth) EnableAuth() bool {

	return auth.Enable
}

//EncryptData 加密数据
func (auth *OWTPAuth) EncryptData(data []byte) ([]byte, error) {
	return data, nil
}

//DecryptData 解密数据
func (auth *OWTPAuth) DecryptData(data []byte) ([]byte, error) {
	return data, nil
}

//VerifyHeader 验证授权头
func (auth *OWTPAuth) VerifyHeader() error {
	//TODO:检查HTTP头，是否进行授权
	return nil
}

//AuthHeader 返回授权头
func (auth *OWTPAuth) AuthHeader() map[string]string {

	if len(auth.localPublicKey) == 0 {
		return nil
	}

	a := base58.Encode(auth.localPublicKey)
	p := base58.Encode(auth.tmpPublicKey)
	n := strconv.FormatInt(time.Now().Unix()+1, 10)
	t := strconv.FormatInt(time.Now().Unix(), 10)
	c := ""

	//组合[a+p+n+t+c]并sha256两次，使用钱包工具配置的本地私钥签名，最后base58编码
	msg := owcrypt.Hash([]byte(fmt.Sprintf("%s%s%s%s%s", a, p, n, t, c)), 0, owcrypt.HASh_ALG_DOUBLE_SHA256)

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
		"p": p,
		"n": n,
		"t": t,
		"c": "",
		"s": s,
	}
}

//InitKeyAgreement 发起协商
func (auth *OWTPAuth) InitKeyAgreement() error {
	//TODO:初始化协商
	return nil
}

//RequestKeyAgreement 请求协商
func (auth *OWTPAuth) RequestKeyAgreement(params map[string]interface{}) (map[string]interface{}, error) {

	localPrivateKey, ok := params["localPrivateKey"].([]byte)
	if !ok {
		return nil, fmt.Errorf("request key agreement failed")
	}
	localPublicKey := params["localPublicKey"].([]byte)
	if !ok {
		return nil, fmt.Errorf("request key agreement failed")
	}

	key, tmpPubkeyResponder, s2, sb, err := auth.requestKeyAgreement(localPrivateKey, localPublicKey)
	if err != nil {
		return nil, err
	}

	auth.localPrivateKey = localPrivateKey
	auth.localPublicKey = localPublicKey
	auth.secretKey = key
	auth.localChecksum = s2

	result := map[string]interface{}{
		"pubkeyResponder":    localPublicKey,
		"tmpPubkeyResponder": tmpPubkeyResponder,
		"sb":                 sb,
	}

	return result, nil
}

//ResponseKeyAgreement 响应协商
func (auth *OWTPAuth) ResponseKeyAgreement(params map[string]interface{}) (map[string]interface{}, error) {

	remotePublicKey, ok := params["remotePublicKey"].(string)
	if !ok {
		return nil, fmt.Errorf("response key agreement failed")
	}
	remoteTmpPublicKey := params["remoteTmpPublicKey"].(string)
	if !ok {
		return nil, fmt.Errorf("response key agreement failed")
	}
	sb := params["sb"].(string)
	if !ok {
		return nil, fmt.Errorf("response key agreement failed")
	}

	sa, err := auth.responseKeyAgreement(remotePublicKey, remoteTmpPublicKey, sb)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"sa": sa,
	}

	return result, nil
}

//VerifyKeyAgreement 验证协商结果
func (auth *OWTPAuth) VerifyKeyAgreement(sa string) error {
	return nil
}

//RequestKeyAgreement 请求协商
func (auth *OWTPAuth) requestKeyAgreement(prikeyResponder, pubkeyResponder []byte) ([]byte, string, string, string, error) {
	return nil, "", "", "", nil
}

//ResponseKeyAgreement 响应协商
func (auth *OWTPAuth) responseKeyAgreement(pubkeyResponder, tmpPubkeyResponder, sb string) (string, error) {
	return "", nil
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

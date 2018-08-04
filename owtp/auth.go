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
	"fmt"
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/pkg/errors"
	"path/filepath"
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
type OWTPAuth struct {
	//对方节点的公钥
	NodeKey string
	//本地用于授权签名的私钥
	PublickKey string
	//本地用于授权签名的私钥
	AuthKey string
	//使用boltDB引擎的Cache文件
	CacheFile string
	//是否开启
	Enable bool
	//是否协商
	isConsult bool
	//协商类型
	consultType string
}

//NewOWTPAuth 创建OWTP授权
func NewOWTPAuth(nodeKey, publickKey, authKey string, enable bool, cacheFile ...string) (*OWTPAuth, error) {

	var (
		cacheFilePath string
	)
	if len(cacheFile) == 0 {
		//创建数据库路径
		file.MkdirAll(defaultCachePath)
		cacheFilePath = filepath.Join(defaultCachePath, defaultCacheFile)

	} else {
		cacheFilePath = cacheFile[0]
	}

	if len(cacheFilePath) == 0 {
		return nil, errors.New("cache file is not exist.")
	}

	//if !file.Exists(cacheFilePath) {
	//	return nil, errors.New("cache file is not exist.")
	//}

	auth := &OWTPAuth{
		NodeKey:    nodeKey,
		PublickKey: publickKey,
		AuthKey:    authKey,
		Enable:     enable,
		CacheFile:  cacheFilePath,
	}

	return auth, nil
}

//ConnectAuth 连接授权，处理连接授权参数，返回完整的授权连接
func (auth *OWTPAuth) ConnectAuth(url string) string {

	//授权签名过程
	a := auth.PublickKey
	n := time.Now().UnixNano()
	t := time.Now().Unix()
	c := auth.consultType
	s := ""

	url = fmt.Sprintf("%s?a=%s&n=%d&t=%d&s=%s&c=%s", url, a, n, t, s, c)
	return url
}

//GenerateSignature 生成签名，并把签名加入到DataPacket中
func (auth *OWTPAuth) AddAuth(data *DataPacket) bool {

	//TODO:添加授权后的数据包到缓存数据库

	return true
}

//VerifySignature 校验签名，若验证错误，可更新错误信息到DataPacket中
func (auth *OWTPAuth) VerifyAuth(data *DataPacket) bool {

	return true
	var (
		status uint64
		errMsg string
	)

	//检查
	status, errMsg = auth.checkNonceReplay(*data)

	if status != StatusSuccess {
		resp := Response{
			Status: status,
			Msg:    errMsg,
			Result: nil,
		}
		data.Req = WSResponse
		data.Data = resp
		data.Timestamp = time.Now().Unix()
		return false
	}

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

func (auth *OWTPAuth) CreateKeyAgreement(pubKey, tmpPubKey []byte) ([]byte, []byte, error) {
	return nil, nil, nil
}

func (auth *OWTPAuth) HandlerKeyAgreement(ctx *Context) {

}

//verifySignature 钱包签名
func (auth *OWTPAuth) verifySignature(packet DataPacket) (uint64, string) {
	//TODO:处理验证签名
	return StatusSuccess, ""
}

//checkNonceReplay 检查nonce是否重放
func (auth *OWTPAuth) checkNonceReplay(packet DataPacket) (uint64, string) {

	if packet.Nonce == 0 || packet.Timestamp == 0 {
		//没有nonce直接跳过
		return ErrReplayAttack, "no nonce"
	}

	//检查是否重放
	db, err := storm.Open(auth.CacheFile)
	if err != nil {
		return ErrReplayAttack, "client system cache error"
	}
	defer db.Close()

	var existPacket DataPacket
	db.One("Nonce", packet.Nonce, &existPacket)
	if &existPacket != nil {
		return ErrReplayAttack, "this is a replay attack"
	}

	return StatusSuccess, ""

}

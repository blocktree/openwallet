package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
)

//MD5 加密
func GetMD5(str string) (md5str string) {
	data := []byte(str)
	has := md5.Sum(data)
	md5str = fmt.Sprintf("%x", has)
	return
}

//MD5 加密
func MD5(data []byte) []byte {
	hash := md5.New()
	md := hash.Sum(nil)
	return md
}

//SHA1 加密
func SHA1(data []byte) []byte {
	hash := sha1.New()
	hash.Write(data)
	md := hash.Sum(nil)
	return md
}

//SHA256 加密
func SHA256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	md := hash.Sum(nil)
	return md
}

//HmacSHA1 加密
func HmacSHA1(secret string, data []byte) []byte {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write(data)
	md := h.Sum(nil)
	return md
}

//HmacMD5 加密
func HmacMD5(secret string, data []byte) []byte {
	h := hmac.New(md5.New, []byte(secret))
	h.Write(data)
	md := h.Sum(nil)
	return md
}
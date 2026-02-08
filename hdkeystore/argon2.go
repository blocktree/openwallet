package hdkeystore

import "golang.org/x/crypto/argon2"

// 推荐参数（机构级安全）：
const (
	argon2Time    = 12              // 迭代次数（增加计算深度）
	argon2Memory  = 1 * 1024 * 1024 // 1 GB 内存（单位：KB）
	argon2Threads = 4               // 利用多核，但不过度
	argon2KeyLen  = 32              // 256 位密钥
)

// DeriveKeyArgon2id 使用 Argon2id 派生 32 字节密钥（适合 AES-256）
func DeriveKeyArgon2id(password, salt []byte) []byte {
	return deriveKeyArgon2idDefault(password, salt, argon2Time, argon2Memory, argon2KeyLen, argon2Threads)
}

func deriveKeyArgon2idDefault(password, salt []byte, time, memory, keyLen uint32, threads uint8) []byte {
	return argon2.IDKey(password, salt, time, memory, threads, keyLen)
}

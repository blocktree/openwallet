##测试指令
        

测试程序主要目的是为了示范相关接口的具体使用流程，具体正确性已经经过多次验证。
```
        sm2p256v1产生公钥    : go test -test.run sm2_genpubkey
        sm2p256v1加密解密    : go test -test.run sm2_encdec
        sm2p256v1签名验签    : go test -test.run sm2_signverify
        sm2p256v1密钥协商    : go test -test.run sm2_keyagreement
```

## go package详情
___
```
适用于go语言的封装包放在工程目录的“for go”目录下，经测试，可以在该种调用方式下正常的完成公钥生成，签名验签、加解密和密钥协商功能，且结果正确。
```
- 产生公钥:


    func GenPubkey(prikey []byte, pubkey []byte, typeChoose uint32) uint16
```
入参:
        prikey     : 私钥
        typeChoose : 算法类型选择，可选参数如下
                            ECC_CURVE_SECP256K1(0xECC00000)
                            ECC_CURVE_SECP256R1(0xECC00001)
                            ECC_CURVE_PRIMEV1(0xECC00001)
                            ECC_CURVE_NIST_P256(0xECC00001)
                            ECC_CURVE_SM2_STANDARD(0xECC00002)
                            ECC_CURVE_ED25519(0xECC00003)
出参:    
        pubkey    : 公钥
返回值： uint16类型，如下：
                        SUCCESS(0x0001)                 : 生成成功
                        ECC_PRIKEY_ILLEGAL(0xE000)      : 传入了非法私钥
                        ECC_WRONG_TYPE(0xE002)          : 传入了错误的type
```
- 数字签名：

   func Signature(prikey []byte, ID []byte, IDlen uint16, message []byte, message_len uint16, signature []byte, typeChoose uint32) uint16 
```
入参：   
        prikey     ： 私钥
        ID         ： 签名方标识符，仅SM2签名时需要传入
        IDlen      ： 签名方标识符长度
        message    ： 待签名的消息
        message_len： 待签名的消息长度
        typeChoose : 算法类型选择，可选参数如下
                            ECC_CURVE_SECP256K1(0xECC00000)
                            ECC_CURVE_SECP256R1(0xECC00001)
                            ECC_CURVE_PRIMEV1(0xECC00001)
                            ECC_CURVE_NIST_P256(0xECC00001)
                            ECC_CURVE_SM2_STANDARD(0xECC00002)
                            ECC_CURVE_ED25519(0xECC00003)
出参：
        signature  ： 签名值
返回值： 
        uint16类型，如下：
                        SUCCESS(0x0001)                 : 生成成功
                        ECC_PRIKEY_ILLEGAL(0xE000)      : 传入了非法私钥
                        ECC_WRONG_TYPE(0xE002)          : 传入了错误的type
                        ECC_MISS_ID(0xE003)             : SM2签名时未传入签名方标识符

```

- 签名验证：

   func Verify(pubkey []byte, ID []byte, IDlen uint16, message []byte, message_len uint16, signature []byte, typeChoose uint32) uint16 

```

入参：
        pubkey     ： 公钥
        ID         ： 待验证方标识符，仅SM2签名时需要传入
        IDlen      ： 待验证方标识符长度
        message    ： 待验证的消息
        message_len： 待签名的消息长度
        signature  ： 签名值
        typeChoose : 算法类型选择，可选参数如下
                            ECC_CURVE_SECP256K1(0xECC00000)
                            ECC_CURVE_SECP256R1(0xECC00001)
                            ECC_CURVE_PRIMEV1(0xECC00001)
                            ECC_CURVE_NIST_P256(0xECC00001)
                            ECC_CURVE_SM2_STANDARD(0xECC00002)
                            ECC_CURVE_ED25519(0xECC00003)
出参：   无
返回值：
        uint16类型， 如下：
                        SUCCESS(0x0001)                      : 签名验证通过
                        FAILURE(0x0000)                      : 签名验证不通过
                        ECC_PUBKEY_ILLEGAL(0xE001)           : 传入了非法公钥
                        ECC_WRONG_TYPE(0xE002)               : 传入了错误的type
                        ECC_MISS_ID(0xE002)                  : SM2验签时未传入被验证方标识符

```
- 加密：

   func Encryption(pubkey []byte, plain []byte, plain_len uint16, cipher []byte, typeChoose uint32) (ret, cipher_len uint16) 
```
入参：
        pubkey     ：公钥
        plain      ：明文
        plain_len  ：明文长度
        typeChoose : 算法类型选择，可选参数如下
                            ECC_CURVE_SECP256K1(0xECC00000)
                            ECC_CURVE_SECP256R1(0xECC00001)
                            ECC_CURVE_PRIMEV1(0xECC00001)
                            ECC_CURVE_NIST_P256(0xECC00001)
                            ECC_CURVE_SM2_STANDARD(0xECC00002)
                            ECC_CURVE_ED25519(0xECC00003)
出参：  
        cipher     ： 密文
返回值：
        ret： 结果返回码，如下：
                        SUCCESS(0x0001)                      : 加密成功
                        ECC_PUBKEY_ILLEGAL(0xE001)           : 传入了非法公钥
                        ECC_WRONG_TYPE(0xE002)               : 传入了错误的type   
        cipher_len： 密文长度(目前架构下，SM2的密文长度为 plain_len + 97)
        
```
- 解密：

   func Decryption(prikey []byte, cipher []byte, cipher_len uint16, plain []byte, typeChoose uint32) (ret, plain_len uint16) 
```
入参：
        prikey     ： 私钥
        cipher     ： 密文
        cipher_len ： 密文长度
        typeChoose ： 算法类型选择，可选参数如下
                            ECC_CURVE_SECP256K1(0xECC00000)
                            ECC_CURVE_SECP256R1(0xECC00001)
                            ECC_CURVE_PRIMEV1(0xECC00001)
                            ECC_CURVE_NIST_P256(0xECC00001)
                            ECC_CURVE_SM2_STANDARD(0xECC00002)
                            ECC_CURVE_ED25519(0xECC00003)
出参： 
        plain      ： 明文
返回值：
        ret： 结果如下：
                        SUCCESS(0x0001)                      : 解密成功
                        FAILURE(0x0000)                      : 解密失败，密文非法
                        ECC_PRIKEY_ILLEGAL(0xE000)           : 传入了非法私钥
                        ECC_WRONG_TYPE(0xE002)               : 传入了错误的type
        plain_len  ： 明文长度(当前架构下，SM2明文长度为cipher_len - 97)
```
- 协商：

   func KeyAgreement_initiator_step1(tmpPrikeyInitiator []byte, tmpPubkeyInitiator []byte, typeChoose uint32)
```
入参：
        typeChoose              ： 算法类型选择，目前仅支持sm2p256v1
                                       ECC_CURVE_SM2_STANDARD(0xECC00002)
出参： 
        tmpPrikeyInitiator      ： 发起方临时私钥
        tmpPubkeyInitiator      ： 发起方临时公钥
返回值：
        无
```

   func KeyAgreement_initiator_step2(IDinitiator []byte, IDinitiator_len uint16, IDresponder []byte, IDresponder_len uint16, prikeyInitiator []byte,  pubkeyInitiator []byte, pubkeyResponder []byte, tmpPrikeyInitiator []byte, tmpPubkeyInitiator []byte, tmpPubkeyResponder []byte, Sin []byte, Sout []byte,  keylen uint16,  key []byte,  typeChoose uint32) uint16
```
入参：
        IDinitiator             ： 发起方标识符
        IDinitiator_len         ： 发起方标识符长度
        IDresponder             ： 响应方标识符
        IDresponder_len         ： 响应方标识符长度
        prikeyInitiator         ： 发起方私钥
        pubkeyInitiator         ： 发起方公钥
        pubkeyResponder         ： 响应方公钥
        tmpPrikeyInitiator      ： 发起方临时私钥
        tmpPubkeyInitiator      ： 发起方临时公钥
        tmpPubkeyResponder      ： 响应方临时公钥
        Sin                     ： 响应方发来的校验值
        keylen                  ： 期待的协商结果长度
        typeChoose              ： 算法类型选择，目前仅支持sm2p256v1
                                       ECC_CURVE_SM2_STANDARD(0xECC00002)
出参： 
        Sout                    ： 发送给响应方的校验值
        key                     ： 协商结果
返回值：
        uint16类型， 结果如下：
                        SUCCESS(0x0001)                      : 发起方协商成功
                        FAILURE(0x0000)                      : 发起方协商失败
                        ECC_WRONG_TYPE(0xE002)               : 传入了错误的type
```

  func KeyAgreement_responder_step1(IDinitiator []byte, IDinitiator_len uint16, IDresponder []byte, IDresponder_len uint16, prikeyResponder []byte, pubkeyResponder []byte, pubkeyInitiator []byte, tmpPubkeyResponder []byte, tmpPubkeyInitiator []byte, Sinner []byte, Souter []byte, keylen uint16, key []byte, typeChoose uint32) uint16
```
入参：
        IDinitiator             ： 发起方标识符
        IDinitiator_len         ： 发起方标识符长度
        IDresponder             ： 响应方标识符
        IDresponder_len         ： 响应方标识符长度
        prikeyResponder         ： 响应方私钥
        pubkeyResponder         ： 响应方公钥
        pubkeyInitiator         ： 发起方公钥
        tmpPubkeyResponder      ： 响应方临时公钥
        tmpPubkeyInitiator      ： 发起方临时公钥
        keylen                  ： 期待的协商结果长度
        typeChoose              ： 算法类型选择，目前仅支持sm2p256v1
                                       ECC_CURVE_SM2_STANDARD(0xECC00002)
出参： 
        Sinner                  ： 本地暂存的校验值
        Souter                  ： 发送给发起方的校验值
        key                     ： 协商结果
返回值：
        uint16类型， 结果如下：
                        SUCCESS(0x0001)                      : 响应方产生成功
                        FAILURE(0x0000)                      : 响应方协商失败
                        ECC_WRONG_TYPE(0xE002)               : 传入了错误的type
```
  func KeyAgreement_responder_step2(Sinitiator []byte, Sresponder []byte, typeChoose uint32) uint16 
```
入参：
        Sinitiator              ： 发起方发来的校验值
        Sresponder              ： 响应方暂存的校验值
        typeChoose              ： 算法类型选择，目前仅支持sm2p256v1
                                       ECC_CURVE_SM2_STANDARD(0xECC00002)
出参： 
       无
返回值：
        uint16类型， 结果如下：
                        SUCCESS(0x0001)                      : 响应方协商成功
                        FAILURE(0x0000)                      : 响应方协商失败
                        ECC_WRONG_TYPE(0xE002)               : 传入了错误的type
```

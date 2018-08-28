## C程序API详情
___
 - 公钥生成：

    uint16_t ECC_genPubkey(uint8_t *prikey, uint8_t *pubkey, uint32_t type);
    ```
    入参:
            prikey: 私钥 - HEX格式大端传入
            type  : 算法类型选择 - 可选参数如下
                            ECC_CURVE_SECP256K1(0xECC00000)
                            ECC_CURVE_SECP256R1(0xECC00001)
                            ECC_CURVE_PRIMEV1(0xECC00001)
                            ECC_CURVE_NIST_P256(0xECC00001)
                            ECC_CURVE_SM2_STANDARD(0xECC00002)
                            ECC_CURVE_ED25519(0xECC00003)
    出参:
            pubkey: 公钥 - HEX格式大端传出，空间需要预申请
    返回值:
            SUCCESS(0x0001)                 : 生成成功
            ECC_PRIKEY_ILLEGAL(0xE000)      : 传入了非法私钥
            ECC_WRONG_TYPE(0xE002)          : 传入了错误的type
    ```
    
 - 数字签名：

    uint16_t ECC_sign(uint8_t *prikey, uint8_t *ID, uint16_t IDlen, uint8_t *message, uint16_t message_len, uint8_t *sig, uint32_t type)
    ```
    入参:
            prikey     : 私钥 - HEX格式大端传入
            ID         : 签名方标识 - HEX格式大端传入，仅在SM2签名算法中使用
            IDlen      : 签名方标识长度，以字节为单位 - 仅在SM2签名算法中使用
            message    : 需要签名的原始消息 - HEX格式大端传入
            message_len: 源消息长度 - 以字节为单位
            type       : 算法类型选择 - 可选参数如下
                            ECC_CURVE_SECP256K1(0xECC00000)
                            ECC_CURVE_SECP256R1(0xECC00001)
                            ECC_CURVE_PRIMEV1(0xECC00001)
                            ECC_CURVE_NIST_P256(0xECC00001)
                            ECC_CURVE_SM2_STANDARD(0xECC00002)
                            ECC_CURVE_ED25519(0xECC00003)
    出参:
            sig: 签名值 - HEX格式大端传出，空间需要预申请
    返回值:
            SUCCESS(0x0001)                 : 生成成功
            ECC_PRIKEY_ILLEGAL(0xE000)      : 传入了非法私钥
            ECC_WRONG_TYPE(0xE002)          : 传入了错误的type
            ECC_MISS_ID(0xE003)             : SM2签名时未传入签名方标识符
    ```
 - 签名验证：

    uint16_t ECC_verify(uint8_t *pubkey, uint8_t *ID, uint16_t IDlen, uint8_t *message, uint16_t message_len, uint8_t *sig, uint32_t type)
    ```
    入参:
            pubkey     : 公钥 - HEX格式大端，x轴在前，y轴在后一并传入
            ID         : 被验证方标识，HEX格式大端传入 - 仅在SM2签名算法中使用
            IDlen      : 被验证方标识长度，以字节为单位 - 仅在SM2算法中使用
            message    : 源消息 - HEX格式大端传入
            message_len: 源消息长度 - 以字节为单位
            type       : 算法类型选择 - 可选参数如下
                            ECC_CURVE_SECP256K1(0xECC00000)
                            ECC_CURVE_SECP256R1(0xECC00001)
                            ECC_CURVE_PRIMEV1(0xECC00001)
                            ECC_CURVE_NIST_P256(0xECC00001)
                            ECC_CURVE_SM2_STANDARD(0xECC00002)
                            ECC_CURVE_ED25519(0xECC00003)
            sig        : 签名值 - HEX格式大端传入
    返回值:
            SUCCESS(0x0001)                      : 签名验证通过
            FAILURE(0x0000)                      : 签名验证不通过
            ECC_PUBKEY_ILLEGAL(0xE001)           : 传入了非法公钥
            ECC_WRONG_TYPE(0xE002)               : 传入了错误的type
            ECC_MISS_ID(0xE002)                  : SM2验签时未传入被验证方标识符
    ```
 - 加密(目前仅支持sm2p256v1)：

    uint16_t ECC_enc(uint8_t *pubkey, uint8_t *plain, uint16_t plain_len, uint8_t *cipher, uint16_t *cipher_len, uint32_t type)
    ```
    入参:
            pubkey     : 公钥 - HEX格式大端，x轴在前，y轴在后一并传入
            plain      : 明文 - HEX格式大端传入
            plain_len  : 明文长度 - 以字节为单位
            type       : 算法类型选择，目前仅支持SM2p256v1
                                ECC_CURVE_SM2_STANDARD(0xECC00002)
    出参:   
            cipher     : 密文 - HEX格式大端传出，空间需要预申请
            cipher_len : 密文长度 - 以字节为单位，目前架构为plain_len + 97
    返回值:
            SUCCESS(0x0001)                      : 加密成功
            ECC_PUBKEY_ILLEGAL(0xE001)           : 传入了非法公钥
            ECC_WRONG_TYPE(0xE002)               : 传入了错误的type
    ```
 - 解密(目前仅支持sm2p256v1)：

    uint16_t ECC_dec(uint8_t *prikey, uint8_t *cipher, uint16_t cipher_len, uint8_t *plain, uint16_t *plain_len, uint32_t type)
    ```
    入参:
            prikey     : 私钥 - HEX格式大端传入
            cipher     : 密文 - HEX格式大端传入
            cipher_len : 密文长度 - 以字节为单位
            type       : 算法类型选择，目前仅支持SM2p256v1
                                ECC_CURVE_SM2_STANDARD(0xECC00002)
    出参:   
            plain      : 明文 - HEX格式大端传出，空间需要预申请
            plain_len  : 明文长度 - 以字节为单位，目前架构为cipher_len - 97
    返回值:
            SUCCESS(0x0001)                      : 解密成功
            FAILURE(0x0000)                      : 解密失败，密文非法
            ECC_PRIKEY_ILLEGAL(0xE000)           : 传入了非法私钥
            ECC_WRONG_TYPE(0xE002)               : 传入了错误的type
    ```
 - 协商(目前仅支持sm2p256v1)：

    uint16_t ECC_key_exchange_initiator_step1(uint8_t *tmpPriInitiator, uint8_t *tmpPubInitiator, uint32_t type)
    ```
    入参:
            type       : 算法类型选择，目前仅支持SM2p256v1
                                ECC_CURVE_SM2_STANDARD(0xECC00002)
    出参:   
            tmpPriInitiator  : 发起方临时私钥
            tmpPubInitiator  : 发起方临时公钥
    返回值:
            SUCCESS(0x0001)                      : 产生成功
            ECC_WRONG_TYPE(0xE002)               : 传入了错误的type
    ```
    uint16_t ECC_key_exchange_initiator_step2(uint8_t *IDinitiator, uint16_t IDinitiator_len, uint8_t *IDresponder,uint16_t IDresponder_len,uint8_t *priInitiator,uint8_t *pubInitiator, uint8_t *pubResponder,uint8_t *tmpPriInitiator, uint8_t *tmpPubInitiator, uint8_t *tmpPubResponder, uint8_t *Sin, uint8_t *Sout, uint16_t keylen, uint8_t *key, uint32_t type)
    ```
    入参:
            IDinitiator     : 发起方标识符
            IDinitiator_len : 发起方标识符长度
            IDresponder     : 响应方标识符
            IDresponder_len : 响应方标识符长度
            priInitiator    : 发起方私钥
            pubInitiator    : 发起方公钥
            pubResponder    : 响应方公钥
            tmpPriInitiator : 发起方临时公钥
            tmpPubInitiator : 发起方临时公钥
            tmpPubResponder : 响应方临时公钥
            Sin             : 响应方发来的校验值
            keylen          : 期望的协商结果长度
            type            : 算法类型选择，目前仅支持SM2p256v1
                                     ECC_CURVE_SM2_STANDARD(0xECC00002)
    出参:   
            Sout            : 发送给响应方的校验值
            key             : 协商结果
    返回值:
            SUCCESS(0x0001)                      : 发起方协商成功
            FAILURE(0x0000)                      : 发起方协商失败
            ECC_WRONG_TYPE(0xE002)               : 传入了错误的type
    ```
    uint16_t ECC_key_exchange_responder_step1(uint8_t *IDinitiator, uint16_t IDinitiator_len, uint8_t *IDresponder, uint16_t IDresponder_len, uint8_t *priResponder, uint8_t *pubResponder, uint8_t *pubInitiator, uint8_t *tmpPubResponder, uint8_t *tmpPubInitiator, uint8_t *Sin, uint8_t *Sout, uint16_t keylen, uint8_t *key, uint32_t type)
    ```
    入参:
            IDinitiator     : 发起方标识符
            IDinitiator_len : 发起方标识符长度
            IDresponder     : 响应方标识符
            IDresponder_len : 响应方标识符长度
            priResponder    : 响应方私钥
            pubResponder    : 响应方公钥
            pubInitiator    : 发起方公钥
            tmpPubResponder : 响应方临时公钥
            tmpPubInitiator : 发起方临时公钥
            keylen          : 期望的协商结果长度
            type            : 算法类型选择，目前仅支持SM2p256v1
                                     ECC_CURVE_SM2_STANDARD(0xECC00002)
    出参:   
            Sin             : 本地保存的校验值
            Sout            : 发送给发起方的校验值
            key             : 协商结果
    返回值:
            SUCCESS(0x0001)                      : 响应方产生成功
            FAILURE(0x0000)                      : 响应方产生失败
            ECC_WRONG_TYPE(0xE002)               : 传入了错误的type
    ```
    uint16_t ECC_key_exchange_responder_step2(uint8_t *Sinitiator, uint8_t *Sresponder, uint32_t type)
    ```
    入参:
            Sinitiator      : 发起方发来的校验值
            Sresponder      : 响应方本地保存的校验值
            type            : 算法类型选择，目前仅支持SM2p256v1
                                     ECC_CURVE_SM2_STANDARD(0xECC00002)
    出参:   
            无
    返回值:
            SUCCESS(0x0001)                      : 响应方协商成功
            FAILURE(0x0000)                      : 响应方协商失败
            ECC_WRONG_TYPE(0xE002)               : 传入了错误的type
  
    ```
 - 点乘+点加   
    uint16_t ECC_point_mul_add(uint8_t *inputpoint1_buf,uint8_t *inputpoint2_buf,uint8_t *k,uint8_t *outpoint_buf,uint32_t type)
    ```
     入参:
            inputpoint1_buf  : 椭圆曲线上的一个点（采用uint8_t类型存储）
            inputpoint2_buf  : 椭圆曲线上的另一个点（采用uint8_t类型存储）
            k                : 点乘的乘数
            type             : 算法类型选择，目前支持SECP256K1（0xECC00000）、SECP256R1（0xECC00001）和SM2_STANDARD（0xECC00002）
     出参:   outpoint_buf     :  (Point)inputpoint1_buf + [k](Point)inputpoint2_buf的运算结果（先点乘，再点加）
     返回值：1:运算成功；0:运算失败     
    ```
- 点乘（基点）+点加    
    uint16_t ECC_point_mul_baseG_add(uint8_t *inputpoint_buf,uint8_t *k,uint8_t *outpoint_buf,uint32_t type)
    ```
     入参:
            inputpoint_buf  : 椭圆曲线上的一个点（采用uint8_t类型存储）
            k               : 点乘的乘数
            type            :  算法类型选择，目前支持SECP256K1（0xECC00000）、SECP256R1（0xECC00001）和SM2_STANDARD（0xECC00002）
    出参:    outpoint_buf    : (Point)inputpoint1_buf + [k]G(椭圆曲线的基点)的运算结果（先点乘，再点加）
    返回值: 1:运算成功；0:运算失败 
    ```
- 点的压缩     
     uint16_t ECC_point_compress(uint8_t *pubKey,uint16_t pubKey_len,uint8_t *x,uint32_t type)
     ```
    入参:
            pubKey     : 待压缩的公钥（需要添加0x04标头）
            pubKey_len : pubKey的字节长度
            type       : 算法类型选择，目前支持SECP256K1（0xECC00000）、SECP256R1（0xECC00001）和SM2_STANDARD（0xECC00002）
    出参:
            x          : 压缩后的横坐标（第1个字节表示纵坐标y奇偶：0x02表示y为偶数; 0x03表示y为奇数）
    返回值: 1:运算成功；0:运算失败    
    ```
- 点的解压缩    
    uint16_t ECC_point_decompress(uint8_t *x,uint16_t x_len,uint8_t *y,uint32_t type)
    ```
    入参:
            x         : 待解压缩点的横坐标中（第1个字节表示y的奇偶性）
            x_len     : x的字节长度
            type      : 算法类型选择，目前支持SECP256K1（0xECC00000）、SECP256R1（0xECC00001）和SM2_STANDARD（0xECC00002）
    出参:
    返回值:  1:运算成功；0:运算失败
    ```
- 哈希运算
    void hash(uint8_t *msg,uint32_t msg_len,uint8_t *digest,uint16_t digest_len,uint32_t Type)
    ```
    入参：
            msg        : 哈希运算的消息
            msg_len    : 消息的字节长度
            digest_len : 摘要长度
            type       : 哈希算法类型选择，目前支持sha1(0xA0000000)、sha256(0xA0000001)、sha512(0xA0000003)、MD4(0xA0000004)、
                         MD5(0xA0000005)、RIPEMD160（0xA0000006）、BLAKE2B（0xA0000007）、BLAKE2S（0xA0000008、SM3（0xA0000009）
    ```    
## 测试结果
___
```
c语言版本在xcode下进行了正确性和抗压力测试；
经测试，产生公钥、加密解密、签名验签和密钥协商结果均正确；
各个功能经受了半个小时左右持续工作的压力测试，测试过程中未出现内存冲突或者程序跑飞的情况，且输出数据均正确无误。
```

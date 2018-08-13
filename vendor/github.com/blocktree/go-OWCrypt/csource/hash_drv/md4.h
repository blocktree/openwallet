//
//  md4.h
//  ECC_set
//
//  Created by zhang.zhenshan on 2018/8/2.
//  Copyright © 2018年 何述超. All rights reserved.
//

#ifndef md4_h
#define md4_h
#include "string.h"
#include <stdio.h>
#include "type.h"

/* MD4 context. */
typedef struct {
    uint32_t state[4];                /* state (ABCD) */
    uint32_t count[2];                /* number of bits, modulo 2^64 (lsb first) */
    unsigned char buffer[64];             /* input buffer */
} MD4_CTX;

/*
 @function:MD4 initialization.Begins an MD4 operation, writing a new context
 @paramter[in]:ctx pointer to MD4_CTX
 @return: NULL
 @notoce: none
 */
void md4_init (MD4_CTX *ctx);

/*
 @function:MD4 block update operation. Continues an MD4 message-digest operation, processing another message block, and updating the context.
 @paramter[in]:ctx pointer to MD4_CTX
 @paramter[in]:msg pointer to the message to do MD4
 @paramter[in]:msg_len,the byte length of input
 @return:NULL
 @notoce:none
 */
void md4_update (MD4_CTX *ctx, const uint8_t *msg, uint32_t msg_len);

/*
 @function:MD4 finalization. Ends an MD4 message-digest operation, writing the the message digest and zeroizing the context
 @paramter[out]:digest pointer to MD4 hash result
 @paramter[in]:ctx pointer to MD4_CTX
 @return:NULL
 @notice:nothing
 */
void md4_final (MD4_CTX *ctx,uint8_t digest[16]);

/*
 @function:MD4 hash
 @paramter[in]:msg pointer to the message to do MD4
 @paramter[in]:msg_len,the byte length of input
 @digest[out]:digest pointer to MD4 hash result
 @return:NULL
 @notice:none
 */
void  md4_hash(const uint8_t *msg,uint32_t msg_len,uint8_t digest[16]);
#endif /* md4_h */

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

#ifndef sha512_h
#define sha512_h

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include "type.h"
#define SHA512_DIGEST_LENGTH       64

typedef struct
{
    uint64_t total[2]; /* number of bits, modulo 2^128 */
    uint64_t state[8]; /*intermediate state*/
    uint8_t buffer[128];/*input buffer*/
}SHA512_CTX;

/*
 @function:init SHA512_CTX,writing a new message
 @paramter[in]:ctx pointer to SHA512_CTX
 @return: NULL
 @notice: none
 */
void sha512_init(SHA512_CTX *ctx);

/*
 @function:update message Continues an sha512 message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:ctx pointer to SHA512_CTX
 @paramter[in]:msg pointer to the message to do sha512
 @paramter[in]:msg_len,the byte length of input
 @return:NULL
 @notoce:none
 */
void sha512_update(SHA512_CTX *ctx, uint8_t *msg, uint16_t msg_len);

/*
 @function:finalization sha512 operation ends an sha1 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to SHA512_CTX
 @paramter[out]:digest pointer to sha512 hash result
 @return:NULL
 @notice:nothing
 */
void sha512_final(SHA512_CTX *ctx, uint8_t *digest);

/*
 @function: sha512 hash
 @parameter[in]:msg pointer to the message to do hash
 @parameter[in]:msg_len,the byte length of input
 @parameter[in]:digest pointer to hash result
 @return: none
 @notice:nothing
 */
void sha512_hash(uint8_t *msg, uint32_t msg_len, uint8_t *digest);




#endif /* sha512_h */

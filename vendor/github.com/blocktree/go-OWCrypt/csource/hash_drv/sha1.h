
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

#ifndef SHA1_H
#define SHA1_H
#include <stdio.h>
#include <string.h>

typedef struct
{
    uint32_t state[5];
    uint32_t count[2];
    unsigned char buffer[64];
} SHA1_CTX;

/*
 @function:init SHA1_CTX,writing a new message
 @paramter[in]:ctx pointer to SHA1_CTX
 @return: NULL
 @notice: none
 */
void sha1_init(SHA1_CTX * ctx);

/*
 @function:update message Continues an sha1 message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:ctx pointer to SHA1_CTX
 @paramter[in]:msg pointer to the message to do sha1
 @paramter[in]:msg_len,the byte length of input
 @return:NULL
 @notoce:none
 */

void sha1_update(SHA1_CTX * ctx,const uint8_t *msg,uint32_t msg_len);

/*
 @function:finalization sha1 operation ends an sha1 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to SHA1_CTX
 @paramter[out]:digest pointer to sha1 hash result
 @return:NULL
 @notice:nothing
 */
void sha1_final(SHA1_CTX * ctx,uint8_t digest[20]);

void sha1_hash(const uint8_t *msg,uint32_t msg_len,uint8_t *digest);

#endif /* SHA1_H */

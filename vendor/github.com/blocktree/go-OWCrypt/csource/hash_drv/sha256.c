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

#include "sha256.h"


static const uint32_t K_SHA256[] =
{
    0x428A2F98, 0x71374491, 0xB5C0FBCF, 0xE9B5DBA5,
    0x3956C25B, 0x59F111F1, 0x923F82A4, 0xAB1C5ED5,
    0xD807AA98, 0x12835B01, 0x243185BE, 0x550C7DC3,
    0x72BE5D74, 0x80DEB1FE, 0x9BDC06A7, 0xC19BF174,
    0xE49B69C1, 0xEFBE4786, 0x0FC19DC6, 0x240CA1CC,
    0x2DE92C6F, 0x4A7484AA, 0x5CB0A9DC, 0x76F988DA,
    0x983E5152, 0xA831C66D, 0xB00327C8, 0xBF597FC7,
    0xC6E00BF3, 0xD5A79147, 0x06CA6351, 0x14292967,
    0x27B70A85, 0x2E1B2138, 0x4D2C6DFC, 0x53380D13,
    0x650A7354, 0x766A0ABB, 0x81C2C92E, 0x92722C85,
    0xA2BFE8A1, 0xA81A664B, 0xC24B8B70, 0xC76C51A3,
    0xD192E819, 0xD6990624, 0xF40E3585, 0x106AA070,
    0x19A4C116, 0x1E376C08, 0x2748774C, 0x34B0BCB5,
    0x391C0CB3, 0x4ED8AA4A, 0x5B9CCA4F, 0x682E6FF3,
    0x748F82EE, 0x78A5636F, 0x84C87814, 0x8CC70208,
    0x90BEFFFA, 0xA4506CEB, 0xBEF9A3F7, 0xC67178F2,
};

static const unsigned char sha256_padding[64] =
{
    0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
};

#define  SHR_SHA256(x,n) ((x & 0xFFFFFFFF) >> n)
#define ROTR_SHA256(x,n) (SHR_SHA256(x,n) | (x << (32 - n)))

#define S0_SHA256(x) (ROTR_SHA256(x, 7) ^ ROTR_SHA256(x,18) ^  SHR_SHA256(x, 3))
#define S1_SHA256(x) (ROTR_SHA256(x,17) ^ ROTR_SHA256(x,19) ^  SHR_SHA256(x,10))

#define S2_SHA256(x) (ROTR_SHA256(x, 2) ^ ROTR_SHA256(x,13) ^ ROTR_SHA256(x,22))
#define S3_SHA256(x) (ROTR_SHA256(x, 6) ^ ROTR_SHA256(x,11) ^ ROTR_SHA256(x,25))

#define F0_SHA256(x,y,z) ((x & y) | (z & (x | y)))
#define F1_SHA256(x,y,z) (z ^ (x & (y ^ z)))

#define R_SHA256(t)                                    \
(                                               \
    W[t] = S1_SHA256(W[t -  2]) + W[t -  7] +          \
    S0_SHA256(W[t - 15]) + W[t - 16]                   \
)

#define P_SHA256(a,b,c,d,e,f,g,h,x,K)                  \
{                                               \
    temp1 = h + S3_SHA256(e) + F1_SHA256(e,f,g) + K + x;      \
    temp2 = S2_SHA256(a) + F0_SHA256(a,b,c);                  \
    d += temp1; h = temp1 + temp2;              \
}

#ifndef GET_UINT32_BE
#define GET_UINT32_BE(n,b,i)                     \
do {                                             \
    (n) = ( (uint32_t) (b)[(i)    ] << 24 )      \
    | ( (uint32_t) (b)[(i) + 1] << 16 )          \
    | ( (uint32_t) (b)[(i) + 2] <<  8 )          \
    | ( (uint32_t) (b)[(i) + 3]       );         \
} while( 0 )
#endif


#ifndef PUT_UINT32_BE
#define PUT_UINT32_BE(n,b,i)                            \
do {                                                    \
    (b)[(i)    ] = (unsigned char) ( (n) >> 24 );       \
    (b)[(i) + 1] = (unsigned char) ( (n) >> 16 );       \
    (b)[(i) + 2] = (unsigned char) ( (n) >>  8 );       \
    (b)[(i) + 3] = (unsigned char) ( (n)       );       \
} while( 0 )
#endif


/*
 @function:init SHA256_CTX,writing a new message
 @paramter[in]:ctx pointer to SHA256_CTX
 @return: NULL
 @notice: none
 */
void sha256_init(SHA256_CTX *ctx)
{
    memset((uint8_t *)ctx, 0, sizeof(SHA256_CTX));
    
    ctx->state[0] = 0x6A09E667;
    ctx->state[1] = 0xBB67AE85;
    ctx->state[2] = 0x3C6EF372;
    ctx->state[3] = 0xA54FF53A;
    ctx->state[4] = 0x510E527F;
    ctx->state[5] = 0x9B05688C;
    ctx->state[6] = 0x1F83D9AB;
    ctx->state[7] = 0x5BE0CD19;
}

void sha256_process(SHA256_CTX *ctx, uint8_t data[64])
{
    uint32_t temp1, temp2, W[64];
    uint32_t A[8];
    unsigned int i;
    
    for( i = 0; i < 8; i++ )
        A[i] = ctx->state[i];
    
    for( i = 0; i < 16; i++ )
        GET_UINT32_BE( W[i], data, 4 * i );
    
    for( i = 0; i < 16; i += 8 )
    {
        P_SHA256( A[0], A[1], A[2], A[3], A[4], A[5], A[6], A[7], W[i+0], K_SHA256[i+0] );
        P_SHA256( A[7], A[0], A[1], A[2], A[3], A[4], A[5], A[6], W[i+1], K_SHA256[i+1] );
        P_SHA256( A[6], A[7], A[0], A[1], A[2], A[3], A[4], A[5], W[i+2], K_SHA256[i+2] );
        P_SHA256( A[5], A[6], A[7], A[0], A[1], A[2], A[3], A[4], W[i+3], K_SHA256[i+3] );
        P_SHA256( A[4], A[5], A[6], A[7], A[0], A[1], A[2], A[3], W[i+4], K_SHA256[i+4] );
        P_SHA256( A[3], A[4], A[5], A[6], A[7], A[0], A[1], A[2], W[i+5], K_SHA256[i+5] );
        P_SHA256( A[2], A[3], A[4], A[5], A[6], A[7], A[0], A[1], W[i+6], K_SHA256[i+6] );
        P_SHA256( A[1], A[2], A[3], A[4], A[5], A[6], A[7], A[0], W[i+7], K_SHA256[i+7] );
    }
    
    for( i = 16; i < 64; i += 8 )
    {
        P_SHA256( A[0], A[1], A[2], A[3], A[4], A[5], A[6], A[7], R_SHA256(i+0), K_SHA256[i+0] );
        P_SHA256( A[7], A[0], A[1], A[2], A[3], A[4], A[5], A[6], R_SHA256(i+1), K_SHA256[i+1] );
        P_SHA256( A[6], A[7], A[0], A[1], A[2], A[3], A[4], A[5], R_SHA256(i+2), K_SHA256[i+2] );
        P_SHA256( A[5], A[6], A[7], A[0], A[1], A[2], A[3], A[4], R_SHA256(i+3), K_SHA256[i+3] );
        P_SHA256( A[4], A[5], A[6], A[7], A[0], A[1], A[2], A[3], R_SHA256(i+4), K_SHA256[i+4] );
        P_SHA256( A[3], A[4], A[5], A[6], A[7], A[0], A[1], A[2], R_SHA256(i+5), K_SHA256[i+5] );
        P_SHA256( A[2], A[3], A[4], A[5], A[6], A[7], A[0], A[1], R_SHA256(i+6), K_SHA256[i+6] );
        P_SHA256( A[1], A[2], A[3], A[4], A[5], A[6], A[7], A[0], R_SHA256(i+7), K_SHA256[i+7] );
    }
    for( i = 0; i < 8; i++ )
        ctx->state[i] += A[i];
}

/*
 @function:update message Continues an sha256 message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:ctx pointer to SHA256_CTX
 @paramter[in]:msg pointer to the message to do sha256
 @paramter[in]:msg_len,the byte length of input
 @return:NULL
 @notoce:none
 */
void sha256_update(SHA256_CTX *ctx, uint8_t *msg, uint16_t msg_len)
{
    size_t fill;
    uint32_t left;
    
    if(msg_len == 0)
        return;
    
    left = ctx->total[0] & 0x3F;
    fill = 64 - left;
    
    ctx -> total[0] += (uint32_t)msg_len;
    ctx -> total[0] &= 0xFFFFFFFF;
    
    if(ctx->total[0] < (uint32_t)msg_len)
        ctx->total[1]++;
    
    if(left && msg_len >= fill)
    {
        memcpy((uint8_t *)(ctx -> buffer + left), msg, fill);
        sha256_process(ctx, ctx->buffer);
        msg += fill;
        msg_len  -= fill;
        left = 0;
    }
    while(msg_len >= 64)
    {
        sha256_process(ctx, msg);
        msg += 64;
        msg_len  -= 64;
    }
    
    if(msg_len > 0)
        memcpy((uint8_t *)(ctx -> buffer + left), msg, msg_len);
}

/*
 @function:finalization sha256 operation ends an sha1 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to SHA256_CTX
 @paramter[out]:digest pointer to sha256 hash result
 @return:NULL
 @notice:nothing
 */
void sha256_final(SHA256_CTX *ctx, uint8_t *digest)
{
    uint32_t last, padn;
    uint32_t high, low;
    uint8_t msglen[8];
    high = (ctx -> total[0] >> 29)
         | (ctx -> total[1] <<  3);
    low  = (ctx -> total[0] <<  3);
    PUT_UINT32_BE(high, msglen, 0);
    PUT_UINT32_BE(low,  msglen, 4);
    last = ctx -> total[0] & 0x3F;
    padn = (last < 56) ? (56 - last) : (120 - last);
    sha256_update(ctx, (uint8_t *)sha256_padding, padn);
    sha256_update(ctx, msglen, 8);
    PUT_UINT32_BE(ctx -> state[0], digest,  0);
    PUT_UINT32_BE(ctx -> state[1], digest,  4);
    PUT_UINT32_BE(ctx -> state[2], digest,  8);
    PUT_UINT32_BE(ctx -> state[3], digest, 12);
    PUT_UINT32_BE(ctx -> state[4], digest, 16);
    PUT_UINT32_BE(ctx -> state[5], digest, 20);
    PUT_UINT32_BE(ctx -> state[6], digest, 24);
    PUT_UINT32_BE(ctx -> state[7], digest, 28);
}

/*
 @function: sha256 hash
 @parameter[in]:msg pointer to the message to do hash
 @parameter[in]:msg_len,the byte length of msg
 @parameter[in]:digest pointer to hash result
 @return: none
 @notice:nothing
 */
void sha256_hash(uint8_t *msg, uint16_t msg_len, uint8_t *digest)
{
    SHA256_CTX *ctx = NULL;
    ctx = calloc(1, sizeof(SHA256_CTX));
    sha256_init(ctx);
    sha256_update(ctx, msg, msg_len);
    sha256_final(ctx, digest);
    free(ctx);
}


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

#include "blake2b.h"

static void blake2b_compress(BLAKE2B_CTX *S,const uint8_t *blocks,size_t len);

static const uint64_t blake2b_IV[8] =
{
    0x6a09e667f3bcc908U,
    0xbb67ae8584caa73bU,
    0x3c6ef372fe94f82bU,
    0xa54ff53a5f1d36f1U,
    0x510e527fade682d1U,
    0x9b05688c2b3e6c1fU,
    0x1f83d9abfb41bd6bU,
    0x5be0cd19137e2179U
};

static const uint8_t blake2b_sigma[12][16] =
{
    {  0,  1,  2,  3,  4,  5,  6,  7,  8,  9, 10, 11, 12, 13, 14, 15 } ,
    { 14, 10,  4,  8,  9, 15, 13,  6,  1, 12,  0,  2, 11,  7,  5,  3 } ,
    { 11,  8, 12,  0,  5,  2, 15, 13, 10, 14,  3,  6,  7,  1,  9,  4 } ,
    {  7,  9,  3,  1, 13, 12, 11, 14,  2,  6,  5, 10,  4,  0, 15,  8 } ,
    {  9,  0,  5,  7,  2,  4, 10, 15, 14,  1, 11, 12,  6,  8,  3, 13 } ,
    {  2, 12,  6, 10,  0, 11,  8,  3,  4, 13,  7,  5, 15, 14,  1,  9 } ,
    { 12,  5,  1, 15, 14, 13,  4, 10,  0,  7,  6,  3,  9,  2,  8, 11 } ,
    { 13, 11,  7, 14, 12,  1,  3,  9,  5,  0, 15,  4,  8,  6,  2, 10 } ,
    {  6, 15, 14,  9, 11,  3,  0,  8, 12,  2, 13,  7,  1,  4, 10,  5 } ,
    { 10,  2,  8,  4,  7,  6,  1,  5, 15, 11,  9, 14,  3, 12, 13 , 0 } ,
    {  0,  1,  2,  3,  4,  5,  6,  7,  8,  9, 10, 11, 12, 13, 14, 15 } ,
    { 14, 10,  4,  8,  9, 15, 13,  6,  1, 12,  0,  2, 11,  7,  5,  3 }
};

static  void blake2b_set_lastblock(BLAKE2B_CTX *S)
{
    S->f[0] = -1;
}

static void blake2b_init0(BLAKE2B_CTX *S)
{
    int i;
    
    memset(S, 0, sizeof(BLAKE2B_CTX));
    for (i = 0; i < 8; ++i) {
        S->h[i] = blake2b_IV[i];
    }
}

static uint64_t rotr64(const uint64_t w, const unsigned int c)
{
    return (w >> c) | (w << (64 - c));
}


static void blake2b_init_param(BLAKE2B_CTX *S, const BLAKE2B_PARAM *P)
{
    size_t i;
    uint64_t tmp;
    const uint8_t *p = (const uint8_t *)(P);
    blake2b_init0(S);
    
    assert(sizeof(BLAKE2B_PARAM) == 64);
    for (i = 0; i < 8; ++i) {
        memcpy((uint8_t *)&tmp, (uint8_t *)(p + sizeof(S->h[i]) * i), 8);
        S->h[i] ^= tmp;
    }
}

/*
 @function:init BLAKE2B_CTX,writing a new message
 @paramter[in]:ctx pointer to BLAKE2B_CTX structure
 @paramter[in]:key pointer to the key(if dosen't need key,please input NULL)
 @paramter[in]:key_len denotes the byte length of key.(if dosen't need key,please set key_bytelen to zero)
 @paramter[in]:digest_len denotes the expected hash result length
 */
void blake2b_init(BLAKE2B_CTX *ctx, uint8_t *key,uint8_t key_len,uint8_t digest_len)
{
    BLAKE2B_PARAM P[1];
    P->digest_length = digest_len;;
    P->key_length    = key_len;
    P->fanout        = 1;
    P->depth         = 1;
    memset(P->leaf_length, 0, 4);
    memset(P->node_offset, 0, 8);
    P->node_depth    = 0;
    P->inner_length  = 0;
    memset(P->reserved, 0, sizeof(P->reserved));
    memset(P->salt,     0, sizeof(P->salt));
    memset(P->personal, 0, sizeof(P->personal));
    blake2b_init_param(ctx, P);
    if(key && key_len)
    {
        memset(ctx->buf,0,BLAKE2B_BLOCKBYTES);
        memcpy(ctx->buf,key,key_len);
        blake2b_compress(ctx,ctx->buf,BLAKE2B_BLOCKBYTES);
    }
    //return 1;
}

static void blake2b_compress(BLAKE2B_CTX *S,const uint8_t *blocks,size_t len)
{
    uint64_t m[16];
    uint64_t v[16];
    int i;
    size_t increment;
    
    assert(len < BLAKE2B_BLOCKBYTES || len % BLAKE2B_BLOCKBYTES == 0);
    
    increment = len < BLAKE2B_BLOCKBYTES ? len : BLAKE2B_BLOCKBYTES;
    
    for (i = 0; i < 8; ++i)
    {
        v[i] = S->h[i];
    }
    do {
        for (i = 0; i < 16; ++i)
        {
            memcpy((uint8_t *)&m[i], (uint8_t *)(blocks + i * sizeof(m[i])), 8);
        }
        S->t[0] += increment;
        S->t[1] += (S->t[0] < increment);
        v[8]  = blake2b_IV[0];
        v[9]  = blake2b_IV[1];
        v[10] = blake2b_IV[2];
        v[11] = blake2b_IV[3];
        v[12] = S->t[0] ^ blake2b_IV[4];
        v[13] = S->t[1] ^ blake2b_IV[5];
        v[14] = S->f[0] ^ blake2b_IV[6];
        v[15] = S->f[1] ^ blake2b_IV[7];
#define G(r,i,a,b,c,d) \
do { \
a = a + b + m[blake2b_sigma[r][2*i+0]]; \
d = rotr64(d ^ a, 32); \
c = c + d; \
b = rotr64(b ^ c, 24); \
a = a + b + m[blake2b_sigma[r][2*i+1]]; \
d = rotr64(d ^ a, 16); \
c = c + d; \
b = rotr64(b ^ c, 63); \
} while (0)
#define ROUND(r)  \
do { \
G(r,0,v[ 0],v[ 4],v[ 8],v[12]); \
G(r,1,v[ 1],v[ 5],v[ 9],v[13]); \
G(r,2,v[ 2],v[ 6],v[10],v[14]); \
G(r,3,v[ 3],v[ 7],v[11],v[15]); \
G(r,4,v[ 0],v[ 5],v[10],v[15]); \
G(r,5,v[ 1],v[ 6],v[11],v[12]); \
G(r,6,v[ 2],v[ 7],v[ 8],v[13]); \
G(r,7,v[ 3],v[ 4],v[ 9],v[14]); \
} while (0)
        
        ROUND(0);
        ROUND(1);
        ROUND(2);
        ROUND(3);
        ROUND(4);
        ROUND(5);
        ROUND(6);
        ROUND(7);
        ROUND(8);
        ROUND(9);
        ROUND(10);
        ROUND(11);
        
        for (i = 0; i < 8; ++i) {
            S->h[i] = v[i] ^= v[i + 8] ^ S->h[i];
        }
#undef G
#undef ROUND
        blocks += increment;
        len -= increment;
    } while (len);
}
/*
 @function:update message Continues an blake2b message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:ctx pointer to BLAKE2B_CTX structure
 @paramter[in]:data pointer to the message to do hash
 @paramter[in]:datalen denotes the byte length of data.
 */
void blake2b_update(BLAKE2B_CTX *ctx, uint8_t *msg, uint32_t msg_len)
{
    const uint8_t *in = msg;
    size_t fill;
    
    fill = sizeof(ctx->buf) - ctx->buflen;
    if (msg_len > fill)
    {
        if (ctx->buflen)
        {
            memcpy(ctx->buf + ctx->buflen, in, fill);
            blake2b_compress(ctx, ctx->buf, BLAKE2B_BLOCKBYTES);
            ctx->buflen = 0;
            in += fill;
            msg_len -= fill;
        }
        if (msg_len > BLAKE2B_BLOCKBYTES)
        {
            uint32_t stashlen = msg_len % BLAKE2B_BLOCKBYTES;
            stashlen = stashlen ? stashlen : BLAKE2B_BLOCKBYTES;
            msg_len -= stashlen;
            blake2b_compress(ctx, in, msg_len);
            in += msg_len;
            msg_len = stashlen;
        }
    }
    assert(msg_len <= BLAKE2B_BLOCKBYTES);
    memcpy(ctx->buf + ctx->buflen, in, msg_len);
    ctx->buflen += msg_len;
    
    //return 1;
}
/*
 @function: end an ripemd160 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:msg pointer to hash intermidate intermidiate result
 @paramter[in]:msg_len denotes the byte length of md
 @paramter[in]:c pointer to BLAKE2B_CTX structure
 */
void blake2b_final(BLAKE2B_CTX *ctx,uint8_t *msg, uint8_t msg_len)
{
    //int i;
    blake2b_set_lastblock(ctx);
    memset(ctx->buf + ctx->buflen, 0, sizeof(ctx->buf) - ctx->buflen);
    blake2b_compress(ctx, ctx->buf, ctx->buflen);
    memcpy(msg, (uint8_t *)&ctx->h[0], msg_len);
    //return 1;
}


/*
 @function:BLAKE2b hash
 @paramter[in]:msg pointer to the data to do hash
 @paramter[in]:msg_len denotes the byte length of msg
 @paramter[in]:key pointer to the key(if dosen't need key,please input NULL)
 @paramter[in]:key_len denotes the byte length of key.(if dosen't need key,please set key_len to zero)
 @paramter[in]:digest_len denotes the expected hash result length(rang in[1,64])
 @paramter[out]:digest pointer to hash result
 */
void blake2b(uint8_t *msg, uint16_t msg_len,uint8_t *key,uint16_t key_len, uint8_t digest_len, uint8_t *digest)
{
    BLAKE2B_CTX param;
    blake2b_init(&param, key,key_len,digest_len);
    blake2b_update(&param, msg, msg_len);
    blake2b_final(&param,digest, digest_len);
}

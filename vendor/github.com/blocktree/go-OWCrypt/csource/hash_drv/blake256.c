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
#include "blake256.h"


#define BLAKE256_U8TO32_BIG(p)                          \
(((uint32_t)((p)[0]) << 24) | ((uint32_t)((p)[1]) << 16) |   \
((uint32_t)((p)[2]) <<  8) | ((uint32_t)((p)[3])      ))


#define BLAKE256_U32TO8_BIG(p, v)                         \
(p)[0] = (uint8_t)((v) >> 24); (p)[1] = (uint8_t)((v) >> 16); \
(p)[2] = (uint8_t)((v) >>  8); (p)[3] = (uint8_t)((v)      );


static uint8_t balke256_sigma[][16] =
{
    { 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15 },
    {14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3 },
    {11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4 },
    { 7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8 },
    { 9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13 },
    { 2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9 },
    {12, 5, 1, 15, 14, 13, 4, 10, 0, 7, 6, 3, 9, 2, 8, 11 },
    {13, 11, 7, 14, 12, 1, 3, 9, 5, 0, 15, 4, 8, 6, 2, 10 },
    { 6, 15, 14, 9, 11, 3, 0, 8, 12, 2, 13, 7, 1, 4, 10, 5 },
    {10, 2, 8, 4, 7, 6, 1, 5, 15, 11, 9, 14, 3, 12, 13 , 0 },
    { 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15 },
    {14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3 },
    {11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4 },
    { 7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8 },
    { 9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13 },
    { 2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9 }
};

static const uint8_t blake256_padding[129] =
{
    0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
};

static uint32_t blake256_u256[16] =
{
    0x243f6a88, 0x85a308d3, 0x13198a2e, 0x03707344,
    0xa4093822, 0x299f31d0, 0x082efa98, 0xec4e6c89,
    0x452821e6, 0x38d01377, 0xbe5466cf, 0x34e90c6c,
    0xc0ac29b7, 0xc97c50dd, 0x3f84d5b5, 0xb5470917
};
static void blake256_compress( BLAKE256_CTX *S, const uint8_t *block )
{
    uint32_t v[16], m[16], i;
#define ROT(x,n) (((x)<<(32-n))|( (x)>>(n)))
#define G(a,b,c,d,e)          \
v[a] += (m[balke256_sigma[i][e]] ^ blake256_u256[balke256_sigma[i][e+1]]) + v[b]; \
v[d] = ROT( v[d] ^ v[a],16);        \
v[c] += v[d];           \
v[b] = ROT( v[b] ^ v[c],12);        \
v[a] += (m[balke256_sigma[i][e+1]] ^ blake256_u256[balke256_sigma[i][e]])+v[b]; \
v[d] = ROT( v[d] ^ v[a], 8);        \
v[c] += v[d];           \
v[b] = ROT( v[b] ^ v[c], 7);
    
    for( i = 0; i < 16; ++i )  m[i] = BLAKE256_U8TO32_BIG( block + i * 4 );
    
    for( i = 0; i < 8; ++i )  v[i] = S->h[i];
    
    v[ 8] = S->s[0] ^ blake256_u256[0];
    v[ 9] = S->s[1] ^ blake256_u256[1];
    v[10] = S->s[2] ^ blake256_u256[2];
    v[11] = S->s[3] ^ blake256_u256[3];
    v[12] = blake256_u256[4];
    v[13] = blake256_u256[5];
    v[14] = blake256_u256[6];
    v[15] = blake256_u256[7];
    
    /* don't xor t when the block is only padding */
    if ( !S->nullt )
    {
        v[12] ^= S->t[0];
        v[13] ^= S->t[0];
        v[14] ^= S->t[1];
        v[15] ^= S->t[1];
    }
    
    for( i = 0; i < 14; ++i )
    {
        /* column step */
        G( 0,  4,  8, 12,  0 );
        G( 1,  5,  9, 13,  2 );
        G( 2,  6, 10, 14,  4 );
        G( 3,  7, 11, 15,  6 );
        /* diagonal step */
        G( 0,  5, 10, 15,  8 );
        G( 1,  6, 11, 12, 10 );
        G( 2,  7,  8, 13, 12 );
        G( 3,  4,  9, 14, 14 );
    }
    
    for( i = 0; i < 16; ++i )  S->h[i % 8] ^= v[i];
    
    for( i = 0; i < 8 ; ++i )  S->h[i] ^= S->s[i % 4];
}

void blake256_init( BLAKE256_CTX *ctx )
{
    ctx->h[0] = 0x6a09e667;
    ctx->h[1] = 0xbb67ae85;
    ctx->h[2] = 0x3c6ef372;
    ctx->h[3] = 0xa54ff53a;
    ctx->h[4] = 0x510e527f;
    ctx->h[5] = 0x9b05688c;
    ctx->h[6] = 0x1f83d9ab;
    ctx->h[7] = 0x5be0cd19;
    ctx->t[0] = ctx->t[1] = ctx->buflen = ctx->nullt = 0;
    ctx->s[0] = ctx->s[1] = ctx->s[2] = ctx->s[3] = 0;
}

void blake256_update( BLAKE256_CTX *ctx, const uint8_t *msg, uint64_t msg_len )
{
    int left = ctx->buflen;
    int fill = 64 - left;
    
    /* data left and data received fill a block  */
    if( left && ( msg_len >= fill ) )
    {
        memcpy( ( void * ) ( ctx->buf + left ), ( void * ) msg, fill );
        ctx->t[0] += 512;
        
        if ( ctx->t[0] == 0 ) ctx->t[1]++;
        
        blake256_compress( ctx, ctx->buf );
        msg += fill;
        msg_len  -= fill;
        left = 0;
    }
    /* compress blocks of data received */
    while( msg_len >= 64 )
    {
        ctx->t[0] += 512;
        
        if ( ctx->t[0] == 0 ) ctx->t[1]++;
        
        blake256_compress( ctx, msg );
        msg += 64;
        msg_len -= 64;
    }
    
    /* store any data left */
    if( msg_len > 0 )
    {
        memcpy( ( void * ) ( ctx->buf + left ), ( void * ) msg, ( size_t ) msg_len );
        ctx->buflen = left + ( int )msg_len;
    }
    else ctx->buflen = 0;
}

void blake256_final( BLAKE256_CTX *ctx, uint8_t *digest )
{
    uint8_t msglen[8], zo = 0x01, oo = 0x81;
    uint32_t lo = ctx->t[0] + ( ctx->buflen << 3 ), hi = ctx->t[1];
    /* support for hashing more than 2^32 bits */
    if (lo < ( ctx->buflen << 3 ) ) hi++;
    BLAKE256_U32TO8_BIG(  msglen + 0, hi );
    BLAKE256_U32TO8_BIG(  msglen + 4, lo );
    if ( ctx->buflen == 55 )   /* one padding byte */
    {
        ctx->t[0] -= 8;
        blake256_update( ctx, &oo, 1 );
    }
    else
    {
        if ( ctx->buflen < 55 )   /* enough space to fill the block  */
        {
            if ( !ctx->buflen ) ctx->nullt = 1;
            
            ctx->t[0] -= 440 - ( ctx->buflen << 3 );
            blake256_update( ctx, blake256_padding, 55 - ctx->buflen );
        }
        else   /* need 2 compressions */
        {
            ctx->t[0] -= 512 - ( ctx->buflen << 3 );
            blake256_update( ctx, blake256_padding, 64 - ctx->buflen );
            ctx->t[0] -= 440;
            blake256_update( ctx, blake256_padding + 1, 55 );
            ctx->nullt = 1;
        }
        blake256_update( ctx, &zo, 1 );
        ctx->t[0] -= 8;
    }
    ctx->t[0] -= 64;
    blake256_update( ctx, msglen, 8 );
    BLAKE256_U32TO8_BIG( digest + 0, ctx->h[0] );
    BLAKE256_U32TO8_BIG( digest + 4, ctx->h[1] );
    BLAKE256_U32TO8_BIG( digest + 8, ctx->h[2] );
    BLAKE256_U32TO8_BIG( digest + 12, ctx->h[3] );
    BLAKE256_U32TO8_BIG( digest + 16, ctx->h[4] );
    BLAKE256_U32TO8_BIG( digest + 20, ctx->h[5] );
    BLAKE256_U32TO8_BIG( digest + 24, ctx->h[6] );
    BLAKE256_U32TO8_BIG( digest + 28, ctx->h[7] );
}


void blake256_hash(const uint8_t *msg, uint64_t msg_len,uint8_t *digest)
{
    BLAKE256_CTX ctx;
    blake256_init( &ctx);
    blake256_update( &ctx, msg, msg_len );
    blake256_final( &ctx, digest);
}

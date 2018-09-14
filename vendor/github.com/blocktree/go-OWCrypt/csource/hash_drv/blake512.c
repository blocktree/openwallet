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

#include "blake512.h"



#define BLAKE512_U8TO32_BIG(p)                          \
(((uint32_t)((p)[0]) << 24) | ((uint32_t)((p)[1]) << 16) |   \
((uint32_t)((p)[2]) <<  8) | ((uint32_t)((p)[3])      ))



#define BLAKE512_U32TO8_BIG(p, v)                         \
(p)[0] = (uint8_t)((v) >> 24); (p)[1] = (uint8_t)((v) >> 16); \
(p)[2] = (uint8_t)((v) >>  8); (p)[3] = (uint8_t)((v)      );



#define BLAKE512_U8TO64_BIG(p) \
(((uint64_t)BLAKE512_U8TO32_BIG(p) << 32) | (uint64_t)BLAKE512_U8TO32_BIG((p) + 4))



#define BLAKE512_U64TO8_BIG(p, v)                \
BLAKE512_U32TO8_BIG((p),     (uint32_t)((v) >> 32)); \
BLAKE512_U32TO8_BIG((p) + 4, (uint32_t)((v)      ));


static uint8_t blake512_sigma[][16] =
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

const uint64_t blake512_u512[16] =
{
    0x243f6a8885a308d3ULL, 0x13198a2e03707344ULL,
    0xa4093822299f31d0ULL, 0x082efa98ec4e6c89ULL,
    0x452821e638d01377ULL, 0xbe5466cf34e90c6cULL,
    0xc0ac29b7c97c50ddULL, 0x3f84d5b5b5470917ULL,
    0x9216d5d98979fb1bULL, 0xd1310ba698dfb5acULL,
    0x2ffd72dbd01adfb7ULL, 0xb8e1afed6a267e96ULL,
    0xba7c9045f12c7f99ULL, 0x24a19947b3916cf7ULL,
    0x0801f2e2858efc16ULL, 0x636920d871574e69ULL
};

static const uint8_t blake512_padding[129] =
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


static void blake512_compress( BLAKE512_CTX *S, const uint8_t *block)
{
    uint64_t v[16], m[16], i;
#define ROT(x,n) (((x)<<(64-n))|( (x)>>(n)))
#define G(a,b,c,d,e)          \
v[a] += (m[blake512_sigma[i][e]] ^ blake512_u512[blake512_sigma[i][e+1]]) + v[b];\
v[d] = ROT( v[d] ^ v[a],32);        \
v[c] += v[d];           \
v[b] = ROT( v[b] ^ v[c],25);        \
v[a] += (m[blake512_sigma[i][e+1]] ^ blake512_u512[blake512_sigma[i][e]])+v[b];  \
v[d] = ROT( v[d] ^ v[a],16);        \
v[c] += v[d];           \
v[b] = ROT( v[b] ^ v[c],11);
    for( i = 0; i < 16; ++i )  m[i] = BLAKE512_U8TO64_BIG( block + i * 8 );
    for( i = 0; i < 8; ++i )  v[i] = S->h[i];
    v[ 8] = S->s[0] ^ blake512_u512[0];
    v[ 9] = S->s[1] ^ blake512_u512[1];
    v[10] = S->s[2] ^ blake512_u512[2];
    v[11] = S->s[3] ^ blake512_u512[3];
    v[12] =  blake512_u512[4];
    v[13] =  blake512_u512[5];
    v[14] =  blake512_u512[6];
    v[15] =  blake512_u512[7];
    /* don't xor t when the block is only padding */
    if ( !S->nullt )
    {
        v[12] ^= S->t[0];
        v[13] ^= S->t[0];
        v[14] ^= S->t[1];
        v[15] ^= S->t[1];
    }
    for( i = 0; i < 16; ++i )
    {
        /* column step */
        G( 0, 4, 8, 12, 0 );
        G( 1, 5, 9, 13, 2 );
        G( 2, 6, 10, 14, 4 );
        G( 3, 7, 11, 15, 6 );
        /* diagonal step */
        G( 0, 5, 10, 15, 8 );
        G( 1, 6, 11, 12, 10 );
        G( 2, 7, 8, 13, 12 );
        G( 3, 4, 9, 14, 14 );
    }
    for( i = 0; i < 16; ++i )  S->h[i % 8] ^= v[i];
    for( i = 0; i < 8 ; ++i )  S->h[i] ^= S->s[i % 4];
}

void blake512_init(BLAKE512_CTX *ctx)
{
    ctx->h[0] = 0x6a09e667f3bcc908ULL;
    ctx->h[1] = 0xbb67ae8584caa73bULL;
    ctx->h[2] = 0x3c6ef372fe94f82bULL;
    ctx->h[3] = 0xa54ff53a5f1d36f1ULL;
    ctx->h[4] = 0x510e527fade682d1ULL;
    ctx->h[5] = 0x9b05688c2b3e6c1fULL;
    ctx->h[6] = 0x1f83d9abfb41bd6bULL;
    ctx->h[7] = 0x5be0cd19137e2179ULL;
    ctx->t[0] = ctx->t[1] = ctx->buflen = ctx->nullt = 0;
    ctx->s[0] = ctx->s[1] = ctx->s[2] = ctx->s[3] = 0;
}


void blake512_update( BLAKE512_CTX *ctx, const uint8_t *msg, uint64_t msg_len)
{
    int left = ctx->buflen;
    int fill = 128 - left;
    
    /* data left and data received fill a block  */
    if( left && ( msg_len >= fill ) )
    {
        memcpy( ( void * ) ( ctx->buf + left ), ( void * ) msg, fill );
        ctx->t[0] += 1024;
        if ( ctx->t[0] == 0 ) ctx->t[1]++;
        blake512_compress( ctx, ctx->buf );
        msg += fill;
        msg_len  -= fill;
        left = 0;
    }
    
    /* compress blocks of data received */
    while( msg_len >= 128 )
    {
        ctx->t[0] += 1024;
        if ( ctx->t[0] == 0 ) ctx->t[1]++;
        blake512_compress( ctx, msg );
        msg += 128;
        msg_len -= 128;
    }
    
    /* store any data left */
    if( msg_len > 0 )
    {
        memcpy( ( void * ) ( ctx->buf + left ),( void * ) msg, ( size_t ) msg_len );
        ctx->buflen = left + ( int )msg_len;
    }
    else ctx->buflen = 0;
}


void blake512_final(BLAKE512_CTX *ctx, uint8_t *out)
{
    uint8_t msglen[16], zo = 0x01, oo = 0x81;
    uint64_t lo = ctx->t[0] + ( ctx->buflen << 3 ), hi = ctx->t[1];
    
    /* support for hashing more than 2^32 bits */
    if ( lo < ( ctx->buflen << 3 ) ) hi++;
    
    BLAKE512_U64TO8_BIG(  msglen + 0, hi );
    BLAKE512_U64TO8_BIG(  msglen + 8, lo );
    
    if ( ctx->buflen == 111 )   /* one padding byte */
    {
        ctx->t[0] -= 8;
        blake512_update( ctx, &oo, 1 );
    }
    else
    {
        if ( ctx->buflen < 111 )  /* enough space to fill the block */
        {
            if ( !ctx->buflen ) ctx->nullt = 1;
            
            ctx->t[0] -= 888 - ( ctx->buflen << 3 );
            blake512_update( ctx, blake512_padding, 111 - ctx->buflen );
        }
        else   /* need 2 compressions */
        {
            ctx->t[0] -= 1024 - ( ctx->buflen << 3 );
            blake512_update( ctx, blake512_padding, 128 - ctx->buflen );
            ctx->t[0] -= 888;
            blake512_update( ctx, blake512_padding + 1, 111 );
            ctx->nullt = 1;
        }
        blake512_update( ctx, &zo, 1 );
        ctx->t[0] -= 8;
    }
    ctx->t[0] -= 128;
    blake512_update( ctx, msglen, 16 );
    BLAKE512_U64TO8_BIG( out + 0, ctx->h[0] );
    BLAKE512_U64TO8_BIG( out + 8, ctx->h[1] );
    BLAKE512_U64TO8_BIG( out + 16, ctx->h[2] );
    BLAKE512_U64TO8_BIG( out + 24, ctx->h[3] );
    BLAKE512_U64TO8_BIG( out + 32, ctx->h[4] );
    BLAKE512_U64TO8_BIG( out + 40, ctx->h[5] );
    BLAKE512_U64TO8_BIG( out + 48, ctx->h[6] );
    BLAKE512_U64TO8_BIG( out + 56, ctx->h[7] );
}

void blake512_hash(const uint8_t *msg, uint64_t msg_len,uint8_t *digest)
{
    BLAKE512_CTX ctx;
    blake512_init( &ctx );
    blake512_update( &ctx, msg, msg_len );
    blake512_final( &ctx, digest);
}

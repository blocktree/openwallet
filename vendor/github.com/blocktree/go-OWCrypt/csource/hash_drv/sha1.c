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

#include "sha1.h"
#ifndef GET_UINT32
#define GET_UINT32(n,b,i)                   \
{                                           \
(n) = ( (uint32_t) (b)[(i)    ] << 24 )       \
| ( (uint32_t) (b)[(i) + 1] << 16 )           \
| ( (uint32_t) (b)[(i) + 2] <<  8 )           \
| ( (uint32_t) (b)[(i) + 3]       );          \
}
#endif

#ifndef PUT_UINT32
#define PUT_UINT32(n,b,i)                    \
{                                            \
(b)[(i)    ] = (uint8_t) ( (n) >> 24 );        \
(b)[(i) + 1] = (uint8_t) ( (n) >> 16 );        \
(b)[(i) + 2] = (uint8_t) ( (n) >>  8 );        \
(b)[(i) + 3] = (uint8_t) ( (n)       );        \
}
#endif
/*
 @function:init SHA1_CTX,writing a new message
 @paramter[in]:ctx pointer to SHA1_CTX
 @return: NULL
 @notice: none
 */
void sha1_init(SHA1_CTX *ctx)
{
    if(!ctx)
        return ;
    ctx->total[0] = 0;
    ctx->total[1] = 0;
    ctx->state[0] = 0x67452301;
    ctx->state[1] = 0xEFCDAB89;
    ctx->state[2] = 0x98BADCFE;
    ctx->state[3] = 0x10325476;
    ctx->state[4] = 0xC3D2E1F0;
}

/*
 @function: sha1 interate operation
 @parameter[in]:ctx pointer to SHA1_CTX
 @parameter[in]:data pointer to the message to do hash
 @return: none
 @notice:nothing
 */

void sha1_process(SHA1_CTX *ctx, uint8_t data[64] )
{
    uint32_t temp, A, B, C, D, E, W[16];
    if(!ctx || !data)
        return ;
    GET_UINT32( W[0],  data,  0 );
    GET_UINT32( W[1],  data,  4 );
    GET_UINT32( W[2],  data,  8 );
    GET_UINT32( W[3],  data, 12 );
    GET_UINT32( W[4],  data, 16 );
    GET_UINT32( W[5],  data, 20 );
    GET_UINT32( W[6],  data, 24 );
    GET_UINT32( W[7],  data, 28 );
    GET_UINT32( W[8],  data, 32 );
    GET_UINT32( W[9],  data, 36 );
    GET_UINT32( W[10], data, 40 );
    GET_UINT32( W[11], data, 44 );
    GET_UINT32( W[12], data, 48 );
    GET_UINT32( W[13], data, 52 );
    GET_UINT32( W[14], data, 56 );
    GET_UINT32( W[15], data, 60 );
#define S_SHA1(x,n) ((x << n) | ((x & 0xFFFFFFFF) >> (32 - n)))
    

#define R_SHA1(t)                                           \
(                                                      \
temp = W[(t -  3) & 0x0F] ^ W[(t - 8) & 0x0F] ^        \
W[(t - 14) & 0x0F] ^ W[ t      & 0x0F],                \
( W[t & 0x0F] = S_SHA1(temp,1) )                            \
)


#define P_SHA1(a,b,c,d,e,x)                                 \
{                                                      \
e += S_SHA1(a,5) + F_SHA1(b,c,d) + K_SHA1 + x; b = S_SHA1(b,30);           \
}

    A = ctx->state[0];
    B = ctx->state[1];
    C = ctx->state[2];
    D = ctx->state[3];
    E = ctx->state[4];
#define F_SHA1(x,y,z) (z ^ (x & (y ^ z)))
#define K_SHA1 0x5A827999
    
    P_SHA1( A, B, C, D, E, W[0]  );
    P_SHA1( E, A, B, C, D, W[1]  );
    P_SHA1( D, E, A, B, C, W[2]  );
    P_SHA1( C, D, E, A, B, W[3]  );
    P_SHA1( B, C, D, E, A, W[4]  );
    P_SHA1( A, B, C, D, E, W[5]  );
    P_SHA1( E, A, B, C, D, W[6]  );
    P_SHA1( D, E, A, B, C, W[7]  );
    P_SHA1( C, D, E, A, B, W[8]  );
    P_SHA1( B, C, D, E, A, W[9]  );
    P_SHA1( A, B, C, D, E, W[10] );
    P_SHA1( E, A, B, C, D, W[11] );
    P_SHA1( D, E, A, B, C, W[12] );
    P_SHA1( C, D, E, A, B, W[13] );
    P_SHA1( B, C, D, E, A, W[14] );
    P_SHA1( A, B, C, D, E, W[15] );
    P_SHA1( E, A, B, C, D, R_SHA1(16) );
    P_SHA1( D, E, A, B, C, R_SHA1(17) );
    P_SHA1( C, D, E, A, B, R_SHA1(18) );
    P_SHA1( B, C, D, E, A, R_SHA1(19) );
#undef K
#undef F

#define F(x,y,z) (x ^ y ^ z)

#define K 0x6ED9EBA1
    P_SHA1( A, B, C, D, E, R_SHA1(20) );
    P_SHA1( E, A, B, C, D, R_SHA1(21) );
    P_SHA1( D, E, A, B, C, R_SHA1(22) );
    P_SHA1( C, D, E, A, B, R_SHA1(23) );
    P_SHA1( B, C, D, E, A, R_SHA1(24) );
    P_SHA1( A, B, C, D, E, R_SHA1(25) );
    P_SHA1( E, A, B, C, D, R_SHA1(26) );
    P_SHA1( D, E, A, B, C, R_SHA1(27) );
    P_SHA1( C, D, E, A, B, R_SHA1(28) );
    P_SHA1( B, C, D, E, A, R_SHA1(29) );
    P_SHA1( A, B, C, D, E, R_SHA1(30) );
    P_SHA1( E, A, B, C, D, R_SHA1(31) );
    P_SHA1( D, E, A, B, C, R_SHA1(32) );
    P_SHA1( C, D, E, A, B, R_SHA1(33) );
    P_SHA1( B, C, D, E, A, R_SHA1(34) );
    P_SHA1( A, B, C, D, E, R_SHA1(35) );
    P_SHA1( E, A, B, C, D, R_SHA1(36) );
    P_SHA1( D, E, A, B, C, R_SHA1(37) );
    P_SHA1( C, D, E, A, B, R_SHA1(38) );
    P_SHA1( B, C, D, E, A, R_SHA1(39) );
#undef K
#undef F
#define F(x,y,z) ((x & y) | (z & (x | y)))
#define K 0x8F1BBCDC
    P_SHA1( A, B, C, D, E, R_SHA1(40) );
    P_SHA1( E, A, B, C, D, R_SHA1(41) );
    P_SHA1( D, E, A, B, C, R_SHA1(42) );
    P_SHA1( C, D, E, A, B, R_SHA1(43) );
    P_SHA1( B, C, D, E, A, R_SHA1(44) );
    P_SHA1( A, B, C, D, E, R_SHA1(45) );
    P_SHA1( E, A, B, C, D, R_SHA1(46) );
    P_SHA1( D, E, A, B, C, R_SHA1(47) );
    P_SHA1( C, D, E, A, B, R_SHA1(48) );
    P_SHA1( B, C, D, E, A, R_SHA1(49) );
    P_SHA1( A, B, C, D, E, R_SHA1(50) );
    P_SHA1( E, A, B, C, D, R_SHA1(51) );
    P_SHA1( D, E, A, B, C, R_SHA1(52) );
    P_SHA1( C, D, E, A, B, R_SHA1(53) );
    P_SHA1( B, C, D, E, A, R_SHA1(54) );
    P_SHA1( A, B, C, D, E, R_SHA1(55) );
    P_SHA1( E, A, B, C, D, R_SHA1(56) );
    P_SHA1( D, E, A, B, C, R_SHA1(57) );
    P_SHA1( C, D, E, A, B, R_SHA1(58) );
    P_SHA1( B, C, D, E, A, R_SHA1(59) );
    
#undef K
#undef F
#define F(x,y,z) (x ^ y ^ z)
#define K 0xCA62C1D6
    
    P_SHA1( A, B, C, D, E, R_SHA1(60) );
    P_SHA1( E, A, B, C, D, R_SHA1(61) );
    P_SHA1( D, E, A, B, C, R_SHA1(62) );
    P_SHA1( C, D, E, A, B, R_SHA1(63) );
    P_SHA1( B, C, D, E, A, R_SHA1(64) );
    P_SHA1( A, B, C, D, E, R_SHA1(65) );
    P_SHA1( E, A, B, C, D, R_SHA1(66) );
    P_SHA1( D, E, A, B, C, R_SHA1(67) );
    P_SHA1( C, D, E, A, B, R_SHA1(68) );
    P_SHA1( B, C, D, E, A, R_SHA1(69) );
    P_SHA1( A, B, C, D, E, R_SHA1(70) );
    P_SHA1( E, A, B, C, D, R_SHA1(71) );
    P_SHA1( D, E, A, B, C, R_SHA1(72) );
    P_SHA1( C, D, E, A, B, R_SHA1(73) );
    P_SHA1( B, C, D, E, A, R_SHA1(74) );
    P_SHA1( A, B, C, D, E, R_SHA1(75) );
    P_SHA1( E, A, B, C, D, R_SHA1(76) );
    P_SHA1( D, E, A, B, C, R_SHA1(77) );
    P_SHA1( C, D, E, A, B, R_SHA1(78) );
    P_SHA1( B, C, D, E, A, R_SHA1(79) );
    
#undef K
#undef F
    
    ctx->state[0] += A;
    ctx->state[1] += B;
    ctx->state[2] += C;
    ctx->state[3] += D;
    ctx->state[4] += E;
}

/*
 @function:update message Continues an sha1 message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:ctx pointer to SHA1_CTX
 @paramter[in]:input pointer to the message to do sha1
 @paramter[in]:inputlen,the byte length of input
 @return:NULL
 @notoce:none
 */

void sha1_update( SHA1_CTX *ctx, uint8_t *msg, uint32_t msg_len )
{
    uint32_t left, fill;
    
    if(!msg_len)
    {
        return ;
    }
    if(!ctx || !msg)
        return;
    left = ( ctx->total[0] >> 3 ) & 0x3F;
    fill = 64 - left;
    
    ctx->total[0] += msg_len <<  3;
    ctx->total[1] += msg_len >> 29;
    
    ctx->total[0] &= 0xFFFFFFFF;
    ctx->total[1] += ctx->total[0] < ( msg_len << 3 );
    
    if( left && msg_len >= fill )
    {
        memcpy( (void *) (ctx->buffer + left), (void *) msg, fill );
        sha1_process( ctx, ctx->buffer );
        msg_len -= fill;
        msg  += fill;
        left = 0;
    }
    while( msg_len >= 64 )
    {
        sha1_process( ctx, msg );
        msg_len -= 64;
        msg  += 64;
    }
    if( msg_len )
    {
        memcpy( (void *) (ctx->buffer + left), (void *) msg, msg_len );
    }
}

static uint8_t sha1_padding[64] =
{
    0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
};

/*
 @function:finalization sha1 operation ends an sha1 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to SHA1_CTX
 @paramter[out]:digest pointer to sha1 hash result
 @return:NULL
 @notice:nothing
 */
void sha1_final(SHA1_CTX *ctx, uint8_t digest[20] )
{
    uint32_t last, padn;
    uint8_t msglen[8];
    if(!ctx || !digest)
    {
        return ;
    }
    PUT_UINT32( ctx->total[1], msglen, 0 );
    PUT_UINT32( ctx->total[0], msglen, 4 );
    last = ( ctx->total[0] >> 3 ) & 0x3F;
    padn = ( last < 56 ) ? ( 56 - last ) : ( 120 - last );
    
    sha1_update( ctx, sha1_padding, padn );
    sha1_update( ctx, msglen, 8 );
    
    PUT_UINT32( ctx->state[0], digest,  0 );
    PUT_UINT32( ctx->state[1], digest,  4 );
    PUT_UINT32( ctx->state[2], digest,  8 );
    PUT_UINT32( ctx->state[3], digest, 12 );
    PUT_UINT32( ctx->state[4], digest, 16 );
    return;
}

/*
 @function: sha1 hash
 @parameter[in]:input pointer to the message to do hash
 @parameter[in]:the byte length of input
 @parameter[in]:digest pointer to hash result
 @return: none
 @notice:nothing
 */

void sha1_hash(uint8_t *msg, uint32_t msg_len,uint8_t digest[20])
{
    if(!msg || !digest)
    {
        return ;
    }
    if(!msg_len)
    {
        return ;
    }
    SHA1_CTX ctx;
    sha1_init(&ctx);
    sha1_update(&ctx, msg, msg_len);
    sha1_final(&ctx, digest);
    return;
}

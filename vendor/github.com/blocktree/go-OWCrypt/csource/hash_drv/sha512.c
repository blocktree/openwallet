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

#include "sha512.h"


#define UL64(x) x##ULL

#ifndef GET_UINT64_BE
#define GET_UINT64_BE(n,b,i)                            \
{                                                       \
(n) = ( (uint64_t) (b)[(i)    ] << 56 )       \
| ( (uint64_t) (b)[(i) + 1] << 48 )       \
| ( (uint64_t) (b)[(i) + 2] << 40 )       \
| ( (uint64_t) (b)[(i) + 3] << 32 )       \
| ( (uint64_t) (b)[(i) + 4] << 24 )       \
| ( (uint64_t) (b)[(i) + 5] << 16 )       \
| ( (uint64_t) (b)[(i) + 6] <<  8 )       \
| ( (uint64_t) (b)[(i) + 7]       );      \
}
#endif /* GET_UINT64_BE */

#ifndef PUT_UINT64_BE
#define PUT_UINT64_BE(n,b,i)                            \
{                                                       \
(b)[(i)    ] = (unsigned char) ( (n) >> 56 );       \
(b)[(i) + 1] = (unsigned char) ( (n) >> 48 );       \
(b)[(i) + 2] = (unsigned char) ( (n) >> 40 );       \
(b)[(i) + 3] = (unsigned char) ( (n) >> 32 );       \
(b)[(i) + 4] = (unsigned char) ( (n) >> 24 );       \
(b)[(i) + 5] = (unsigned char) ( (n) >> 16 );       \
(b)[(i) + 6] = (unsigned char) ( (n) >>  8 );       \
(b)[(i) + 7] = (unsigned char) ( (n)       );       \
}
#endif /* PUT_UINT64_BE */


static const uint64_t K_SHA512[80] =
{
    UL64(0x428A2F98D728AE22),  UL64(0x7137449123EF65CD),
    UL64(0xB5C0FBCFEC4D3B2F),  UL64(0xE9B5DBA58189DBBC),
    UL64(0x3956C25BF348B538),  UL64(0x59F111F1B605D019),
    UL64(0x923F82A4AF194F9B),  UL64(0xAB1C5ED5DA6D8118),
    UL64(0xD807AA98A3030242),  UL64(0x12835B0145706FBE),
    UL64(0x243185BE4EE4B28C),  UL64(0x550C7DC3D5FFB4E2),
    UL64(0x72BE5D74F27B896F),  UL64(0x80DEB1FE3B1696B1),
    UL64(0x9BDC06A725C71235),  UL64(0xC19BF174CF692694),
    UL64(0xE49B69C19EF14AD2),  UL64(0xEFBE4786384F25E3),
    UL64(0x0FC19DC68B8CD5B5),  UL64(0x240CA1CC77AC9C65),
    UL64(0x2DE92C6F592B0275),  UL64(0x4A7484AA6EA6E483),
    UL64(0x5CB0A9DCBD41FBD4),  UL64(0x76F988DA831153B5),
    UL64(0x983E5152EE66DFAB),  UL64(0xA831C66D2DB43210),
    UL64(0xB00327C898FB213F),  UL64(0xBF597FC7BEEF0EE4),
    UL64(0xC6E00BF33DA88FC2),  UL64(0xD5A79147930AA725),
    UL64(0x06CA6351E003826F),  UL64(0x142929670A0E6E70),
    UL64(0x27B70A8546D22FFC),  UL64(0x2E1B21385C26C926),
    UL64(0x4D2C6DFC5AC42AED),  UL64(0x53380D139D95B3DF),
    UL64(0x650A73548BAF63DE),  UL64(0x766A0ABB3C77B2A8),
    UL64(0x81C2C92E47EDAEE6),  UL64(0x92722C851482353B),
    UL64(0xA2BFE8A14CF10364),  UL64(0xA81A664BBC423001),
    UL64(0xC24B8B70D0F89791),  UL64(0xC76C51A30654BE30),
    UL64(0xD192E819D6EF5218),  UL64(0xD69906245565A910),
    UL64(0xF40E35855771202A),  UL64(0x106AA07032BBD1B8),
    UL64(0x19A4C116B8D2D0C8),  UL64(0x1E376C085141AB53),
    UL64(0x2748774CDF8EEB99),  UL64(0x34B0BCB5E19B48A8),
    UL64(0x391C0CB3C5C95A63),  UL64(0x4ED8AA4AE3418ACB),
    UL64(0x5B9CCA4F7763E373),  UL64(0x682E6FF3D6B2B8A3),
    UL64(0x748F82EE5DEFB2FC),  UL64(0x78A5636F43172F60),
    UL64(0x84C87814A1F0AB72),  UL64(0x8CC702081A6439EC),
    UL64(0x90BEFFFA23631E28),  UL64(0xA4506CEBDE82BDE9),
    UL64(0xBEF9A3F7B2C67915),  UL64(0xC67178F2E372532B),
    UL64(0xCA273ECEEA26619C),  UL64(0xD186B8C721C0C207),
    UL64(0xEADA7DD6CDE0EB1E),  UL64(0xF57D4F7FEE6ED178),
    UL64(0x06F067AA72176FBA),  UL64(0x0A637DC5A2C898A6),
    UL64(0x113F9804BEF90DAE),  UL64(0x1B710B35131C471B),
    UL64(0x28DB77F523047D84),  UL64(0x32CAAB7B40C72493),
    UL64(0x3C9EBE0A15C9BEBC),  UL64(0x431D67C49C100D4C),
    UL64(0x4CC5D4BECB3E42B6),  UL64(0x597F299CFC657E2A),
    UL64(0x5FCB6FAB3AD6FAEC),  UL64(0x6C44198C4A475817)
};

static const unsigned char sha512_padding[128] =
{
     0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
};

/*
 @function:init SHA512_CTX,writing a new message
 @paramter[in]:ctx pointer to SHA512_CTX
 @return: NULL
 @notice: none
 */
void sha512_init(SHA512_CTX *ctx)
{
    memset((uint8_t*)ctx, 0, sizeof(SHA512_CTX));
    
    ctx -> state[0] = UL64(0x6A09E667F3BCC908);
    ctx -> state[1] = UL64(0xBB67AE8584CAA73B);
    ctx -> state[2] = UL64(0x3C6EF372FE94F82B);
    ctx -> state[3] = UL64(0xA54FF53A5F1D36F1);
    ctx -> state[4] = UL64(0x510E527FADE682D1);
    ctx -> state[5] = UL64(0x9B05688C2B3E6C1F);
    ctx -> state[6] = UL64(0x1F83D9ABFB41BD6B);
    ctx -> state[7] = UL64(0x5BE0CD19137E2179);
}

void sha512_process(SHA512_CTX *ctx, uint8_t data[128])
{
    int i;
    uint64_t temp1, temp2, W[80];
    uint64_t A, B, C, D, E, F, G, H;
    
#define  SHR_SHA512(x,n) (x >> n)
#define ROTR_SHA512(x,n) (SHR_SHA512(x,n) | (x << (64 - n)))
    
#define S0_SHA512(x) (ROTR_SHA512(x, 1) ^ ROTR_SHA512(x, 8) ^  SHR_SHA512(x, 7))
#define S1_SHA512(x) (ROTR_SHA512(x,19) ^ ROTR_SHA512(x,61) ^  SHR_SHA512(x, 6))
    
#define S2_SHA512(x) (ROTR_SHA512(x,28) ^ ROTR_SHA512(x,34) ^ ROTR_SHA512(x,39))
#define S3_SHA512(x) (ROTR_SHA512(x,14) ^ ROTR_SHA512(x,18) ^ ROTR_SHA512(x,41))
    
#define F0_SHA512(x,y,z) ((x & y) | (z & (x | y)))
#define F1_SHA512(x,y,z) (z ^ (x & (y ^ z)))
    
#define P_SHA512(a,b,c,d,e,f,g,h,x,K)                  \
{                                               \
    temp1 = h + S3_SHA512(e) + F1_SHA512(e,f,g) + K + x;      \
    temp2 = S2_SHA512(a) + F0_SHA512(a,b,c);                  \
    d += temp1; h = temp1 + temp2;              \
}
    
    
    for(i = 0; i < 16; i++)
    {
        GET_UINT64_BE(W[i], data, i << 3);
    }
    
    for( ; i < 80; i ++)
    {
        W[i] = S1_SHA512(W[i -  2]) + W[i -  7] +
        S0_SHA512(W[i - 15]) + W[i - 16];
    }
    
    A = ctx -> state[0];
    B = ctx -> state[1];
    C = ctx -> state[2];
    D = ctx -> state[3];
    E = ctx -> state[4];
    F = ctx -> state[5];
    G = ctx -> state[6];
    H = ctx -> state[7];
    i = 0;
    
    do
    {
        P_SHA512(A, B, C, D, E, F, G, H, W[i], K_SHA512[i]); i ++;
        P_SHA512(H, A, B, C, D, E, F, G, W[i], K_SHA512[i]); i ++;
        P_SHA512(G, H, A, B, C, D, E, F, W[i], K_SHA512[i]); i ++;
        P_SHA512(F, G, H, A, B, C, D, E, W[i], K_SHA512[i]); i ++;
        P_SHA512(E, F, G, H, A, B, C, D, W[i], K_SHA512[i]); i ++;
        P_SHA512(D, E, F, G, H, A, B, C, W[i], K_SHA512[i]); i ++;
        P_SHA512(C, D, E, F, G, H, A, B, W[i], K_SHA512[i]); i ++;
        P_SHA512(B, C, D, E, F, G, H, A, W[i], K_SHA512[i]); i ++;
    }
    while( i < 80 );
    
    ctx -> state[0] += A;
    ctx -> state[1] += B;
    ctx -> state[2] += C;
    ctx -> state[3] += D;
    ctx -> state[4] += E;
    ctx -> state[5] += F;
    ctx -> state[6] += G;
    ctx -> state[7] += H;
}

/*
 @function:update message Continues an sha512 message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:ctx pointer to SHA512_CTX
 @paramter[in]:msg pointer to the message to do sha512
 @paramter[in]:msg_len,the byte length of input
 @return:NULL
 @notoce:none
 */
void sha512_update(SHA512_CTX *ctx, uint8_t *msg, uint16_t msg_len)
{
    size_t fill;
    unsigned int left;
    
    if(msg_len == 0)
        return;
    
    left = (unsigned int) (ctx -> total[0] & 0x7F);
    fill = 128 - left;
    
    ctx -> total[0] += (uint64_t)msg_len;
    
    if(ctx -> total[0] < (uint64_t)msg_len)
        ctx -> total[1]++;
    
    if(left && msg_len >= fill)
    {
        memcpy((uint8_t *)(ctx -> buffer + left), msg, fill);
        sha512_process(ctx, ctx -> buffer);
        msg += fill;
        msg_len  -= fill;
        left = 0;
    }
    while(msg_len >= 128)
    {
        sha512_process(ctx, msg);
        msg += 128;
        msg_len  -= 128;
    }
    if(msg_len > 0)
        memcpy((uint8_t *) (ctx -> buffer + left), msg, msg_len);
}

/*
 @function:finalization sha512 operation ends an sha1 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to SHA512_CTX
 @paramter[out]:digest pointer to sha512 hash result
 @return:NULL
 @notice:nothing
 */
void sha512_final(SHA512_CTX *ctx, uint8_t *digest)
{
    size_t last, padn;
    uint64_t high, low;
    unsigned char msglen[16];
    
    high = (ctx -> total[0] >> 61)
    | (ctx -> total[1] <<  3);
    low  = (ctx -> total[0] <<  3);
    
    PUT_UINT64_BE(high, msglen, 0);
    PUT_UINT64_BE(low,  msglen, 8);
    
    last = (size_t)( ctx -> total[0] & 0x7F );
    padn = (last < 112) ? (112 - last) : (240 - last);
    
    sha512_update(ctx, (uint8_t *)sha512_padding, padn);
    sha512_update(ctx, msglen, 16);
    
    PUT_UINT64_BE(ctx -> state[0], digest,  0);
    PUT_UINT64_BE(ctx -> state[1], digest,  8);
    PUT_UINT64_BE(ctx -> state[2], digest, 16);
    PUT_UINT64_BE(ctx -> state[3], digest, 24);
    PUT_UINT64_BE(ctx -> state[4], digest, 32);
    PUT_UINT64_BE(ctx -> state[5], digest, 40);
    PUT_UINT64_BE(ctx -> state[6], digest, 48);
    PUT_UINT64_BE(ctx -> state[7], digest, 56);
}

/*
 @function: sha512 hash
 @parameter[in]:msg pointer to the message to do hash
 @parameter[in]:msg_len,the byte length of input
 @parameter[in]:digest pointer to hash result
 @return: none
 @notice:nothing
 */
void sha512_hash(uint8_t *msg, uint32_t msg_len, uint8_t *digest)
{
    SHA512_CTX *ctx = NULL;
    ctx = calloc(1, sizeof(SHA512_CTX));
    sha512_init(ctx);
    sha512_update(ctx, msg, msg_len);
    sha512_final(ctx, digest);
    free(ctx);
}


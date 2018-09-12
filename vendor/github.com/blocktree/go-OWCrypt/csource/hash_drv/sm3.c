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

#include "sm3.h"

void sm3_init(SM3_CTX *ctx)
{
    ctx->Nl = 0;
    ctx->Nh = 0;
    
    ctx->h0 = 0x7380166F;
    ctx->h1 = 0x4914B2B9;
    ctx->h2 = 0x172442D7;
    ctx->h3 = 0xDA8A0600;
    ctx->h4 = 0xA96F30BC;
    ctx->h5 = 0x163138AA;
    ctx->h6 = 0xE38DEE4D;
    ctx->h7 = 0xB0FB0E4E;
    memset(ctx->data, 0, SM3_CBLOCK);
}

static uint32_t SS1, SS2, TT1, TT2;
static uint32_t A, B, C, D, E, F, G, H;
static uint32_t Temp1,Temp2,Temp3,Temp4,Temp5;
static uint32_t j;

#define GET_ULONG_BE(n,b,i)      (n = ( (uint32_t) (b)[i] << 24 ) | ( (uint32_t) (b)[i + 1] << 16 ) | ( (uint32_t) (b)[i + 2] <<  8 ) | ( (uint32_t) (b)[i + 3]))
void PUT_ULONG_BE(uint32_t n,uint8_t* b)
{
    b[0] = (uint8_t) ( (n) >> 24 );
    b[1] = (uint8_t) ( (n) >> 16 );
    b[2] = (uint8_t) ( (n) >>  8 );
    b[3] = (uint8_t) ( (n) );
    
}

const uint32_t T[64] =
{
    0x79CC4519,0x79CC4519,0x79CC4519,0x79CC4519,0x79CC4519,0x79CC4519,0x79CC4519,0x79CC4519,
    0x79CC4519,0x79CC4519,0x79CC4519,0x79CC4519,0x79CC4519,0x79CC4519,0x79CC4519,0x79CC4519,
    0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,
    0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,
    0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,
    0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,
    0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,
    0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,0x7A879D8A,
};


void sm3_process(SM3_CTX *c)
{
    
    uint32_t* W = (uint32_t*)c->data;
    
    GET_ULONG_BE( W[ 0], (uint8_t*)c->data,  0 );
    GET_ULONG_BE( W[ 1], (uint8_t*)c->data,  4 );
    GET_ULONG_BE( W[ 2], (uint8_t*)c->data,  8 );
    GET_ULONG_BE( W[ 3], (uint8_t*)c->data, 12 );
    GET_ULONG_BE( W[ 4], (uint8_t*)c->data, 16 );
    GET_ULONG_BE( W[ 5], (uint8_t*)c->data, 20 );
    GET_ULONG_BE( W[ 6], (uint8_t*)c->data, 24 );
    GET_ULONG_BE( W[ 7], (uint8_t*)c->data, 28 );
    GET_ULONG_BE( W[ 8], (uint8_t*)c->data, 32 );
    GET_ULONG_BE( W[ 9], (uint8_t*)c->data, 36 );
    GET_ULONG_BE( W[10], (uint8_t*)c->data, 40 );
    GET_ULONG_BE( W[11], (uint8_t*)c->data, 44 );
    GET_ULONG_BE( W[12], (uint8_t*)c->data, 48 );
    GET_ULONG_BE( W[13], (uint8_t*)c->data, 52 );
    GET_ULONG_BE( W[14], (uint8_t*)c->data, 56 );
    GET_ULONG_BE( W[15], (uint8_t*)c->data, 60 );
    
    
#define FF0(x,y,z) ( (x) ^ (y) ^ (z))
#define FF1(x,y,z) (((x) & (y)) | ( (x) & (z)) | ( (y) & (z)))
    
#define GG0(x,y,z) ( (x) ^ (y) ^ (z))
#define GG1(x,y,z) (((x) & (y)) | ( (~(x)) & (z)) )
    
    
#define  SHL(x,n) (((x) & 0xFFFFFFFF) << n)
#define ROTL(x,n) (n<=32?(SHL((x),n) | ((x) >> (32 - n))):(SHL((x),(n%32)) | ((x) >> (32 - (n%32)))))
    
    
#define P0(x) ((x) ^  ROTL((x),9) ^ ROTL((x),17))
#define P1(x) ((x) ^  ROTL((x),15) ^ ROTL((x),23))
    
    for(j = 16; j < 68; j++ )
    {
        Temp1 = W[j-16] ^ W[j-9];
        Temp2 = ROTL(W[j-3],15);
        Temp3 = Temp1 ^ Temp2;
        Temp4 = P1(Temp3);
        Temp5 =  ROTL(W[j - 13],7 ) ^ W[j-6];
        W[j] = Temp4 ^ Temp5;
    }
    A = c->h0;
    B = c->h1;
    C = c->h2;
    D = c->h3;
    E = c->h4;
    F = c->h5;
    G = c->h6;
    H = c->h7;
    
    for(j =0; j < 16; j++)
    {
        SS1 = ROTL((ROTL(A,12) + E + ROTL(T[j],j)), 7);
        SS2 = SS1 ^ ROTL(A,12);
        TT1 = FF0(A,B,C) + D + SS2 + (W[j] ^ W[j+4]);
        TT2 = GG0(E,F,G) + H + SS1 + W[j];
        D = C;
        C = ROTL(B,9);
        B = A;
        A = TT1;
        H = G;
        G = ROTL(F,19);
        F = E;
        E = P0(TT2);
    }
    
    for(j =16; j < 64; j++)
    {
        SS1 = ROTL((ROTL(A,12) + E + ROTL(T[j],j)), 7);
        SS2 = SS1 ^ ROTL(A,12);
        TT1 = FF1(A,B,C) + D + SS2 + (W[j] ^ W[j+4]);
        TT2 = GG1(E,F,G) + H + SS1 + W[j];
        D = C;
        C = ROTL(B,9);
        B = A;
        A = TT1;
        H = G;
        G = ROTL(F,19);
        F = E;
        E = P0(TT2);
    }
    
    c->h0 ^= A;
    c->h1 ^= B;
    c->h2 ^= C;
    c->h3 ^= D;
    c->h4 ^= E;
    c->h5 ^= F;
    c->h6 ^= G;
    c->h7 ^= H;
    
}

void sm3_update(SM3_CTX *ctx, uint8_t *msg, size_t msg_len)
{
    uint32_t fill;
    uint32_t left;
    
    if( msg_len <= 0 )
        return;
    
    left = (uint8_t)(ctx->Nl & 0x3F);
    fill = 64 - left;
    
    ctx->Nl += msg_len;
    ctx->Nl &= 0xFFFFFFFF;
    
    if( ctx->Nl <  msg_len )
        ctx->Nh++;
    
    if(msg_len >= fill )
    {
        memcpy( ((uint8_t*)ctx->data + left),(uint8_t *) msg, fill );
        sm3_process(ctx);
        msg += fill;
        msg_len  -= fill;
        left = 0;
    }
    
    while( msg_len >= 64 )
    {
        memcpy( ((uint8_t *)ctx->data),(uint8_t *) msg, 64 );
        sm3_process(ctx);
        msg += 64;
        msg_len  -= 64;
    }
    
    if( msg_len > 0 )
    {
        memcpy(  ((uint8_t*)ctx->data + left),(uint8_t *) msg, msg_len );
    }
}
static const uint8_t sm3_padding[64] =
{
     0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
};


/*
 @function:finalization sm3 operation ends an sha1 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to SM3_CTX
 @paramter[out]:digest pointer to sm3 hash result
 @return:NULL
 @notice:nothing
 */
void sm3_final(SM3_CTX *ctx, uint8_t *digest)
{
    uint32_t last, padn;
    uint32_t high, low;
    uint8_t msglen[8];
    
    high = ( ctx->Nl >> 29 )
    | ( ctx->Nh <<  3 );
    low  = (ctx->Nl <<  3 );
    
    PUT_ULONG_BE( high, msglen);
    PUT_ULONG_BE( low,  msglen + 4);
    last = ctx->Nl & 0x3F;
    padn = ( last < 56 ) ? ( 56 - last ) : ( 120 - last );
    sm3_update( ctx,(uint8_t *) sm3_padding, padn );
    sm3_update( ctx,msglen, 8 );
    PUT_ULONG_BE( ctx->h0, (uint8_t *)ctx->data+0 );
    PUT_ULONG_BE( ctx->h1, (uint8_t *)ctx->data+4 );
    PUT_ULONG_BE( ctx->h2, (uint8_t *)ctx->data+8 );
    PUT_ULONG_BE( ctx->h3, (uint8_t *)ctx->data+12 );
    PUT_ULONG_BE( ctx->h4, (uint8_t *)ctx->data+16 );
    PUT_ULONG_BE( ctx->h5, (uint8_t *)ctx->data+20 );
    PUT_ULONG_BE( ctx->h6, (uint8_t *)ctx->data+24 );
    PUT_ULONG_BE( ctx->h7, (uint8_t *)ctx->data+28 );
    memcpy(digest, (uint8_t *)ctx->data, 32);
}

/*
 @function: sm3 hash
 @parameter[in]:msg pointer to the message to do hash
 @parameter[in]:msg_len,the byte length of msg
 @parameter[in]:digest pointer to hash result
 @return: none
 @notice:nothing
 */
void sm3_hash(uint8_t *msg, uint32_t msg_len, uint8_t *hash)
{
    SM3_CTX *sm3_ctx = NULL;
    sm3_ctx = calloc(1, sizeof(SM3_CTX));

    sm3_init(sm3_ctx);
    sm3_update(sm3_ctx, msg, msg_len);
    sm3_final(sm3_ctx, hash);
    
    free(sm3_ctx);
}

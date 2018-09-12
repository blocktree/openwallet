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

#include "ripemd160.h"

#define RMDsize  160

static const unsigned char ripemd160_padding[64] =
{
    0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
};

//split one word to four bytes
#ifndef PUT_UINT32
#define PUT_UINT32(n,b,i)                    \
{                                            \
(b)[(i)+  3] = (uint8_t) ( (n) >> 24 );        \
(b)[(i) + 2] = (uint8_t) ( (n) >> 16 );        \
(b)[(i) + 1] = (uint8_t) ( (n) >>  8 );        \
(b)[(i) + 0] = (uint8_t) ( (n)       );        \
}
#endif

/* collect four bytes into one word: */
#ifndef BYTES_TO_DWORD
#define BYTES_TO_DWORD(strptr)       \
(((uint32_t) *((strptr)+3) << 24) |  \
((uint32_t) *((strptr)+2) << 16) |   \
((uint32_t) *((strptr)+1) <<  8) |   \
((uint32_t) *(strptr)))              \

#endif
/* ROL_RIPEMD160(x, n) cyclically rotates x over n bits to the left */
/* x must be of an unsigned 32 bits type and 0 <= n < 32. */

#define ROL_RIPEMD160(x,n)         (((x) << (n))|((x) >> (32-(n))))
/* the five basic functions F(), G() and H() */

#define F_RIPEMD160(x, y, z)        ((x) ^ (y) ^ (z))
#define G_RIPEMD160(x, y, z)        (((x) & (y)) | (~(x) & (z)))
#define H_RIPEMD160(x, y, z)        (((x) | ~(y)) ^ (z))
#define I_RIPEMD160(x, y, z)        (((x) & (z)) | ((y) & ~(z)))
#define J_RIPEMD160(x, y, z)        ((x) ^ ((y) | ~(z)))

/* the ten basic operations FF_RIPEMD160() through III_RIPEMD160() */
#define FF_RIPEMD160(a, b, c, d, e, x, s){\
(a) += F_RIPEMD160((b), (c), (d)) + (x);\
(a) = ROL_RIPEMD160((a), (s)) + (e);\
(c) = ROL_RIPEMD160((c), 10);\
}

#define GG_RIPEMD160(a, b, c, d, e, x, s) {\
(a) += G_RIPEMD160((b), (c), (d)) + (x) + 0x5a827999UL;\
(a) = ROL_RIPEMD160((a), (s)) + (e);\
(c) = ROL_RIPEMD160((c), 10);\
}

#define HH_RIPEMD160(a, b, c, d, e, x, s) {\
(a) += H_RIPEMD160((b), (c), (d)) + (x) + 0x6ed9eba1UL;\
(a) = ROL_RIPEMD160((a), (s)) + (e);\
(c) = ROL_RIPEMD160((c), 10);\
}

#define II_RIPEMD160(a, b, c, d, e, x, s){\
(a) += I_RIPEMD160((b), (c), (d)) + (x) + 0x8f1bbcdcUL;\
(a) = ROL_RIPEMD160((a), (s)) + (e);\
(c) = ROL_RIPEMD160((c), 10);\
}

#define JJ_RIPEMD160(a, b, c, d, e, x, s){\
(a) += J_RIPEMD160((b), (c), (d)) + (x) + 0xa953fd4eUL;\
(a) = ROL_RIPEMD160((a), (s)) + (e);\
(c) = ROL_RIPEMD160((c), 10);\
}

#define FFF_RIPEMD160(a, b, c, d, e, x, s) {\
(a) += F_RIPEMD160((b), (c), (d)) + (x);\
(a) = ROL_RIPEMD160((a), (s)) + (e);\
(c) = ROL_RIPEMD160((c), 10);\
}

#define GGG_RIPEMD160(a, b, c, d, e, x, s) {\
(a) += G_RIPEMD160((b), (c), (d)) + (x) + 0x7a6d76e9UL;\
(a) = ROL_RIPEMD160((a), (s)) + (e);\
(c) = ROL_RIPEMD160((c), 10);\
}

#define HHH_RIPEMD160(a, b, c, d, e, x, s) {\
(a) += H_RIPEMD160((b), (c), (d)) + (x) + 0x6d703ef3UL;\
(a) = ROL_RIPEMD160((a), (s)) + (e);\
(c) = ROL_RIPEMD160((c), 10);\
}

#define III_RIPEMD160(a, b, c, d, e, x, s) {\
(a) += I_RIPEMD160((b), (c), (d)) + (x) + 0x5c4dd124UL;\
(a) = ROL_RIPEMD160((a), (s)) + (e);\
(c) = ROL_RIPEMD160((c), 10);\
}

#define JJJ_RIPEMD160(a, b, c, d, e, x, s) {\
(a) += J_RIPEMD160((b), (c), (d)) + (x) + 0x50a28be6UL;\
(a) = ROL_RIPEMD160((a), (s)) + (e);\
(c) = ROL_RIPEMD160((c), 10);\
}

/*
 @function:init RIPEMD160_CTX,writing a new message
 @paramter[in]:ctx pointer to RIPEMD160_CTX
 @return: NULL
 @notoce: none
 */
void ripemd160_init(RIPEMD160_CTX *ctx)
{
    if(!ctx)
        return;
    memset((uint8_t *)ctx, 0, sizeof(RIPEMD160_CTX));
    ctx->state[0] = 0x67452301UL;
    ctx->state[1] = 0xefcdab89UL;
    ctx->state[2] = 0x98badcfeUL;
    ctx->state[3] = 0x10325476UL;
    ctx->state[4] = 0xc3d2e1f0UL;
    return;
}

/*
 @function:processing a block message(512 bit)
 @paramter[in]:ctx pointer to RIPEMD160_CTX
 @return: NULL
 @notoce: none
 */
static void compress(RIPEMD160_CTX *ctx)
{
    uint32_t aa = ctx->state[0], bb = ctx->state[1], cc = ctx->state[2],
    dd = ctx->state[3], ee = ctx->state[4];
    uint32_t aaa = ctx->state[0], bbb = ctx->state[1], ccc = ctx->state[2],
    ddd = ctx->state[3], eee = ctx->state[4];
    if(!ctx)
        return;
    /* round 1 */
    FF_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[0], 11);
    FF_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[1], 14);
    FF_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[2], 15);
    FF_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[3], 12);
    FF_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[4], 5);
    FF_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[5], 8);
    FF_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[6], 7);
    FF_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[7], 9);
    FF_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[8], 11);
    FF_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[9], 13);
    FF_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[10], 14);
    FF_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[11], 15);
    FF_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[12], 6);
    FF_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[13], 7);
    FF_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[14], 9);
    FF_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[15], 8);
    
    /* round 2 */
    GG_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[7], 7);
    GG_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[4], 6);
    GG_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[13], 8);
    GG_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[1], 13);
    GG_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[10], 11);
    GG_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[6], 9);
    GG_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[15], 7);
    GG_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[3], 15);
    GG_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[12], 7);
    GG_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[0], 12);
    GG_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[9], 15);
    GG_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[5], 9);
    GG_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[2], 11);
    GG_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[14], 7);
    GG_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[11], 13);
    GG_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[8], 12);
    
    /* round 3 */
    HH_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[3], 11);
    HH_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[10], 13);
    HH_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[14], 6);
    HH_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[4], 7);
    HH_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[9], 14);
    HH_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[15], 9);
    HH_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[8], 13);
    HH_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[1], 15);
    HH_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[2], 14);
    HH_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[7], 8);
    HH_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[0], 13);
    HH_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[6], 6);
    HH_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[13], 5);
    HH_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[11], 12);
    HH_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[5], 7);
    HH_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[12], 5);
    
    /* round 4 */
    II_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[1], 11);
    II_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[9], 12);
    II_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[11], 14);
    II_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[10], 15);
    II_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[0], 14);
    II_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[8], 15);
    II_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[12], 9);
    II_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[4], 8);
    II_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[13], 9);
    II_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[3], 14);
    II_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[7], 5);
    II_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[15], 6);
    II_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[14], 8);
    II_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[5], 6);
    II_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[6], 5);
    II_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[2], 12);
    
    /* round 5 */
    JJ_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[4], 9);
    JJ_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[0], 15);
    JJ_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[5], 5);
    JJ_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[9], 11);
    JJ_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[7], 6);
    JJ_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[12], 8);
    JJ_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[2], 13);
    JJ_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[10], 12);
    JJ_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[14], 5);
    JJ_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[1], 12);
    JJ_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[3], 13);
    JJ_RIPEMD160(aa, bb, cc, dd, ee, ctx->buffer[8], 14);
    JJ_RIPEMD160(ee, aa, bb, cc, dd, ctx->buffer[11], 11);
    JJ_RIPEMD160(dd, ee, aa, bb, cc, ctx->buffer[6], 8);
    JJ_RIPEMD160(cc, dd, ee, aa, bb, ctx->buffer[15], 5);
    JJ_RIPEMD160(bb, cc, dd, ee, aa, ctx->buffer[13], 6);
    
    /* parallel round 1 */
    JJJ_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[5], 8);
    JJJ_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[14], 9);
    JJJ_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[7], 9);
    JJJ_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[0], 11);
    JJJ_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[9], 13);
    JJJ_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[2], 15);
    JJJ_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[11], 15);
    JJJ_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[4], 5);
    JJJ_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[13], 7);
    JJJ_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[6], 7);
    JJJ_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[15], 8);
    JJJ_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[8], 11);
    JJJ_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[1], 14);
    JJJ_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[10], 14);
    JJJ_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[3], 12);
    JJJ_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[12], 6);
    
    /* parallel round 2 */
    III_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[6], 9);
    III_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[11], 13);
    III_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[3], 15);
    III_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[7], 7);
    III_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[0], 12);
    III_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[13], 8);
    III_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[5], 9);
    III_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[10], 11);
    III_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[14], 7);
    III_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[15], 7);
    III_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[8], 12);
    III_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[12], 7);
    III_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[4], 6);
    III_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[9], 15);
    III_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[1], 13);
    III_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[2], 11);
    
    /* parallel round 3 */
    HHH_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[15], 9);
    HHH_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[5], 7);
    HHH_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[1], 15);
    HHH_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[3], 11);
    HHH_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[7], 8);
    HHH_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[14], 6);
    HHH_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[6], 6);
    HHH_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[9], 14);
    HHH_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[11], 12);
    HHH_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[8], 13);
    HHH_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[12], 5);
    HHH_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[2], 14);
    HHH_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[10], 13);
    HHH_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[0], 13);
    HHH_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[4], 7);
    HHH_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[13], 5);
    
    /* parallel round 4 */
    GGG_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[8], 15);
    GGG_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[6], 5);
    GGG_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[4], 8);
    GGG_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[1], 11);
    GGG_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[3], 14);
    GGG_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[11], 14);
    GGG_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[15], 6);
    GGG_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[0], 14);
    GGG_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[5], 6);
    GGG_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[12], 9);
    GGG_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[2], 12);
    GGG_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[13], 9);
    GGG_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[9], 12);
    GGG_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[7], 5);
    GGG_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[10], 15);
    GGG_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[14], 8);
    
    /* parallel round 5 */
    FFF_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[12], 8);
    FFF_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[15], 5);
    FFF_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[10], 12);
    FFF_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[4], 9);
    FFF_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[1], 12);
    FFF_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[5], 5);
    FFF_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[8], 14);
    FFF_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[7], 6);
    FFF_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[6], 8);
    FFF_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[2], 13);
    FFF_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[13], 6);
    FFF_RIPEMD160(aaa, bbb, ccc, ddd, eee, ctx->buffer[14], 5);
    FFF_RIPEMD160(eee, aaa, bbb, ccc, ddd, ctx->buffer[0], 15);
    FFF_RIPEMD160(ddd, eee, aaa, bbb, ccc, ctx->buffer[3], 13);
    FFF_RIPEMD160(ccc, ddd, eee, aaa, bbb, ctx->buffer[9], 11);
    FFF_RIPEMD160(bbb, ccc, ddd, eee, aaa, ctx->buffer[11], 11);
    
    /* combine results */
    ddd += cc + ctx->state[1];               /* final result for MDbuf[0] */
    ctx->state[1] = ctx->state[2] + dd + eee;
    ctx->state[2] = ctx->state[3] + ee + aaa;
    ctx->state[3] = ctx->state[4] + aa + bbb;
    ctx->state[4] = ctx->state[0] + bb + ccc;
    ctx->state[0] = ddd;
    return;
}

/*
 @function:update message Continues an ripemd160 message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:ctx pointer to RIPEMD160_CTX
 @paramter[in]:msg pointer to the message to do ripemd160
 @paramter[in]:msg_len,the byte length of input
 @return:NULL
 @notoce:none
 */
void ripemd160_update(RIPEMD160_CTX *ctx,uint8_t *msg,uint32_t msg_len)
{
    uint32_t fill,left;
    if(!msg_len)
        return;
    if(!ctx || !msg)
        return;
    left = ctx->total[0] & 0x3F;
    fill = 64 - left;
    ctx -> total[0] += msg_len;
    ctx -> total[0] &= 0xFFFFFFFF;
    if(ctx->total[0] < msg_len)
        ctx->total[1]++;
    //memset(ctx->buffer, 0, 64);
    if(left && (msg_len >= fill))
    {
        memcpy((uint8_t *)(ctx -> buffer) + left, msg, fill);
         compress(ctx);
        msg += fill;
        msg_len  -= fill;
        left = 0;
    }
    while(msg_len >= 64)
    {
        //sha256_process(ctx, data);
        compress(ctx);
        msg += 64;
        msg_len  -= 64;
    }
    if(msg_len > 0)
        memcpy((uint8_t *)(ctx -> buffer) + left, msg, msg_len);
    return;
}
/*
 @function: End an ripemd160 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to RIPEMD160_CTX
 @paramter[out]:digest pointer to md5 hash result
 @return:NULL
 @notice:nothing
 */
void ripemd160_final(RIPEMD160_CTX *ctx,uint8_t digest[20])
{
        uint32_t last, padn;
        uint32_t high, low;
        uint8_t msglen[8];
        high = (ctx -> total[0] >> 29)
        | (ctx -> total[1] <<  3);
        low  = (ctx -> total[0] <<  3);
        
        //PUT_UINT32(high, msglen, 0);
        //PUT_UINT32(low,  msglen, 4);
        PUT_UINT32(high, msglen, 4);
        PUT_UINT32(low,  msglen, 0);
        
        last = ctx -> total[0] & 0x3F;
        padn = (last < 56) ? (56 - last) : (120 - last);
        
        ripemd160_update(ctx, (uint8_t *)ripemd160_padding, padn);
        ripemd160_update(ctx, msglen, 8);
        
        PUT_UINT32(ctx -> state[0], digest,  0);
        PUT_UINT32(ctx -> state[1], digest,  4);
        PUT_UINT32(ctx -> state[2], digest,  8);
        PUT_UINT32(ctx -> state[3], digest, 12);
        PUT_UINT32(ctx -> state[4], digest, 16);
        return;
}


/*
 @function:ripemd160 hash
 @paramter[in]:msg pointer to the message to do ripemd160
 @paramter[in]:msg_len,the byte length of input
 @digest[out]:digest piointer to  ripemd160 hash result
 @return:NULL
 @notice:none
 */
void ripemd160_hash(uint8_t *msg,uint32_t msg_len,uint8_t digest[20])
{
    RIPEMD160_CTX ctx;
    if(! msg_len)
        return;
    if(!msg || !digest)
        return;
    ripemd160_init(&ctx);
    ripemd160_update(&ctx,msg,msg_len);
    ripemd160_final(&ctx,digest);
    return;
}


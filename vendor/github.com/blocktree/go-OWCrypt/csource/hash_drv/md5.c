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


#include "md5.h"

#define MD5STR_LEN        32


/* Constants for _MD5Transform routine. */
#define S11_MD5 7
#define S12_MD5 12
#define S13_MD5 17
#define S14_MD5 22
#define S21_MD5 5
#define S22_MD5 9
#define S23_MD5 14
#define S24_MD5 20
#define S31_MD5 4
#define S32_MD5 11
#define S33_MD5 16
#define S34_MD5 23
#define S41_MD5 6
#define S42_MD5 10
#define S43_MD5 15
#define S44_MD5 21

static void _MD5Transform(uint32_t[4], const uint8_t[64]);
//static void _Encode(unsigned char *, unsigned int *, unsigned int);
static void _Encode(uint8_t *output, uint32_t *input, uint32_t len);
static void _Decode(uint32_t *output, const uint8_t *input, uint32_t len);
static unsigned char PADDING_MD5[64] = {
    0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
};

/* F, G, H and I are basic MD5 functions. */
#define F_MD5(x, y, z) (((x) & (y)) | ((~x) & (z)))
#define G_MD5(x, y, z) (((x) & (z)) | ((y) & (~z)))
#define H_MD5(x, y, z) ((x) ^ (y) ^ (z))
#define I_MD5(x, y, z) ((y) ^ ((x) | (~z)))

/* ROTATE_LEFT rotates x left n bits. */
#define ROTATE_LEFT_MD5(x, n) (((x) << (n)) | ((x) >> (32-(n))))
/* FF, GG, HH, and II transformations for rounds 1, 2, 3, and 4.
 Rotation is separate from addition to prevent recomputation. */
#define FF_MD5(a, b, c, d, x, s, ac) {\
(a) += F_MD5 ((b), (c), (d)) + (x) + (unsigned int)(ac);\
(a) = ROTATE_LEFT_MD5 ((a), (s));\
(a) += (b);\
}

#define GG_MD5(a, b, c, d, x, s, ac) {\
(a) += G_MD5 ((b), (c), (d)) + (x) + (unsigned int)(ac);\
(a) = ROTATE_LEFT_MD5 ((a), (s));\
(a) += (b);\
}

#define HH_MD5(a, b, c, d, x, s, ac) {\
(a) += H_MD5 ((b), (c), (d)) + (x) + (unsigned int)(ac);\
(a) = ROTATE_LEFT_MD5 ((a), (s));\
(a) += (b);\
}

#define II_MD5(a, b, c, d, x, s, ac) {\
(a) += I_MD5 ((b), (c), (d)) + (x) + (unsigned int)(ac);\
(a) = ROTATE_LEFT_MD5 ((a), (s));\
(a) += (b);\
}
/*
 @function:init MD5_CTX,writing a new message
 @paramter[in]:ctx pointer to MD5_CTX
 @return: NULL
 @notoce: none
 */
void md5_init(MD5_CTX * ctx)
{
    ctx->count[0] = ctx->count[1] = 0;
    /* Load magic initialization constants. */
    ctx->state[0] = 0x67452301;
    ctx->state[1] = 0xefcdab89;
    ctx->state[2] = 0x98badcfe;
    ctx->state[3] = 0x10325476;
}

/*
 @function:update message Continues an MD5 message-digest operation,
 processing another message block, and updating the context.
 @paramter[in]:ctx pointer to MD5_CTX
 @paramter[in]:msg pointer to the message to do md5
 @paramter[in]:msg_len,the byte length of input
 @return:NULL
 @notoce:none
 */
void md5_update(MD5_CTX * ctx, const uint8_t *msg, const uint32_t msg_len)
{
    uint32_t i, index, partLen;
    if(!ctx || !msg)
        return;
    if(!msg_len)
        return;
    /* Compute number of bytes mod 64 */
    index = (uint32_t) ((ctx->count[0] >> 3) & 0x3F);
    
    /* Update number of bits */
    if ((ctx->count[0] += ((uint32_t) msg_len << 3)) < ((uint32_t) msg_len << 3))
        ctx->count[1]++;
    ctx->count[1] += ((uint32_t) msg_len >> 29);
    
    partLen = 64 - index;
    
    /* Transform as many times as possible. */
    if (msg_len >= partLen) {
        memcpy((void *) &ctx->buffer[index], (void *) msg, partLen);
        _MD5Transform(ctx->state, ctx->buffer);
        
        for (i = partLen; i + 63 < msg_len; i += 64)
            _MD5Transform(ctx->state, &msg[i]);
        index = 0;
    }
    else
        i = 0;
    /* Buffer remaining input */
    memcpy((void *) &ctx->buffer[index], (void *) &msg[i], msg_len - i);
    return;
}

/*
 @function:finalization md5 operation Ends an MD5 message-digest operation, writing the message digest and zeroizing the context
 @paramter[in]:ctx pointer to MD5_CTX
 @paramter[out]:digest pointer to md5 hash result
 @return:NULL
 @notice:nothing
 */
void md5_final(MD5_CTX * ctx,uint8_t digest[16])
{
    uint8_t bits[8];
    uint32_t index, padLen;
    if(!ctx || !digest)
        return;
    /* Save number of bits */
    _Encode(bits, ctx->count, 8);
    /* Pad out to 56 mod 64. */
    index = (uint32_t) ((ctx->count[0] >> 3) & 0x3f);
    padLen = (index < 56) ? (56 - index) : (120 - index);
    md5_update(ctx, PADDING_MD5, padLen);
    
    /* Append length (before padding) */
    md5_update(ctx, bits, 8);
    
    /* Store state in digest */
    _Encode(digest, ctx->state, 16);
    
    /* Zeroize sensitive information.*/
    memset((void *) ctx, 0, sizeof (*ctx));
    return;
}

/* MD5 basic transformation. Transforms state based on block. */
static void _MD5Transform(uint32_t state[4], const uint8_t block[64])
{
    uint32_t a = state[0],
    b = state[1],
    c = state[2],
    d = state[3],
    x[16];
    _Decode(x, block, 64);
    /* Round 1 */
    FF_MD5(a, b, c, d, x[0], S11_MD5, 0xd76aa478);    /* 1 */
    FF_MD5(d, a, b, c, x[1], S12_MD5, 0xe8c7b756);    /* 2 */
    FF_MD5(c, d, a, b, x[2], S13_MD5, 0x242070db);    /* 3 */
    FF_MD5(b, c, d, a, x[3], S14_MD5, 0xc1bdceee);    /* 4 */
    FF_MD5(a, b, c, d, x[4], S11_MD5, 0xf57c0faf);    /* 5 */
    FF_MD5(d, a, b, c, x[5], S12_MD5, 0x4787c62a);    /* 6 */
    FF_MD5(c, d, a, b, x[6], S13_MD5, 0xa8304613);    /* 7 */
    FF_MD5(b, c, d, a, x[7], S14_MD5, 0xfd469501);    /* 8 */
    FF_MD5(a, b, c, d, x[8], S11_MD5, 0x698098d8);    /* 9 */
    FF_MD5(d, a, b, c, x[9], S12_MD5, 0x8b44f7af);    /* 10 */
    FF_MD5(c, d, a, b, x[10], S13_MD5, 0xffff5bb1);    /* 11 */
    FF_MD5(b, c, d, a, x[11], S14_MD5, 0x895cd7be);    /* 12 */
    FF_MD5(a, b, c, d, x[12], S11_MD5, 0x6b901122);    /* 13 */
    FF_MD5(d, a, b, c, x[13], S12_MD5, 0xfd987193);    /* 14 */
    FF_MD5(c, d, a, b, x[14], S13_MD5, 0xa679438e);    /* 15 */
    FF_MD5(b, c, d, a, x[15], S14_MD5, 0x49b40821);    /* 16 */
    
    /* Round 2 */
    GG_MD5(a, b, c, d, x[1], S21_MD5, 0xf61e2562);    /* 17 */
    GG_MD5(d, a, b, c, x[6], S22_MD5, 0xc040b340);    /* 18 */
    GG_MD5(c, d, a, b, x[11], S23_MD5, 0x265e5a51);    /* 19 */
    GG_MD5(b, c, d, a, x[0], S24_MD5, 0xe9b6c7aa);    /* 20 */
    GG_MD5(a, b, c, d, x[5], S21_MD5, 0xd62f105d);    /* 21 */
    GG_MD5(d, a, b, c, x[10], S22_MD5, 0x2441453);    /* 22 */
    GG_MD5(c, d, a, b, x[15], S23_MD5, 0xd8a1e681);    /* 23 */
    GG_MD5(b, c, d, a, x[4], S24_MD5, 0xe7d3fbc8);    /* 24 */
    GG_MD5(a, b, c, d, x[9], S21_MD5, 0x21e1cde6);    /* 25 */
    GG_MD5(d, a, b, c, x[14], S22_MD5, 0xc33707d6);    /* 26 */
    GG_MD5(c, d, a, b, x[3], S23_MD5, 0xf4d50d87);    /* 27 */
    GG_MD5(b, c, d, a, x[8], S24_MD5, 0x455a14ed);    /* 28 */
    GG_MD5(a, b, c, d, x[13], S21_MD5, 0xa9e3e905);    /* 29 */
    GG_MD5(d, a, b, c, x[2], S22_MD5, 0xfcefa3f8);    /* 30 */
    GG_MD5(c, d, a, b, x[7], S23_MD5, 0x676f02d9);    /* 31 */
    GG_MD5(b, c, d, a, x[12], S24_MD5, 0x8d2a4c8a);    /* 32 */
    
    /* Round 3 */
    HH_MD5(a, b, c, d, x[5], S31_MD5, 0xfffa3942);    /* 33 */
    HH_MD5(d, a, b, c, x[8], S32_MD5, 0x8771f681);    /* 34 */
    HH_MD5(c, d, a, b, x[11], S33_MD5, 0x6d9d6122);    /* 35 */
    HH_MD5(b, c, d, a, x[14], S34_MD5, 0xfde5380c);    /* 36 */
    HH_MD5(a, b, c, d, x[1], S31_MD5, 0xa4beea44);    /* 37 */
    HH_MD5(d, a, b, c, x[4], S32_MD5, 0x4bdecfa9);    /* 38 */
    HH_MD5(c, d, a, b, x[7], S33_MD5, 0xf6bb4b60);    /* 39 */
    HH_MD5(b, c, d, a, x[10], S34_MD5, 0xbebfbc70);    /* 40 */
    HH_MD5(a, b, c, d, x[13], S31_MD5, 0x289b7ec6);    /* 41 */
    HH_MD5(d, a, b, c, x[0], S32_MD5, 0xeaa127fa);    /* 42 */
    HH_MD5(c, d, a, b, x[3], S33_MD5, 0xd4ef3085);    /* 43 */
    HH_MD5(b, c, d, a, x[6], S34_MD5, 0x4881d05);    /* 44 */
    HH_MD5(a, b, c, d, x[9], S31_MD5, 0xd9d4d039);    /* 45 */
    HH_MD5(d, a, b, c, x[12], S32_MD5, 0xe6db99e5);    /* 46 */
    HH_MD5(c, d, a, b, x[15], S33_MD5, 0x1fa27cf8);    /* 47 */
    HH_MD5(b, c, d, a, x[2], S34_MD5, 0xc4ac5665);    /* 48 */
    
    /* Round 4 */
    II_MD5(a, b, c, d, x[0], S41_MD5, 0xf4292244);    /* 49 */
    II_MD5(d, a, b, c, x[7], S42_MD5, 0x432aff97);    /* 50 */
    II_MD5(c, d, a, b, x[14], S43_MD5, 0xab9423a7);    /* 51 */
    II_MD5(b, c, d, a, x[5], S44_MD5, 0xfc93a039);    /* 52 */
    II_MD5(a, b, c, d, x[12], S41_MD5, 0x655b59c3);    /* 53 */
    II_MD5(d, a, b, c, x[3], S42_MD5, 0x8f0ccc92);    /* 54 */
    II_MD5(c, d, a, b, x[10], S43_MD5, 0xffeff47d);    /* 55 */
    II_MD5(b, c, d, a, x[1], S44_MD5, 0x85845dd1);    /* 56 */
    II_MD5(a, b, c, d, x[8], S41_MD5, 0x6fa87e4f);    /* 57 */
    II_MD5(d, a, b, c, x[15], S42_MD5, 0xfe2ce6e0);    /* 58 */
    II_MD5(c, d, a, b, x[6], S43_MD5, 0xa3014314);    /* 59 */
    II_MD5(b, c, d, a, x[13], S44_MD5, 0x4e0811a1);    /* 60 */
    II_MD5(a, b, c, d, x[4], S41_MD5, 0xf7537e82);    /* 61 */
    II_MD5(d, a, b, c, x[11], S42_MD5, 0xbd3af235);    /* 62 */
    II_MD5(c, d, a, b, x[2], S43_MD5, 0x2ad7d2bb);    /* 63 */
    II_MD5(b, c, d, a, x[9], S44_MD5, 0xeb86d391);    /* 64 */
    state[0] += a;
    state[1] += b;
    state[2] += c;
    state[3] += d;
    /* Zeroize sensitive information. */
    memset((void *) x, 0, sizeof (x));
}


/* Encodes input (unsigned int) into output (unsigned char). Assumes len is a multiple of 4. */
static void _Encode(uint8_t *output,  uint32_t *input, uint32_t len)
{
    uint32_t i, j;
    if(!output || !input)
        return;
    if(!len)
        return;
    for (i = 0, j = 0; j < len; i++, j += 4) {
        output[j] = (uint8_t) (input[i] & 0xff);
        output[j + 1] = (uint8_t ) ((input[i] >> 8) & 0xff);
        output[j + 2] = (uint8_t ) ((input[i] >> 16) & 0xff);
        output[j + 3] = (uint8_t ) ((input[i] >> 24) & 0xff);
    }
}

/* Decodes input (unsigned char) into output (unsigned int). Assumes len is a multiple of 4.*/
static void _Decode(uint32_t *output, const uint8_t *input, uint32_t len)
{
   uint32_t i, j;
    if(!output || !input)
        return;
    if(!len)
        return;
    for (i = 0, j = 0; j < len; i++, j += 4) {
        output[i] = ((uint32_t) input[j]) | (((uint32_t) input[j + 1]) << 8) |
        (((uint32_t ) input[j + 2]) << 16) | (((uint32_t) input[j + 3]) << 24);
    }
    return;
}

/*
 @function:md5 hash
 @paramter[in]:msg pointer to the message to do md5
 @paramter[in]:msg_len,the byte length of input
 @digest[out]:digest pointer to md5 hash result
 @return:NULL
 @notice:none
 */
void  md5_hash(const uint8_t *msg,uint32_t msg_len,uint8_t digest[16])
{
     MD5_CTX ctx;
    if(!msg || !digest)
        return;
    if(!msg_len)
        return;
     md5_init(&ctx);
     md5_update(&ctx, msg, msg_len);
     md5_final(&ctx,digest);
    return;
}

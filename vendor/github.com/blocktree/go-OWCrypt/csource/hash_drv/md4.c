
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

#include "md4.h"
/* Constants for MD4Transform routine.  */
#define S11_MD4 3
#define S12_MD4 7
#define S13_MD4 11
#define S14_MD4 19
#define S21_MD4 3
#define S22_MD4 5
#define S23_MD4 9
#define S24_MD4 13
#define S31_MD4 3
#define S32_MD4 9
#define S33_MD4 11
#define S34_MD4 15

static void md4_transform (uint32_t [4], const uint8_t [64]);
static void Encode (uint8_t *,  uint32_t *, uint32_t);
static void Decode (uint32_t *, const uint8_t *, uint32_t);

static unsigned char PADDING_MD4[64] = {
    0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
};

/* F, G and H are basic MD4 functions. */
#define F_MD4(x, y, z) (((x) & (y)) | ((~x) & (z)))
#define G_MD4(x, y, z) (((x) & (y)) | ((x) & (z)) | ((y) & (z)))
#define H_MD4(x, y, z) ((x) ^ (y) ^ (z))

/* ROTATE_LEFT rotates x left n bits. */
#define ROTATE_LEFT_MD4(x, n) (((x) << (n)) | ((x) >> (32-(n))))

/* FF, GG and HH are transformations for rounds 1, 2 and 3 */
/* Rotation is separate from addition to prevent recomputation */
#define FF_MD4(a, b, c, d, x, s) {(a) += F_MD4 ((b), (c), (d)) + (x); (a) = ROTATE_LEFT_MD4 ((a), (s));}

#define GG_MD4(a, b, c, d, x, s) {(a) += G_MD4 ((b), (c), (d)) + (x) + (uint32_t)0x5a827999; (a) = ROTATE_LEFT_MD4 ((a), (s));}

#define HH_MD4(a, b, c, d, x, s) {(a) += H_MD4 ((b), (c), (d)) + (x) + (uint32_t)0x6ed9eba1; (a) = ROTATE_LEFT_MD4 ((a), (s));}


/*
 @function:MD4 initialization.Begins an MD4 operation, writing a new context
 @paramter[in]:ctx pointer to MD4_CTX
 @return: NULL
 @notoce: none
 */
void md4_init (MD4_CTX *ctx)
{
    ctx->count[0] = ctx->count[1] = 0;
    
    /* Load magic initialization constants.*/
    ctx->state[0] = 0x67452301;
    ctx->state[1] = 0xefcdab89;
    ctx->state[2] = 0x98badcfe;
    ctx->state[3] = 0x10325476;
}


/*
 @function:MD4 block update operation. Continues an MD4 message-digest operation, processing another message block, and updating the context.
 @paramter[in]:ctx pointer to MD4_CTX
 @paramter[in]:msg pointer to the message to do MD4
 @paramter[in]:msg_len,the byte length of input
 @return:NULL
 @notoce:none
 */
void md4_update (MD4_CTX *ctx, const uint8_t *msg, uint32_t msg_len)
{
    uint32_t i, index, partLen;
    
    /* Compute number of bytes mod 64 */
    index = (uint32_t)((ctx->count[0] >> 3) & 0x3F);
    
    /* Update number of bits */
    if ((ctx->count[0] += ((uint32_t)msg_len << 3))< ((uint32_t)msg_len << 3))
        ctx->count[1]++;
    
    ctx->count[1] += (msg_len >> 29);
    
    partLen = 64 - index;
    
    /* Transform as many times as possible.*/
    if (msg_len >= partLen)
    {
        memcpy((void *)&ctx->buffer[index], (void *)msg, partLen);
        md4_transform (ctx->state, ctx->buffer);
        
        for (i = partLen; i + 63 < msg_len; i += 64)
            md4_transform (ctx->state, &msg[i]);
        
        index = 0;
    }
    else
        i = 0;
    /* Buffer remaining input */
    memcpy ((void *)&ctx->buffer[index], (void *)&msg[i], msg_len-i);
}


/*
 @function:MD4 finalization. Ends an MD4 message-digest operation, writing the the message digest and zeroizing the context
 @paramter[out]:digest pointer to MD4 hash result
 @paramter[in]:ctx pointer to MD4_CTX
 @return:NULL
 @notice:nothing
 */
void md4_final (MD4_CTX *ctx,uint8_t digest[16])
{
    unsigned char bits[8];
    unsigned int index, padLen;
    /* Save number of bits */
    Encode (bits, ctx->count, 8);
    /* Pad out to 56 mod 64.*/
    index = (unsigned int)((ctx->count[0] >> 3) & 0x3f);
    padLen = (index < 56) ? (56 - index) : (120 - index);
    md4_update (ctx, PADDING_MD4, padLen);
    /* Append length (before padding) */
    md4_update (ctx, bits, 8);
    
    /* Store state in digest */
    Encode (digest, ctx->state, 16);
    
    /* Zeroize sensitive information.*/
    memset (ctx, 0, sizeof (*ctx));
}


/* MD4 basic transformation. Transforms state based on block. */
static void md4_transform (uint32_t state[4], const uint8_t block[64])
{
    uint32_t a = state[0], b = state[1], c = state[2], d = state[3], x[16];
    
    Decode (x, block, 64);
    
    /* Round 1 */
    FF_MD4 (a, b, c, d, x[ 0], S11_MD4);                 /* 1 */
    FF_MD4 (d, a, b, c, x[ 1], S12_MD4);                 /* 2 */
    FF_MD4 (c, d, a, b, x[ 2], S13_MD4);                 /* 3 */
    FF_MD4 (b, c, d, a, x[ 3], S14_MD4);                 /* 4 */
    FF_MD4 (a, b, c, d, x[ 4], S11_MD4);                 /* 5 */
    FF_MD4 (d, a, b, c, x[ 5], S12_MD4);                 /* 6 */
    FF_MD4 (c, d, a, b, x[ 6], S13_MD4);                 /* 7 */
    FF_MD4 (b, c, d, a, x[ 7], S14_MD4);                 /* 8 */
    FF_MD4 (a, b, c, d, x[ 8], S11_MD4);                 /* 9 */
    FF_MD4 (d, a, b, c, x[ 9], S12_MD4);                 /* 10 */
    FF_MD4 (c, d, a, b, x[10], S13_MD4);             /* 11 */
    FF_MD4 (b, c, d, a, x[11], S14_MD4);             /* 12 */
    FF_MD4 (a, b, c, d, x[12], S11_MD4);             /* 13 */
    FF_MD4 (d, a, b, c, x[13], S12_MD4);             /* 14 */
    FF_MD4 (c, d, a, b, x[14], S13_MD4);             /* 15 */
    FF_MD4 (b, c, d, a, x[15], S14_MD4);             /* 16 */
    
    /* Round 2 */
    GG_MD4 (a, b, c, d, x[ 0], S21_MD4);             /* 17 */
    GG_MD4 (d, a, b, c, x[ 4], S22_MD4);             /* 18 */
    GG_MD4 (c, d, a, b, x[ 8], S23_MD4);             /* 19 */
    GG_MD4 (b, c, d, a, x[12], S24_MD4);             /* 20 */
    GG_MD4 (a, b, c, d, x[ 1], S21_MD4);             /* 21 */
    GG_MD4 (d, a, b, c, x[ 5], S22_MD4);             /* 22 */
    GG_MD4 (c, d, a, b, x[ 9], S23_MD4);             /* 23 */
    GG_MD4 (b, c, d, a, x[13], S24_MD4);             /* 24 */
    GG_MD4 (a, b, c, d, x[ 2], S21_MD4);             /* 25 */
    GG_MD4 (d, a, b, c, x[ 6], S22_MD4);             /* 26 */
    GG_MD4 (c, d, a, b, x[10], S23_MD4);             /* 27 */
    GG_MD4 (b, c, d, a, x[14], S24_MD4);             /* 28 */
    GG_MD4 (a, b, c, d, x[ 3], S21_MD4);             /* 29 */
    GG_MD4 (d, a, b, c, x[ 7], S22_MD4);             /* 30 */
    GG_MD4 (c, d, a, b, x[11], S23_MD4);             /* 31 */
    GG_MD4 (b, c, d, a, x[15], S24_MD4);             /* 32 */
    
    /* Round 3 */
    HH_MD4 (a, b, c, d, x[ 0], S31_MD4);                /* 33 */
    HH_MD4 (d, a, b, c, x[ 8], S32_MD4);             /* 34 */
    HH_MD4 (c, d, a, b, x[ 4], S33_MD4);             /* 35 */
    HH_MD4 (b, c, d, a, x[12], S34_MD4);             /* 36 */
    HH_MD4 (a, b, c, d, x[ 2], S31_MD4);             /* 37 */
    HH_MD4 (d, a, b, c, x[10], S32_MD4);             /* 38 */
    HH_MD4 (c, d, a, b, x[ 6], S33_MD4);             /* 39 */
    HH_MD4 (b, c, d, a, x[14], S34_MD4);             /* 40 */
    HH_MD4 (a, b, c, d, x[ 1], S31_MD4);             /* 41 */
    HH_MD4 (d, a, b, c, x[ 9], S32_MD4);             /* 42 */
    HH_MD4 (c, d, a, b, x[ 5], S33_MD4);             /* 43 */
    HH_MD4 (b, c, d, a, x[13], S34_MD4);             /* 44 */
    HH_MD4 (a, b, c, d, x[ 3], S31_MD4);             /* 45 */
    HH_MD4 (d, a, b, c, x[11], S32_MD4);             /* 46 */
    HH_MD4 (c, d, a, b, x[ 7], S33_MD4);             /* 47 */
    HH_MD4 (b, c, d, a, x[15], S34_MD4);            /* 48 */
    
    state[0] += a;
    state[1] += b;
    state[2] += c;
    state[3] += d;
    
    /* Zeroize sensitive information.*/
    memset (x, 0, sizeof (x));
}


/* Encodes input (UINT4) into output (unsigned char). Assumes len is a multiple of 4. */
static void Encode (unsigned char *output, uint32_t *input, unsigned int len)
{
    unsigned int i, j;
    
    for (i = 0, j = 0; j < len; i++, j += 4)
    {
        output[j] = (unsigned char)(input[i] & 0xff);
        output[j+1] = (unsigned char)((input[i] >> 8) & 0xff);
        output[j+2] = (unsigned char)((input[i] >> 16) & 0xff);
        output[j+3] = (unsigned char)((input[i] >> 24) & 0xff);
    }
}


/* Decodes input (unsigned char) into output (UINT4). Assumes len is a multiple of 4. */
static void Decode (uint32_t *output, const uint8_t *input, uint32_t len)
{
    unsigned int i, j;
    for (i = 0, j = 0; j < len; i++, j += 4)
        output[i] = ((uint32_t)input[j]) | (((uint32_t)input[j+1]) << 8) | (((uint32_t)input[j+2]) << 16) | (((uint32_t)input[j+3]) << 24);
}

/*
 @function:MD4 hash
 @paramter[in]:msg pointer to the message to do MD4
 @paramter[in]:msg_len,the byte length of input
 @digest[out]:digest pointer to MD4 hash result
 @return:NULL
 @notice:none
 */
void  md4_hash(const uint8_t *msg,uint32_t msg_len,uint8_t digest[16])
{
    MD4_CTX ctx;
    if(!msg || !digest)
        return;
    if(!msg_len)
        return;
    md4_init(&ctx);
    md4_update(&ctx, msg, msg_len);
    md4_final(&ctx,digest);
    return;
}



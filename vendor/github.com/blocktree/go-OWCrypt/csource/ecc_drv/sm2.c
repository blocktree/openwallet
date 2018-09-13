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

#include "sm2.h"

/////////////////////////////////////////////////////////////////////////////////////SM2 STANDARD CURVE///////////////////////////////////////////////////////////////////////////////////////////////////////
static const uint8_t curve_sm2_std_p[32]   = {0xFF,0xFF,0xFF,0xFE,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0x00,0x00,0x00,0x00,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF};
static const uint8_t curve_sm2_std_a[32]   = {0xFF,0xFF,0xFF,0xFE,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0x00,0x00,0x00,0x00,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFC};
static const uint8_t curve_sm2_std_b[32]   = {0x28,0xE9,0xFA,0x9E,0x9D,0x9F,0x5E,0x34,0x4D,0x5A,0x9E,0x4B,0xCF,0x65,0x09,0xA7,0xF3,0x97,0x89,0xF5,0x15,0xAB,0x8F,0x92,0xDD,0xBC,0xBD,0x41,0x4D,0x94,0x0E,0x93};
static const uint8_t curve_sm2_std_x[32]   = {0x32,0xC4,0xAE,0x2C,0x1F,0x19,0x81,0x19,0x5F,0x99,0x04,0x46,0x6A,0x39,0xC9,0x94,0x8F,0xE3,0x0B,0xBF,0xF2,0x66,0x0B,0xE1,0x71,0x5A,0x45,0x89,0x33,0x4C,0x74,0xC7};
static const uint8_t curve_sm2_std_y[32]   = {0xBC,0x37,0x36,0xA2,0xF4,0xF6,0x77,0x9C,0x59,0xBD,0xCE,0xE3,0x6B,0x69,0x21,0x53,0xD0,0xA9,0x87,0x7C,0xC6,0x2A,0x47,0x40,0x02,0xDF,0x32,0xE5,0x21,0x39,0xF0,0xA0};
static const uint8_t curve_sm2_std_n[32]   = {0xFF,0xFF,0xFF,0xFE,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0x72,0x03,0xDF,0x6B,0x21,0xC6,0x05,0x2B,0x53,0xBB,0xF4,0x09,0x39,0xD5,0x41,0x23};

void sm2_std_get_order(uint8_t *order)
{
    memcpy(order, curve_sm2_std_n, ECC_LEN);
}

uint16_t sm2_std_genPubkey(uint8_t *prikey, uint8_t *pubkey)
{
    ECC_CURVE_PARAM *curveParam = NULL;
    ECC_POINT *point = NULL;
    
    curveParam = calloc(1, sizeof(ECC_CURVE_PARAM));
    point = calloc(1, sizeof(ECC_POINT));
    
    curveParam -> p = (uint8_t *)curve_sm2_std_p;
    curveParam -> a = (uint8_t *)curve_sm2_std_a;
    curveParam -> b = (uint8_t *)curve_sm2_std_b;
    curveParam -> x = (uint8_t *)curve_sm2_std_x;
    curveParam -> y = (uint8_t *)curve_sm2_std_y;
    curveParam -> n = (uint8_t *)curve_sm2_std_n;
    
    if(!is_prikey_legal(curveParam, prikey))
    {
        free(curveParam);
        free(point);
        return ECC_PRIKEY_ILLEGAL;
    }
    
    memcpy(point -> x, curveParam -> x, ECC_LEN);
    memcpy(point -> y, curveParam -> y, ECC_LEN);
    
    if(point_mul(curveParam, point, prikey, point))
    {
        free(curveParam);
        free(point);
        return ECC_PRIKEY_ILLEGAL;
    }
    
    memcpy(pubkey, point -> x, ECC_LEN);
    memcpy(pubkey + ECC_LEN, point -> y, ECC_LEN);
    
    free(curveParam);
    free(point);
    return SUCCESS;
}


//mode = 0: sign   -> key = prikey
//mode = 1: verify -> key = pubkey
void sm2_get_e(ECC_CURVE_PARAM *curveParam, uint8_t *ID, uint16_t IDlen, uint8_t *key, uint8_t *message, uint16_t message_len, uint8_t mode, uint8_t *e)
{
    uint8_t ENTLA[2] = {0};
    uint8_t *pubkey = NULL;
    SM3_CTX *sm3_ctx = NULL;
    
    ENTLA[0] = ((IDlen * 8) >> 8) & 0xFF;
    ENTLA[1] = (IDlen * 8) & 0xFF;
    
    
    sm3_ctx = calloc(1, sizeof(SM3_CTX));
    
    
    if(mode) // verify
    {
        pubkey = key;
    }
    else // sign
    {
        pubkey = calloc(ECC_LEN * 2, sizeof(uint8_t));
        sm2_std_genPubkey(key, pubkey);
    }
    
    sm3_init(sm3_ctx);
    sm3_update(sm3_ctx, ENTLA, 2);
    sm3_update(sm3_ctx, ID, IDlen);
    sm3_update(sm3_ctx, curveParam -> a, ECC_LEN);
    sm3_update(sm3_ctx, curveParam -> b, ECC_LEN);
    sm3_update(sm3_ctx, curveParam -> x, ECC_LEN);
    sm3_update(sm3_ctx, curveParam -> y, ECC_LEN);
    sm3_update(sm3_ctx, pubkey, ECC_LEN * 2);
    sm3_final(sm3_ctx, e);
    
    sm3_init(sm3_ctx);
    sm3_update(sm3_ctx, e, ECC_LEN);
    sm3_update(sm3_ctx, message, message_len);
    sm3_final(sm3_ctx, e);
    
    free(sm3_ctx);
    if(!mode)
        free(pubkey);
}

uint16_t sm2_std_sign(uint8_t *prikey, uint8_t *ID, uint16_t IDlen, uint8_t *message, uint16_t message_len,uint8_t *rand,uint8_t hash_flag, uint8_t *sig)
{
    ECC_CURVE_PARAM *curveParam = NULL;
    ECC_POINT *point = NULL;
    uint8_t *e = NULL, *k = NULL, *tmp = NULL;
    
    curveParam = calloc(1, sizeof(ECC_CURVE_PARAM));
    curveParam -> p = (uint8_t *)curve_sm2_std_p;
    curveParam -> a = (uint8_t *)curve_sm2_std_a;
    curveParam -> b = (uint8_t *)curve_sm2_std_b;
    curveParam -> x = (uint8_t *)curve_sm2_std_x;
    curveParam -> y = (uint8_t *)curve_sm2_std_y;
    curveParam -> n = (uint8_t *)curve_sm2_std_n;
    if(!is_prikey_legal(curveParam, prikey))
    {
        free(curveParam);
        return ECC_PRIKEY_ILLEGAL;
    }
    point = calloc(1, sizeof(ECC_POINT));
    e = calloc(ECC_LEN, sizeof(uint8_t));
    k = calloc(ECC_LEN, sizeof(uint8_t));
    tmp = calloc(ECC_LEN, sizeof(uint8_t));
    if(!hash_flag)//需要内部计算哈希值
    {
        sm2_get_e(curveParam, ID, IDlen, prikey, message, message_len, 0, e);
    }
    else //外部已经计算哈希值
    {
        if(message_len != ECC_LEN)
        {
            return HASH_LENGTH_ERROR;
        }
        else
        {
            memcpy(e, message, message_len);
        }
    }
    while(1)
    {
        if(rand==NULL) //内部产生随机数
        {
            bigrand_get_rand_range(k, curveParam -> n, ECC_LEN);
        }
        else //外部传入随机数
        {
            memcpy(k,rand,ECC_LEN);
        }
        memcpy(point -> x, curveParam -> x, ECC_LEN);
        memcpy(point -> y, curveParam -> y, ECC_LEN);
        point_mul(curveParam, point, k, point);
        bignum_mod_add(e, point -> x, curveParam -> n, sig);
        if(is_all_zero(sig, ECC_LEN))
            continue;
        bignum_mod_add(sig, k, curveParam -> n, tmp);
        if(is_all_zero(tmp, ECC_LEN))
            continue;
        memcpy(tmp, prikey, ECC_LEN);
        bignum_add_by_1(tmp);
        bignum_mod_inv(tmp, curveParam -> n, tmp);
        bignum_mod_mul(sig, prikey, curveParam -> n, point -> y);
        bignum_mod_sub(k, point -> y, curveParam -> n, k);
        
        bignum_mod_mul(tmp, k, curveParam -> n, sig + ECC_LEN);
        
        if(is_all_zero(sig + ECC_LEN, ECC_LEN))
            continue;
        else
            break;
    }
    free(curveParam);
    free(point);
    free(e);
    free(k);
    free(tmp);
    return SUCCESS;
}

uint16_t sm2_std_verify(uint8_t *pubkey, uint8_t *ID, uint16_t IDlen, uint8_t *message, uint16_t message_len, uint8_t hash_flag,uint8_t *sig)
{
    ECC_CURVE_PARAM *curveParam = NULL;
    ECC_POINT *point1 = NULL, *point2 = NULL;
    uint8_t *e = NULL, *tmp = NULL;
    
    curveParam = calloc(1, sizeof(ECC_CURVE_PARAM));
    point1 = calloc(1, sizeof(ECC_POINT));
    
    curveParam -> p = (uint8_t *)curve_sm2_std_p;
    curveParam -> a = (uint8_t *)curve_sm2_std_a;
    curveParam -> b = (uint8_t *)curve_sm2_std_b;
    curveParam -> x = (uint8_t *)curve_sm2_std_x;
    curveParam -> y = (uint8_t *)curve_sm2_std_y;
    curveParam -> n = (uint8_t *)curve_sm2_std_n;
    
    memcpy(point1 -> x, pubkey, ECC_LEN);
    memcpy(point1 -> y, pubkey + ECC_LEN, ECC_LEN);
    
    if(!is_pubkey_legal(curveParam, point1))
    {
        free(curveParam);
        free(point1);
        return ECC_PUBKEY_ILLEGAL;
    }
    
    if(is_all_zero(sig, ECC_LEN) || memcmp(sig, curveParam -> n, ECC_LEN) >= 0 || is_all_zero(sig + ECC_LEN, ECC_LEN) || memcmp(sig + ECC_LEN, curveParam -> n, ECC_LEN) >= 0)
    {
        free(curveParam);
        free(point1);
        return FAILURE;
    }

    tmp = calloc(ECC_LEN, sizeof(uint8_t));
    bignum_mod_add(sig, sig + ECC_LEN, curveParam -> n, tmp);
    
    if(is_all_zero(tmp, ECC_LEN))
    {
        free(curveParam);
        free(point1);
        free(tmp);
    }
    
    e = calloc(ECC_LEN, sizeof(uint8_t));
    if(!hash_flag)//需要内部计算哈希值
    {
       sm2_get_e(curveParam, ID, IDlen, pubkey, message, message_len, 1, e);
    }
    else//外部已经计算哈希值
    {
        if(message_len != ECC_LEN)
        {
            return HASH_LENGTH_ERROR;
        }
        memcpy(e,message,message_len);
    }
    
    
    point2 = calloc(1, sizeof(ECC_POINT));
    memcpy(point2 -> x, curveParam -> x, ECC_LEN);
    memcpy(point2 -> y, curveParam -> y, ECC_LEN);
    
    point_mul(curveParam, point1, tmp, point1);
    point_mul(curveParam, point2, sig + ECC_LEN, point2);
    point_add(curveParam, point1, point2, point1);
    
    bignum_mod_add(e, point1 -> x, curveParam -> n, tmp);
    
    if(memcmp(tmp, sig, ECC_LEN))
    {
        free(curveParam);
        free(point1);
        free(point2);
        free(e);
        free(tmp);
        return FAILURE;
    }
    
    free(curveParam);
    free(point1);
    free(point2);
    free(e);
    free(tmp);
    return SUCCESS;
}

static void sm2Kdf(uint8_t *Z, uint16_t Zlen, int klen_bit, uint8_t *key)
{
#define V 256
    uint32_t hlen1, i;
    uint8_t generater[4] = {0};
    SM3_CTX *sm3_ctx = NULL;
    uint8_t *tmp = NULL;
    
    if(klen_bit % V == 0)
        hlen1 = klen_bit / V;
    else
        hlen1 = klen_bit / V + 1;
    
    sm3_ctx = calloc(1, sizeof(SM3_CTX));
    tmp = calloc(SM3_DIGEST_LENGTH, sizeof(uint8_t));
    
    for(i=1;i<=hlen1;i++)
    {
        generater[0] = (i >> 24) & 0xff;
        generater[1] = (i >> 16) & 0xff;
        generater[2] = (i >> 8) & 0xff;
        generater[3] = (i & 0xff);
        
        sm3_init(sm3_ctx);
        sm3_update(sm3_ctx, Z, Zlen);
        sm3_update(sm3_ctx, generater, 4);
        sm3_final(sm3_ctx, tmp);
        
        if(klen_bit >= V)
            memcpy(key + (i - 1) * SM3_DIGEST_LENGTH, tmp, SM3_DIGEST_LENGTH);
        else
            memcpy(key + (i - 1) * SM3_DIGEST_LENGTH,  tmp, klen_bit / 8);
        
        klen_bit -= V;
    }
    
    free(sm3_ctx);
    free(tmp);
#undef V
}



uint16_t sm2_std_enc(uint8_t *pubkey, uint8_t *plain, uint16_t plain_len, uint8_t *cipher, uint16_t *cipher_len)
{
    ECC_CURVE_PARAM *curveParam = NULL;
    ECC_POINT *point = NULL;
    SM3_CTX *sm3_ctx = NULL;
    uint8_t *k = NULL, *tmp = NULL;
    uint16_t i = 0;
    
    curveParam = calloc(1, sizeof(ECC_CURVE_PARAM));
    point = calloc(1, sizeof(ECC_POINT));
    
    curveParam -> p = (uint8_t *)curve_sm2_std_p;
    curveParam -> a = (uint8_t *)curve_sm2_std_a;
    curveParam -> b = (uint8_t *)curve_sm2_std_b;
    curveParam -> x = (uint8_t *)curve_sm2_std_x;
    curveParam -> y = (uint8_t *)curve_sm2_std_y;
    curveParam -> n = (uint8_t *)curve_sm2_std_n;
    
    memcpy(point -> x, pubkey, ECC_LEN);
    memcpy(point -> y, pubkey + ECC_LEN, ECC_LEN);
    
    if(!is_pubkey_legal(curveParam, point))
    {
        free(curveParam);
        free(point);
        return ECC_PUBKEY_ILLEGAL;
    }
    
    *cipher = 0x04;
    
    tmp = calloc(ECC_LEN * 2, sizeof(uint8_t));
    
    while(1)
    {
        k = calloc(ECC_LEN, sizeof(uint8_t));
        bigrand_get_rand_range(k, curveParam -> n, ECC_LEN);
        
        point_mul(curveParam, point, k, point);
        
        memcpy(tmp, point -> x, ECC_LEN);
        memcpy(tmp + ECC_LEN, point -> y, ECC_LEN);
        
        sm2Kdf(tmp, ECC_LEN * 2, plain_len * 8, cipher + 97);
        
        if(is_all_zero(cipher + 97, plain_len))
            continue;
        break;
    }
    
    for(i = 0; i < plain_len; i ++)
        *(cipher + 97 + i) ^= *(plain + i);
    
    sm3_ctx = calloc(1, sizeof(SM3_CTX));
    
    sm3_init(sm3_ctx);
    sm3_update(sm3_ctx, tmp, ECC_LEN);
    sm3_update(sm3_ctx, plain, plain_len);
    sm3_update(sm3_ctx, tmp + ECC_LEN, ECC_LEN);
    sm3_final(sm3_ctx, cipher + 65);
    
    memcpy(point -> x, curveParam -> x, ECC_LEN);
    memcpy(point -> y, curveParam -> y, ECC_LEN);
    
    point_mul(curveParam, point, k, point);
    
    memcpy(cipher + 1, point -> x, ECC_LEN);
    memcpy(cipher + 33, point -> y, ECC_LEN);
    
    free(curveParam);
    free(point);
    free(sm3_ctx);
    free(tmp);
    free(k);

    *cipher_len = plain_len + 97;
    return SUCCESS;
}

uint16_t sm2_std_dec(uint8_t *prikey, uint8_t *cipher, uint16_t cipher_len, uint8_t *plain, uint16_t *plain_len)
{
    ECC_CURVE_PARAM *curveParam = NULL;
    ECC_POINT *point = NULL;
    SM3_CTX *sm3_ctx = NULL;
    uint8_t *tmp = NULL;
    uint16_t i = 0;
    
    if(*cipher != 0x04)
        return FAILURE;
    
    curveParam = calloc(1, sizeof(ECC_CURVE_PARAM));
    
    curveParam -> p = (uint8_t *)curve_sm2_std_p;
    curveParam -> a = (uint8_t *)curve_sm2_std_a;
    curveParam -> b = (uint8_t *)curve_sm2_std_b;
    curveParam -> x = (uint8_t *)curve_sm2_std_x;
    curveParam -> y = (uint8_t *)curve_sm2_std_y;
    curveParam -> n = (uint8_t *)curve_sm2_std_n;
    
    if(!is_prikey_legal(curveParam, prikey))
    {
        free(curveParam);
        return ECC_PRIKEY_ILLEGAL;
    }

    point = calloc(1, sizeof(ECC_POINT));
    memcpy(point -> x, cipher + 1, ECC_LEN);
    memcpy(point -> y, cipher + 33, ECC_LEN);
    
    if(!is_pubkey_legal(curveParam, point))
    {
        free(curveParam);
        free(point);
        return FAILURE;
    }
    
    point_mul(curveParam, point, prikey, point);
    tmp = calloc(ECC_LEN * 2, sizeof(uint8_t));
    
    memcpy(tmp, point -> x, ECC_LEN);
    memcpy(tmp + ECC_LEN, point -> y, ECC_LEN);
    
    sm2Kdf(tmp, ECC_LEN * 2, (cipher_len - 97) * 8, plain);
    
    if(is_all_zero(plain, cipher_len - 97))
    {
        free(curveParam);
        free(point);
        free(tmp);
        return FAILURE;
    }
    
    for(i = 0; i < cipher_len - 97; i ++)
    {
        *(plain + i) ^= *(cipher + 97 + i);
    }
    
    sm3_ctx = calloc(1, sizeof(SM3_CTX));
    
    sm3_init(sm3_ctx);
    sm3_update(sm3_ctx, tmp, ECC_LEN);
    sm3_update(sm3_ctx, plain, cipher_len - 97);
    sm3_update(sm3_ctx, tmp + ECC_LEN, ECC_LEN);
    sm3_final(sm3_ctx, tmp);
    
    if(memcmp(tmp, cipher + 65, ECC_LEN))
    {
        free(curveParam);
        free(point);
        free(tmp);
        free(sm3_ctx);
        return FAILURE;
    }
    
    free(curveParam);
    free(point);
    free(tmp);
    free(sm3_ctx);
    
    *plain_len = cipher_len - 97;
    return SUCCESS;
}

/////////////////////////////////////////////////////////////////////密钥协商///////////////////////////////////////////////////////////////
void sm2_ka_get_Z(uint8_t *ID, uint16_t IDlen, uint8_t *pubkey, uint8_t *Z)
{
    SM3_CTX *sm3_ctx = NULL;
    uint8_t ENTL[2] = {0};
    
    ECC_CURVE_PARAM *curveParam = NULL;
    curveParam = calloc(1, sizeof(ECC_CURVE_PARAM));
    
    curveParam -> p = (uint8_t *)curve_sm2_std_p;
    curveParam -> a = (uint8_t *)curve_sm2_std_a;
    curveParam -> b = (uint8_t *)curve_sm2_std_b;
    curveParam -> x = (uint8_t *)curve_sm2_std_x;
    curveParam -> y = (uint8_t *)curve_sm2_std_y;
    curveParam -> n = (uint8_t *)curve_sm2_std_n;
    
    ENTL[0] = ((IDlen * 8) >> 8) & 0xFF;
    ENTL[1] = (IDlen * 8) & 0xFF;
    
    sm3_ctx = calloc(1, sizeof(SM3_CTX));
    
    sm3_init(sm3_ctx);
    sm3_update(sm3_ctx, ENTL, 2);
    sm3_update(sm3_ctx, ID, IDlen);
    sm3_update(sm3_ctx, curveParam -> a, ECC_LEN);
    sm3_update(sm3_ctx, curveParam -> b, ECC_LEN);
    sm3_update(sm3_ctx, curveParam -> x, ECC_LEN);
    sm3_update(sm3_ctx, curveParam -> y, ECC_LEN);
    sm3_update(sm3_ctx, pubkey, ECC_LEN * 2);
    sm3_final(sm3_ctx, Z);
    
    free(sm3_ctx);
    free(curveParam);
}

void sm2_ka_KDF(uint8_t *x, uint8_t *y, uint8_t *Zinitiator, uint8_t *Zresponder, uint16_t klen_bit, uint8_t*key)
{
#define V 256
    uint32_t hlen1, i;
    uint8_t generater[4] = {0};
    SM3_CTX *sm3_ctx = NULL;
    uint8_t *tmp = NULL;
    
    if(klen_bit % V == 0)
        hlen1 = klen_bit / V;
    else
        hlen1 = klen_bit / V + 1;
    
    sm3_ctx = calloc(1, sizeof(SM3_CTX));
    tmp = calloc(SM3_DIGEST_LENGTH, sizeof(uint8_t));

    for(i=1;i<=hlen1;i++)
    {
        generater[0] = (i >> 24) & 0xff;
        generater[1] = (i >> 16) & 0xff;
        generater[2] = (i >> 8) & 0xff;
        generater[3] = (i & 0xff);
        
        sm3_init(sm3_ctx);
        sm3_update(sm3_ctx, x, ECC_LEN);
        sm3_update(sm3_ctx, y, ECC_LEN);
        sm3_update(sm3_ctx, Zinitiator, ECC_LEN);
        sm3_update(sm3_ctx, Zresponder, ECC_LEN);
        sm3_update(sm3_ctx, generater, 4);
        sm3_final(sm3_ctx, tmp);
        
        if(klen_bit >= V)
            memcpy(key + (i - 1) * SM3_DIGEST_LENGTH, tmp, SM3_DIGEST_LENGTH);
        else
            memcpy(key + (i - 1) * SM3_DIGEST_LENGTH,  tmp, klen_bit / 8);
        
        klen_bit -= V;
    }
    
    free(sm3_ctx);
    free(tmp);
#undef V
}

void sm2_ka_check(uint8_t value, uint8_t *Zinitiator, uint8_t *Zresponder, uint8_t *Rinitiator, uint8_t *Rresponder, ECC_POINT *UV, uint8_t *S)
{
    SM3_CTX *sm3_ctx = NULL;
    sm3_ctx = calloc(1, sizeof(SM3_CTX));
    
    sm3_init(sm3_ctx);
    sm3_update(sm3_ctx, UV -> x, ECC_LEN);
    sm3_update(sm3_ctx, Zinitiator, ECC_LEN);
    sm3_update(sm3_ctx, Zresponder, ECC_LEN);
    sm3_update(sm3_ctx, Rinitiator, ECC_LEN * 2);
    sm3_update(sm3_ctx, Rresponder, ECC_LEN * 2);
    sm3_final(sm3_ctx, S);
    
    sm3_init(sm3_ctx);
    sm3_update(sm3_ctx, &value, 1);
    sm3_update(sm3_ctx, UV -> y, ECC_LEN);
    sm3_update(sm3_ctx, S, ECC_LEN);
    sm3_final(sm3_ctx, S);
    
    free(sm3_ctx);
}



void  sm2_std_ka_initiator_step1(uint8_t *tmpPriInitiator, uint8_t *tmpPubInitiator)
{
    bigrand_get_rand_range(tmpPriInitiator, (uint8_t *)curve_sm2_std_n, ECC_LEN);
    sm2_std_genPubkey(tmpPriInitiator, tmpPubInitiator);

}

uint16_t sm2_std_ka_initiator_step2(uint8_t *IDinitiator,         \
                                    uint16_t IDinitiator_len,     \
                                    uint8_t *IDresponder,         \
                                    uint16_t IDresponder_len,     \
                                    uint8_t *priInitiator,        \
                                    uint8_t *pubInitiator,        \
                                    uint8_t *pubResponder,        \
                                    uint8_t *tmpPriInitiator,     \
                                    uint8_t *tmpPubInitiator,     \
                                    uint8_t *tmpPubResponder,     \
                                    uint8_t *Sin,                 \
                                    uint8_t *Sout,                \
                                    uint16_t keylen,              \
                                    uint8_t *key)
{
    uint8_t *tmp1 = NULL, *tmp2 = NULL;
    uint8_t *Zinitiator = NULL, *Zresponder = NULL;
    ECC_POINT *point1 = NULL, *point2 = NULL;
    ECC_CURVE_PARAM *curveParam = NULL;
    curveParam = calloc(1, sizeof(ECC_CURVE_PARAM));
    
    curveParam -> p = (uint8_t *)curve_sm2_std_p;
    curveParam -> a = (uint8_t *)curve_sm2_std_a;
    curveParam -> b = (uint8_t *)curve_sm2_std_b;
    curveParam -> x = (uint8_t *)curve_sm2_std_x;
    curveParam -> y = (uint8_t *)curve_sm2_std_y;
    curveParam -> n = (uint8_t *)curve_sm2_std_n;
    point1 = calloc(1, sizeof(ECC_POINT));
    
    memcpy(point1 -> x, tmpPubResponder, ECC_LEN);
    memcpy(point1 -> y, tmpPubResponder + ECC_LEN, ECC_LEN);
    
    if(!is_pubkey_legal(curveParam, point1))
    {
        free(curveParam);
        free(point1);
        return FAILURE;
    }
    
    tmp1 = calloc(ECC_LEN, sizeof(uint8_t));
    tmp2 = calloc(ECC_LEN, sizeof(uint8_t));
    
    memcpy(tmp1 + ECC_LEN / 2, tmpPubInitiator + ECC_LEN / 2, ECC_LEN / 2);
    *(tmp1 + ECC_LEN / 2) |= 0x80;
    
    bignum_mod_mul(tmp1, tmpPriInitiator, curveParam -> n, tmp2);
    bignum_mod_add(tmp2, priInitiator, curveParam -> n, tmp1);
    
    memset(tmp2, 0, ECC_LEN / 2);
    memcpy(tmp2 + ECC_LEN / 2, tmpPubResponder + ECC_LEN / 2, ECC_LEN / 2);
    *(tmp2 + ECC_LEN / 2) |= 0x80;
    
    point_mul(curveParam, point1, tmp2, point1);
    
    point2 = calloc(1, sizeof(ECC_POINT));
    memcpy(point2 -> x, pubResponder, ECC_LEN);
    memcpy(point2 -> y, pubResponder + ECC_LEN, ECC_LEN);
    
    point_add(curveParam, point1, point2, point1);
    
    if(point_mul(curveParam, point1, tmp1, point1))
    {
        free(curveParam);
        free(point1);
        free(point2);
        free(tmp1);
        free(tmp2);
        return FAILURE;
    }
    
    Zinitiator = calloc(SM3_DIGEST_LENGTH, sizeof(uint8_t));
    Zresponder = calloc(SM3_DIGEST_LENGTH, sizeof(uint8_t));
    sm2_ka_get_Z(IDinitiator, IDinitiator_len, pubInitiator, Zinitiator);
    sm2_ka_get_Z(IDresponder, IDresponder_len, pubResponder, Zresponder);
    
    sm2_ka_KDF(point1 -> x, point1 -> y, Zinitiator, Zresponder, keylen * 8, key);
    
    sm2_ka_check(0x02, Zinitiator, Zresponder, tmpPubInitiator, tmpPubResponder, point1, tmp2);
    
    if(memcmp(tmp2, Sin, SM3_DIGEST_LENGTH))
    {
        free(curveParam);
        free(point1);
        free(point2);
        free(tmp1);
        free(tmp2);
        free(Zinitiator);
        free(Zresponder);
        return FAILURE;
    }
    
    sm2_ka_check(0x03, Zinitiator, Zresponder, tmpPubInitiator, tmpPubResponder, point1, Sout);
    
    free(curveParam);
    free(point1);
    free(point2);
    free(tmp1);
    free(tmp2);
    free(Zinitiator);
    free(Zresponder);
    return SUCCESS;
}

uint16_t sm2_std_ka_responder_step1(uint8_t *IDinitiator,         \
                                    uint16_t IDinitiator_len,     \
                                    uint8_t *IDresponder,         \
                                    uint16_t IDresponder_len,     \
                                    uint8_t *priResponder,        \
                                    uint8_t *pubResponder,        \
                                    uint8_t *pubInitiator,        \
                                    uint8_t *tmpPubResponder,     \
                                    uint8_t *tmpPubInitiator,     \
                                    uint8_t *Sinner,              \
                                    uint8_t *Souter,              \
                                    uint16_t keylen,              \
                                    uint8_t *key)
{
    uint8_t *tmpPriResponder = NULL, *tmp1 = NULL, *tmp2 = NULL;
    uint8_t *Zinitiator = NULL, *Zresponder = NULL;
    ECC_POINT *point1 = NULL, *point2 = NULL;
    ECC_CURVE_PARAM *curveParam = NULL;
    curveParam = calloc(1, sizeof(ECC_CURVE_PARAM));
    
    curveParam -> p = (uint8_t *)curve_sm2_std_p;
    curveParam -> a = (uint8_t *)curve_sm2_std_a;
    curveParam -> b = (uint8_t *)curve_sm2_std_b;
    curveParam -> x = (uint8_t *)curve_sm2_std_x;
    curveParam -> y = (uint8_t *)curve_sm2_std_y;
    curveParam -> n = (uint8_t *)curve_sm2_std_n;
    point1 = calloc(1, sizeof(ECC_POINT));
    
    memcpy(point1 -> x, tmpPubInitiator, ECC_LEN);
    memcpy(point1 -> y, tmpPubInitiator + ECC_LEN, ECC_LEN);
    
    if(!is_pubkey_legal(curveParam, point1))
    {
        free(curveParam);
        free(point1);
        return FAILURE;
    }
    
    tmpPriResponder = calloc(ECC_LEN, sizeof(uint8_t));

    bigrand_get_rand_range(tmpPriResponder, curveParam -> n, ECC_LEN);

    sm2_std_genPubkey(tmpPriResponder, tmpPubResponder);
    
    tmp1 = calloc(ECC_LEN, sizeof(uint8_t));
    tmp2 = calloc(ECC_LEN, sizeof(uint8_t));
    memcpy(tmp1 + ECC_LEN / 2, tmpPubResponder + ECC_LEN / 2, ECC_LEN / 2);
    *(tmp1 + ECC_LEN / 2) |= 0x80;
    
    bignum_mod_mul(tmp1, tmpPriResponder, curveParam -> n, tmp2);
    bignum_mod_add(tmp2, priResponder, curveParam -> n, tmp1);
    
    memset(tmp2, 0, ECC_LEN / 2);
    memcpy(tmp2 + ECC_LEN / 2, tmpPubInitiator + ECC_LEN / 2, ECC_LEN / 2);
    *(tmp2 + ECC_LEN / 2) |= 0x80;
    
    point_mul(curveParam, point1, tmp2, point1);
    
    point2 = calloc(1, sizeof(ECC_POINT));
    memcpy(point2 -> x, pubInitiator, ECC_LEN);
    memcpy(point2 -> y, pubInitiator + ECC_LEN, ECC_LEN);
    
    point_add(curveParam, point1, point2, point1);
    
    if(point_mul(curveParam, point1, tmp1, point1))
    {
        free(curveParam);
        free(tmp1);
        free(tmp2);
        free(point1);
        free(point2);
        return FAILURE;
    }
    
    Zinitiator = calloc(SM3_DIGEST_LENGTH, sizeof(uint8_t));
    Zresponder = calloc(SM3_DIGEST_LENGTH, sizeof(uint8_t));
    sm2_ka_get_Z(IDinitiator, IDinitiator_len, pubInitiator, Zinitiator);
    sm2_ka_get_Z(IDresponder, IDresponder_len, pubResponder, Zresponder);
    
    sm2_ka_KDF(point1 -> x, point1 -> y, Zinitiator, Zresponder, keylen * 8, key);
    
    sm2_ka_check(0x03, Zinitiator, Zresponder, tmpPubInitiator, tmpPubResponder, point1, Sinner);
    
    sm2_ka_check(0x02, Zinitiator, Zresponder, tmpPubInitiator, tmpPubResponder, point1, Souter);
    
    free(curveParam);
    free(tmp1);
    free(tmp2);
    free(point1);
    free(point2);
    free(Zinitiator);
    free(Zresponder);
    return SUCCESS;
}

uint16_t sm2_std_ka_responder_step2(uint8_t *Sinitiator, uint8_t *Sresponder)
{
    if(memcmp(Sinitiator, Sresponder, ECC_LEN))
        return FAILURE;
    return SUCCESS;
}
/*
 @function:(Point)outpointbuf = (Point)inputpoint1_buf+[k](Point)inputpoint2_buf
 @paramter[in]:inputpoint1_buf pointer to one point(stored by byte string) on the elliptic
 @paramter[in]:inputpoint2_buf pointer to another point(stored by byte string) on the elliptic
 @paramter[in]:k pointer to the multiplicator
 @paramter[out]:outpointbuf pointer to the result(stored by byte string)
 @return:0表示运算失败；1表示运算成功.
 */
uint16_t sm2_point_mul_add(uint8_t *inputpoint1_buf,uint8_t *inputpoint2_buf,uint8_t *k,uint8_t *outpointbuf)
{
    uint16_t ret;
    ECC_CURVE_PARAM *curveParam;
    curveParam = calloc(1,sizeof(ECC_CURVE_PARAM));
    curveParam->a =(uint8_t *)curve_sm2_std_a;
    curveParam->b = (uint8_t *)curve_sm2_std_b;
    curveParam->p=(uint8_t *)curve_sm2_std_p;
    curveParam->n =(uint8_t *)curve_sm2_std_n;
    curveParam->x =(uint8_t *)curve_sm2_std_x;
    curveParam->y =(uint8_t *)curve_sm2_std_y;
    ret=point_mul_add(curveParam,inputpoint1_buf,inputpoint2_buf,k,outpointbuf);
    free(curveParam);
    return ret;
}

/*
 @function:(Point)outpoint_buf = (Point)inputpoint_buf+[k]G(G is the base point of curve elliptic)
 @paramter[in]:P pointer to one point(stored by byte string) on the elliptic
 @paramter[in]:k pointer to the multiplicator
 @paramter[out]:outpoint_buf pointer to the result(stored by byte string)
 @return:0表示运算失败；1表示运算成功.
 */
uint16_t sm2_point_mul_baseG_add(uint8_t *inputpoint_buf,uint8_t *k,uint8_t *outpoint_buf)
{
    uint16_t ret;
    uint8_t *temp_point_buf=NULL;
    ECC_CURVE_PARAM *curveParam;
    curveParam = calloc(1,sizeof(ECC_CURVE_PARAM));
    temp_point_buf=calloc(ECC_LEN<<1,sizeof(uint8_t));
    curveParam->a =(uint8_t *)curve_sm2_std_a;
    curveParam->b = (uint8_t *)curve_sm2_std_b;
    curveParam->p=(uint8_t *)curve_sm2_std_p;
    curveParam->n =(uint8_t *)curve_sm2_std_n;
    curveParam->x =(uint8_t *)curve_sm2_std_x;
    curveParam->y =(uint8_t *)curve_sm2_std_y;
    memcpy(temp_point_buf,curve_sm2_std_x,ECC_LEN);
    memcpy(temp_point_buf + ECC_LEN,curve_sm2_std_y,ECC_LEN);
   // G->infinity = 0;
    ret=point_mul_add(curveParam,inputpoint_buf,temp_point_buf,k,outpoint_buf);
    free(curveParam);
    free(temp_point_buf);
    return ret;
}

/*
 @function:椭圆曲线（sm2）上点的压缩
 @paramter[in]:point_buf,待压缩的点（stored by byte string）
 @paramter[in]:point_buf_len表示point_buf的字节长度
 @paramter[out]:x,点压缩后的横坐标（长度为ECC_LEN+1 字节）
 @return：0 表示压缩失败；1 表示压缩成功
 */
uint16_t sm2_point_compress(uint8_t *point_buf,uint16_t point_buf_len,uint8_t *x)
{
    return point_compress(point_buf,point_buf_len,x);
}
/*
 @function:椭圆曲线(sm2)点的解压缩
 @paramter[in]:x pointer to the x-coordiate of the point on curve elliptic
 @paramter[in]:x_len denotes the byte length of x(x_len=ECC_LEN=1)
 @paramter[out]:point_buf pointer to the xy-coordiate(with 0x04) of the point on curve elliptic
 @return：0 表示解压缩失败；1 表示解压缩成功
 */
uint16_t sm2_point_decompress(uint8_t *x,uint16_t x_len,uint8_t *point_buf)
{
    uint16_t ret;
    ECC_CURVE_PARAM *curveParam=NULL;
    curveParam = calloc(1,sizeof(ECC_CURVE_PARAM));
    curveParam->a = (uint8_t *)curve_sm2_std_a;
    curveParam->b = (uint8_t *)curve_sm2_std_b;
    curveParam->n =(uint8_t *)curve_sm2_std_n;
    curveParam->p =(uint8_t *)curve_sm2_std_p;
    curveParam->x = (uint8_t *)curve_sm2_std_x;
    curveParam->y =(uint8_t *)curve_sm2_std_y;
    ret=point_decompress(curveParam, x,x_len,point_buf);
    free(curveParam);
    return ret;
}




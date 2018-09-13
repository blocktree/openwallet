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

#include "ECDSA.h"

uint16_t ECDSA_genPubkey(ECC_CURVE_PARAM *curveParam, uint8_t *prikey, ECC_POINT *pubkey)
{
    ECC_POINT *point_g = NULL;
    
    if(!is_prikey_legal(curveParam, prikey))
        return ECC_PRIKEY_ILLEGAL;
    
    point_g = calloc(1, sizeof(ECC_POINT));
    
    memcpy(point_g -> x, curveParam -> x, ECC_LEN);
    memcpy(point_g -> y, curveParam -> y, ECC_LEN);
    
    if(point_mul(curveParam, point_g, prikey, pubkey))
    {
        free(point_g);
        return ECC_PRIKEY_ILLEGAL;
    }
    
    free(point_g);
    return SUCCESS;
}



uint16_t ECDSA_sign(ECC_CURVE_PARAM *curveParam, uint8_t *prikey, uint8_t *message, uint16_t message_len, uint8_t *sig)
{
    uint8_t *k = NULL, *tmp = NULL;
    ECC_POINT *point = NULL;
    
    if(!is_prikey_legal(curveParam, prikey))
        return ECC_PRIKEY_ILLEGAL;
    
    k = calloc(ECC_LEN, sizeof(uint8_t));
    tmp = calloc(ECC_LEN, sizeof(uint8_t));
    point = calloc(1, sizeof(ECC_POINT));
    
    while(1)
    {
        memcpy(point -> x, curveParam -> x, ECC_LEN);
        memcpy(point -> y, curveParam -> y, ECC_LEN);
        
        bigrand_get_rand_range(k, curveParam -> n, ECC_LEN);
        
        point_mul(curveParam, point, k, point);
        
        bignum_mod(point -> x, curveParam -> n, sig);
        
        if(is_all_zero(sig, ECC_LEN))
            continue;
        
        sha256_hash(message, message_len, tmp);
        
        bignum_mod_mul(prikey, sig, curveParam -> n, sig + ECC_LEN);
        bignum_mod_add(tmp, sig + ECC_LEN, curveParam -> n, tmp);
        bignum_mod_inv(k, curveParam -> n, k);
        bignum_mod_mul(k, tmp, curveParam -> n, sig + ECC_LEN);
        
        if(is_all_zero(sig + ECC_LEN, ECC_LEN))
            continue;
        else
            break;
    }
    
    free(k);
    free(tmp);
    free(point);
    
    return SUCCESS;
}

uint16_t ECDSA_verify(ECC_CURVE_PARAM *curveParam, ECC_POINT *pubkey, uint8_t *message, uint16_t message_len, uint8_t *sig)
{
    uint8_t *tmp1 = NULL, *tmp2 = NULL;
    ECC_POINT *point1 = NULL, *point2 = NULL;
    
    if(!is_pubkey_legal(curveParam, pubkey))
        return ECC_PUBKEY_ILLEGAL;
    
    if(is_all_zero(sig, ECC_LEN) || memcmp(sig, curveParam -> n, ECC_LEN) >= 0 || is_all_zero(sig + ECC_LEN, ECC_LEN) || memcmp(sig + ECC_LEN, curveParam -> n, ECC_LEN) >= 0)
        return FAILURE;
    
    tmp1 = calloc(ECC_LEN, sizeof(uint8_t));
    tmp2 = calloc(ECC_LEN, sizeof(uint8_t));
    point1 = calloc(1, sizeof(ECC_POINT));
    point2 = calloc(1, sizeof(ECC_POINT));
    
    sha256_hash(message, message_len, tmp1);
    bignum_mod_inv(sig + ECC_LEN, curveParam -> n, tmp2);
    
    bignum_mod_mul(tmp1, tmp2, curveParam -> n, tmp1);
    bignum_mod_mul(sig, tmp2, curveParam -> n, tmp2);
    
    memcpy(point2 -> x, curveParam -> x, ECC_LEN);
    memcpy(point2 -> y, curveParam -> y, ECC_LEN);
    
    point_mul(curveParam, point2, tmp1, point1);
    point_mul(curveParam, pubkey, tmp2, point2);
    
    if(point_add(curveParam, point1, point2, point1))
    {
        free(tmp1);
        free(tmp2);
        free(point1);
        free(point2);
        return FAILURE;
    }
    
    bignum_mod(point1 -> x, curveParam -> n, tmp1);
    
    if(memcmp(tmp1, sig, ECC_LEN))
    {
        free(tmp1);
        free(tmp2);
        free(point1);
        free(point2);
        return FAILURE;
    }
    
    free(tmp1);
    free(tmp2);
    free(point1);
    free(point2);
    return SUCCESS;
}


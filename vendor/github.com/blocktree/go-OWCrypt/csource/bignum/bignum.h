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

#ifndef bignum_h
#define bignum_h

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include "type.h"

#define BN_LEN 32

uint8_t  bignum_add(uint8_t *a, uint8_t *b, uint16_t len, uint8_t *r);
void     bignum_add_by_1(uint8_t *a);  //不考虑全0xFF的情况
uint8_t  bignum_sub(uint8_t *a, uint8_t *b, uint16_t len, uint8_t *r);
uint16_t get_bit_len(uint8_t *a, uint16_t len);
uint8_t  get_bit_value(uint8_t *a, uint16_t alen, uint16_t index);
void     bignum_mul(uint8_t *a, uint8_t *b, uint8_t *r);//采用二进制展开方式计算
void     bignum_div(uint8_t *a, uint8_t *b, uint8_t *r);//不判断被除数是否为零
void     bignum_mod(uint8_t *a, uint8_t *b, uint8_t *r);
void     bignum_mod_with_carry(uint8_t *a, uint8_t *b, uint8_t *r);//不考虑a为二进制全1的情况
void     bignum_mod_add(uint8_t *a, uint8_t *b, uint8_t *n, uint8_t *r);
void     bignum_mod_sub(uint8_t *a, uint8_t *b, uint8_t *n, uint8_t *r);
void     bignum_shr_1bit(uint8_t *a, uint16_t len);
uint8_t  is_all_zero(uint8_t *dest, uint16_t len);
void     bignum_mod_mul(uint8_t *a, uint8_t *b, uint8_t *n, uint8_t *r);
void     bignum_mod_exp(uint8_t *a, uint8_t *b, uint8_t *n, uint8_t *r);
void     bignum_mod_inv(uint8_t *a, uint8_t *n, uint8_t *r);//扩展欧几里得与费马小定理//r = a ^ (n - 2) mod n


/*
 @function:big number compare
 @paramter[in]:a pointer to one big number
 @paramter[in]:alen,the byte length of a
 @paramter[in]:b pointer to another big number
 @parametr[in]:blen,the byte length of b
 @return:0:a=b;1:a>b;-1:a<b
 */
int8_t bignum_cmp(uint8_t *a, uint16_t alen,uint8_t *b,uint16_t blen);

#endif /* bignum_h */

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

#include "bignum.h"
#include "montgamery.h"

extern int32_t mont_get_bits(const mont_bignum *a);
void r_align(mont_bignum *mp_r, uint8_t *r)
{
    uint16_t rlen = 0, offset = 0;
    int16_t i = 0;
    
    rlen = mont_get_bits(mp_r);
    offset = (256 - rlen) / 8;
    
    if(offset)
    {
        for(i = 31; i >= offset; i --)
            *(r + i) = *(r + i - offset);
        for(i = 0; i < offset; i ++)
            *(r + i) = 0;
    }
}

uint8_t bignum_add(uint8_t *a, uint8_t *b, uint16_t len, uint8_t *r)
{
    int8_t i = 0;
    uint8_t carry = 0; //进位
    uint16_t tmp = 0;
    
    for(i = len - 1; i >= 0; i --)
    {
        tmp = (uint16_t)(*(a + i)) + (uint16_t)(*(b + i)) + carry;
        *(r + i) = tmp & 0x00FF;
        if(tmp & 0x0100)
            carry = 1;
        else
            carry = 0;
    }
    
    return carry;
}

//不考虑全0xFF的情况
void bignum_add_by_1(uint8_t *a)
{
    int8_t i = 0;
    
    for(i = BN_LEN - 1; i >= 0; i --)
    {
        if(*(a + i) == 0xFF)
        {
            *(a + i) = 0x00;
        }
        else
        {
            *(a + i) += 0x01;
            break;
        }
    }
}

uint8_t bignum_sub(uint8_t *a, uint8_t *b, uint16_t len, uint8_t *r)
{
    int8_t i = 0;
    uint8_t borrow = 0;
    
    for(i = len - 1; i >= 0; i --)
    {
        if(borrow)
        {
            if((*(a + i) - 1) == *(b + i))
            {
                *(r + i) = 0;
                borrow = 0;
            }
            else if((*(a + i) - 1) > *(b + i))
            {
                *(r + i) = *(a + i) - 1 - *(b + i);
                borrow = 0;
            }
            else
            {
                *(r + i) = (0xFF - (*(b + i) - (*(a + i) - 1))) + 0x01;
                borrow = 1;
            }
        }
        else
        {
            if(*(a + i) == *(b + i))
                *(r + i) = 0;
            else if(*(a + i) > *(b + i))
                *(r + i) = *(a + i) - *(b + i);
            else
            {
                *(r + i) = (0xFF - (*(b + i) - *(a + i))) + 0x01;
                borrow = 1;
            }
        }

    }
    return borrow;
}

uint16_t get_bit_len(uint8_t *a, uint16_t len)
{
    uint16_t i = 0;
    int16_t j = 0;
    uint16_t ret_len = len;
    
    for(i = 0; i < len; i ++)
    {
        if(*(a + i) == 0)
            ret_len --;
        else
            break;
    }
    
    ret_len *= 8;
    
    if(ret_len)
    {
        for(j = 7; j >= 0; j --)
        {
            if(*(a + i) & 1 << j)
                break;
            else
                ret_len --;
        }
    }
    return ret_len;
}

uint8_t get_bit_value(uint8_t *a, uint16_t alen, uint16_t index)
{
    uint16_t byte_index = BN_LEN - 1 - (index / 8);
    uint8_t bit_index = index % 8;
    
    return *(a + byte_index) & ( 1 << bit_index);
}

//采用二进制展开方式计算
void bignum_mul(uint8_t *a, uint8_t *b, uint8_t *r)
{
    uint8_t *tmp = NULL, *count = NULL;
    uint16_t bit_len = 0;
    uint16_t i = 0;
    
    memset(r, 0, BN_LEN * 2);
    
    tmp = calloc(BN_LEN * 2, sizeof(uint8_t));
    
    if(memcmp(a, b, BN_LEN) >= 0)
    {
        memcpy(tmp + BN_LEN, a, BN_LEN);
        count = b;
    }
    else
    {
        memcpy(tmp + BN_LEN, b, BN_LEN);
        count = a;
    }
   
    bit_len = get_bit_len(count, BN_LEN);
    
    if(get_bit_value(count, BN_LEN, 0))
    {
        memcpy(r + BN_LEN, tmp + BN_LEN, BN_LEN);
    }
    
    for(i = 1; i < bit_len; i ++)
    {
        bignum_add(tmp, tmp, BN_LEN * 2, tmp);
        if(get_bit_value(count, BN_LEN * 2, i))
        {
            bignum_add(tmp, r, BN_LEN * 2, r);
        }
    }
    
    free(tmp);
}

//不判断被除数是否为零
void bignum_div(uint8_t *a, uint8_t *b, uint8_t *r)
{
    uint8_t *tmp = NULL;
    int ret = 0;
    memset( r, 0, BN_LEN);

    ret = memcmp(a, b, BN_LEN) ;
    if(ret == 0)
    {
        *(r + BN_LEN - 1) = 0x01;
    }
    else if(ret < 0)
    {
        ;
    }
    else
    {
        tmp = calloc(BN_LEN, sizeof(uint8_t));
        memcpy(tmp, a, BN_LEN);
        while(1)
        {
            bignum_sub(tmp, b, BN_LEN, tmp);
            bignum_add_by_1(r);
            if(memcmp(tmp, b, BN_LEN) < 0)
            {
                free(tmp);
                break;
            }
        }
    }
}

void bignum_mod(uint8_t *a, uint8_t *b, uint8_t *r)
{
    int ret = 0;
    
    ret = memcmp(a, b, BN_LEN) ;
    if(ret == 0)
    {
        memset( r, 0, BN_LEN);
    }
    else if(ret < 0)
    {
        memcpy(r, a, BN_LEN);
    }
    else
    {
        memcpy(r, a, BN_LEN);
        while(1)
        {
            bignum_sub(r, b, BN_LEN, r);
            if(memcmp(r, b, BN_LEN) < 0)
                break;
        }
    }
}

//不考虑a为二进制全1的情况
void bignum_mod_with_carry(uint8_t *a, uint8_t *b, uint8_t *r)
{
    uint8_t *tmpa = NULL, *tmpF = NULL;
    
    tmpa = calloc(BN_LEN, sizeof(uint8_t));
    tmpF = calloc(BN_LEN, sizeof(uint8_t));
    
    memcpy(tmpa, a, BN_LEN);
    memset(tmpF, 0xFF, BN_LEN);
    
    bignum_add_by_1(tmpa);
    
    bignum_mod(tmpa, b, tmpa);
    bignum_mod(tmpF, b, tmpF);
    
    bignum_add(tmpa, tmpF, BN_LEN, r);
    
    bignum_mod(r, b, r);
    
    free(tmpa);
    free(tmpF);
}

void bignum_mod_add(uint8_t *a, uint8_t *b, uint8_t *n, uint8_t *r)
{
    uint8_t *tmp = NULL;
    
    tmp = calloc(BN_LEN, sizeof(uint8_t));
    
    if(bignum_add(a, b, BN_LEN, tmp))
    {
        bignum_mod_with_carry(tmp, n, r);
    }
    else
    {
        bignum_mod(tmp, n, r);
    }
    
    free(tmp);
}

void bignum_mod_sub(uint8_t *a, uint8_t *b, uint8_t *n, uint8_t *r)
{
    int ret = 0;
    uint8_t *tmp = NULL;
    
    ret = memcmp(a, b, BN_LEN);
    
    if(ret == 0)
    {
        memset(r, 0x00, BN_LEN);
    }
    else if(ret > 0)
    {
        tmp = calloc(BN_LEN, sizeof(uint8_t));
        bignum_sub(a, b, BN_LEN, tmp);
        bignum_mod(tmp, n, r);
        free(tmp);
    }
    else
    {
        tmp = calloc(BN_LEN, sizeof(uint8_t));
        bignum_sub(b, a, BN_LEN, tmp);
        bignum_sub(n, tmp, BN_LEN, r);
        bignum_mod(r, n, r);
        free(tmp);
    }
}

void bignum_shr_1bit(uint8_t *a, uint16_t len)
{
    uint16_t i = 0;
    uint8_t tmp1 = 0, tmp2 = 0;
    
    for(i = 0; i < len; i ++)
    {
        tmp1 = (*(a + i) & 0x01) << 7;
        *(a + i) >>= 1;
        *(a + i) |= tmp2;
        tmp2 = tmp1;
    }
}

uint8_t is_all_zero(uint8_t *dest, uint16_t len)
{
    while(len --)
    {
        if(*(dest + len))
            return 0;
    }
    return 1;
}

void bignum_mod_mul(uint8_t *a, uint8_t *b, uint8_t *n, uint8_t *r)
{
    mont_bignum ma, mb, mn, mr;
    mont_bignum_init(&ma);
    mont_bignum_init(&mb);
    mont_bignum_init(&mn);
    mont_bignum_init(&mr);
    
    mont_bignum_bin2bn(&ma, a, BN_LEN);
    mont_bignum_bin2bn(&mb, b, BN_LEN);
    mont_bignum_bin2bn(&mn, n, BN_LEN);
    mont_bignum_bin2bn(&mr, r, BN_LEN);
    
    mont_mod_mul(&ma, &mb, &mn, &mr);
    
    mont_bignum_bn2bin(&mr, r);
    r_align(&mr, r);
    
    mont_bignum_free(&ma);
    mont_bignum_free(&mb);
    mont_bignum_free(&mn);
    mont_bignum_free(&mr);
}

void bignum_mod_exp(uint8_t *a, uint8_t *b, uint8_t *n, uint8_t *r)
{
    mont_bignum ma, mb, mn, mr;
    mont_bignum_init(&ma);
    mont_bignum_init(&mb);
    mont_bignum_init(&mn);
    mont_bignum_init(&mr);
    
    mont_bignum_bin2bn(&ma, a, BN_LEN);
    mont_bignum_bin2bn(&mb, b, BN_LEN);
    mont_bignum_bin2bn(&mn, n, BN_LEN);
    mont_bignum_bin2bn(&mr, r, BN_LEN);
    
    mont_mod_exp(&ma, &mb, &mn, &mr);
    
    mont_bignum_bn2bin(&mr, r);
    r_align(&mr, r);
    
    mont_bignum_free(&ma);
    mont_bignum_free(&mb);
    mont_bignum_free(&mn);
    mont_bignum_free(&mr);
}

void bignum_mod_inv(uint8_t *a, uint8_t *n, uint8_t *r)
{
    mont_bignum ma, mn, mr;
    mont_bignum_init(&ma);
    mont_bignum_init(&mn);
    mont_bignum_init(&mr);
    
    mont_bignum_bin2bn(&ma, a, BN_LEN);
    mont_bignum_bin2bn(&mn, n, BN_LEN);
    mont_bignum_bin2bn(&mr, r, BN_LEN);
    
    mont_mod_inv(&ma, &mn, &mr);
    
    mont_bignum_bn2bin(&mr, r);
    r_align(&mr, r);
    
    mont_bignum_free(&ma);
    mont_bignum_free(&mn);
    mont_bignum_free(&mr);
}


/*
 @function:big number compare
 @paramter[in]:a pointer to one big number
 @paramter[in]:alen,the byte length of a
 @paramter[in]:b pointer to another big number
 @parametr[in]:blen,the byte length of b
 @return:0:a=b;1:a>b;-1:a<b
 */
int8_t bignum_cmp(uint8_t *a, uint16_t alen,uint8_t *b,uint16_t blen)
{
    uint16_t i,a_templen,b_templen;
    
    //make sure the init i value is zero
    i=0;
    while(a[i]==0)
    {
        i++;
    }
    a_templen = alen-i;
    //make sure the init i value is zero
    i=0;
    while(b[i]==0)
    {
        i++;
    }
    b_templen=blen - i;
    if(a_templen > b_templen)
    {
        return 1;
    }
    else if(a_templen < b_templen)
    {
        return -1;
    }
    else
    {
        for(i=0;i < a_templen;i++)
        {
            if(a[i] > b[i])
            {
                return 1;
            }
            else if(a[i] < b[i])
            {
                return -1;
            }
            else
            {
                ;
            }
        }
        return 0;
    }
    
}




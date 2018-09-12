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

#include "montgamery.h"

#define MONT_UNIT_BIT    30
#define MONT_PRE_SIZE    32
#define MONT_NO    0
#define MONT_YES    1
#define CHAR_BIT         8
#define MONT_MASK        ((((mont_unit)1)<<((mont_unit)MONT_UNIT_BIT))-((mont_unit)1))
#define MONT_LIMIT       (1u << (((sizeof(mp_double_unit) * CHAR_BIT) - (2 * MONT_UNIT_BIT)) + 1))
#define mont_iseven(a)   ((((a)->unit_count == 0) || (((a)->data[0] & 1u) == 0u)) ? MONT_YES : MONT_NO)
#define mont_isodd(a)    ((((a)->unit_count > 0) && (((a)->data[0] & 1u) == 1u)) ? MONT_YES : MONT_NO)

#define MONT_CMP_LITTLE -1
#define MONT_CMP_EQUAL   0
#define MONT_CMP_GREAT   1

int32_t mont_bignum_init(mont_bignum *a)
{
    int32_t i = 0;
    a -> data = malloc(sizeof(mont_unit) * (size_t)MONT_PRE_SIZE);
    if(a -> data == NULL)
        return MONT_ERROR;
    for(i = 0; i < MONT_PRE_SIZE; i ++)
        a -> data[i] = 0;
    a -> unit_count = 0;
    a -> alloc_size = MONT_PRE_SIZE;
    a -> is_negative = MONT_NO;
    
    return MONT_OKAY;
}

int32_t mont_bignum_init_size(mont_bignum *a, int32_t size)
{
    int32_t x = 0;
    size += (MONT_PRE_SIZE * 2) - (size % MONT_PRE_SIZE);
    a -> data = malloc(sizeof(mont_unit) * (size_t)size);
    if(a -> data == NULL)
        return MONT_ERROR;
    a -> unit_count  = 0;
    a -> alloc_size = size;
    a -> is_negative  = MONT_NO;
    for (x = 0; x < size; x++)
        a -> data[x] = 0;
    return MONT_OKAY;
}

int32_t mont_copy(const mont_bignum *a, mont_bignum *b)
{
    int32_t n = 0;
    if(a == b)
        return MONT_OKAY;
    mont_unit *tmpa, *tmpb;
    tmpa = a -> data;
    tmpb = b -> data;
    for (n = 0; n < a -> unit_count; n++)
        *tmpb++ = *tmpa++;
    for (; n < b -> unit_count; n++)
        *tmpb++ = 0;
    b -> unit_count = a -> unit_count;
    b -> is_negative = a -> is_negative;
    return MONT_OKAY;
}

int32_t mont_bignum_init_copy(mont_bignum *a, const mont_bignum *b)
{
    int32_t res = 0;
    if((res = mont_bignum_init_size(a, b -> unit_count)) != MONT_OKAY)
        return res;
    if((res = mont_copy(b, a)) != MONT_OKAY)
        mont_bignum_free(a);
    return res;
}

void mont_bignum_free(mont_bignum *a)
{
    int32_t i = 0;
    if(a -> data != NULL)
    {
        for (i = 0; i < a -> unit_count; i ++)
            a -> data[i] = 0;
        free(a -> data);
        a -> data = NULL;
        a -> alloc_size = a -> unit_count = 0;
        a -> is_negative  = MONT_NO;
    }
}

void mp_zero(mont_bignum *a)
{
    int32_t n;
    mont_unit *tmp;
    a -> is_negative = MONT_NO;
    a -> unit_count = 0;
    tmp = a -> data;
    for (n = 0; n < a -> alloc_size; n++)
        *tmp++ = 0;
}

int32_t mont_bignum_extend(mont_bignum *a, int size)
{
    int32_t i = 0;
    mont_unit *tmp;
    if(a -> alloc_size < size)
    {
        size += (MONT_PRE_SIZE * 2) - (size % MONT_PRE_SIZE);
        tmp = realloc(a -> data, sizeof(mont_unit) * (size_t)size);
        if(tmp == NULL)
            return MONT_ERROR;
        a -> data = tmp;
        i = a -> alloc_size;
        a -> alloc_size = size;
        for (; i < a -> alloc_size; i++)
            a -> data[i] = 0;
    }
    return MONT_OKAY;
}

void mont_bignum_trim_zero(mont_bignum *a)
{
    while((a -> unit_count > 0) && (a -> data[a -> unit_count - 1] == 0u))
        --(a -> unit_count);
    if(a -> unit_count == 0)
        a -> is_negative = MONT_NO;
}

int32_t mont_cmp_mag(const mont_bignum *a, const mont_bignum *b)
{
    int32_t n;
    mont_unit *tmpa, *tmpb;
    if(a -> unit_count > b -> unit_count)
        return MONT_CMP_GREAT;
    if(a -> unit_count < b -> unit_count)
        return MONT_CMP_LITTLE;
    tmpa = a -> data + (a -> unit_count - 1);
    tmpb = b -> data + (a -> unit_count - 1);
    for (n = 0; n < a -> unit_count; ++n, --tmpa, --tmpb)
    {
        if(*tmpa > *tmpb)
            return MONT_CMP_GREAT;
        if(*tmpa < *tmpb)
            return MONT_CMP_LITTLE;
    }
    return MONT_CMP_EQUAL;
}

int32_t mont_cmp(const mont_bignum *a, const mont_bignum *b)
{
    if(a -> is_negative != b -> is_negative)
        return (a -> is_negative == MONT_YES) ? MONT_CMP_LITTLE : MONT_CMP_GREAT;
    return (a -> is_negative == MONT_YES) ? mont_cmp_mag(b, a) : mont_cmp_mag(a, b);
}

int32_t mont_mod_phi(const mont_bignum *a, int32_t b, mont_bignum *c)
{
    int32_t x, res;
    if(b <= 0)
    {
        mp_zero(c);
        return MONT_OKAY;
    }
    if(b >= (a -> unit_count * MONT_UNIT_BIT))
    {
        res = mont_copy(a, c);
        return res;
    }
    if((res = mont_copy(a, c)) != MONT_OKAY)
        return res;
    for (x = (b / MONT_UNIT_BIT) + (((b % MONT_UNIT_BIT) == 0) ? 0 : 1); x < c -> unit_count; x++)
        c -> data[x] = 0;
    c -> data[b / MONT_UNIT_BIT] &= ((mont_unit)1 << (mont_unit)(b % MONT_UNIT_BIT)) - (mont_unit)1;
    mont_bignum_trim_zero(c);
    return MONT_OKAY;
}

int32_t shift_left_by_units(mont_bignum *a, int32_t b)
{
    int32_t x = 0, res = 0;
    if(b <= 0)
        return MONT_OKAY;
    if(a -> unit_count == 0)
        return MONT_OKAY;
    if(a -> alloc_size < (a -> unit_count + b))
        if((res = mont_bignum_extend(a, a -> unit_count + b)) != MONT_OKAY)
            return res;
    mont_unit *top, *bottom;
    a -> unit_count += b;
    top = a -> data + a -> unit_count - 1;
    bottom = (a -> data + a -> unit_count - 1) - b;
    for (x = a -> unit_count - 1; x >= b; x--)
        *top-- = *bottom--;
    top = a -> data;
    for (x = 0; x < b; x++)
        *top++ = 0;
    return MONT_OKAY;
}

void shift_right_by_units(mont_bignum *a, int32_t b)
{
    int32_t x = 0;
    if(b <= 0)
        return;
    if(a -> unit_count <= b)
    {
        mp_zero(a);
        return;
    }
    mont_unit *bottom, *top;
    bottom = a -> data;
    top = a -> data + b;
    for (x = 0; x < (a -> unit_count - b); x++)
        *bottom++ = *top++;
    for (; x < a -> unit_count; x++)
        *bottom++ = 0;
    a -> unit_count -= b;
}

int32_t shift_left_by_bits(mont_bignum *a, int32_t b)
{
    mont_unit d;
    int res = 0;
    if(a -> alloc_size < (a -> unit_count + (b / MONT_UNIT_BIT) + 1))
        if((res = mont_bignum_extend(a, a -> unit_count + (b / MONT_UNIT_BIT) + 1)) != MONT_OKAY)
            return res;
    d = (mont_unit)(b % MONT_UNIT_BIT);
    if(d != 0u)
    {
        mont_unit *tmpa, shift, mask, r, rr;
        int32_t x;
        mask = ((mont_unit)1 << d) - (mont_unit)1;
        shift = (mont_unit)MONT_UNIT_BIT - d;
        tmpa = a -> data;
        r = 0;
        for (x = 0; x < a -> unit_count; x++)
        {
            rr = (*tmpa >> shift) & mask;
            *tmpa = ((*tmpa << d) | r) & MONT_MASK;
            ++tmpa;
            r = rr;
        }
        if(r != 0u)
            a -> data[(a -> unit_count)++] = r;
    }
    mont_bignum_trim_zero(a);
    return MONT_OKAY;
}

int32_t shift_right_by_bits(const mont_bignum *a, int32_t b, mont_bignum *c, mont_bignum *d)
{
    mont_unit D, r, rr;
    int32_t x, res;
    if(b <= 0)
    {
        res = mont_copy(a, c);
        if(d != NULL)
            mp_zero(d);
        return res;
    }
    if((res = mont_copy(a, c)) != MONT_OKAY)
        return res;
    if(d != NULL)
        if((res = mont_mod_phi(a, b, d)) != MONT_OKAY)
            return res;
    if(b >= MONT_UNIT_BIT)
        shift_right_by_units(c, b / MONT_UNIT_BIT);
    D = (mont_unit)(b % MONT_UNIT_BIT);
    if(D != 0u)
    {
        mont_unit *tmpc, mask, shift;
        mask = ((mont_unit)1 << D) - 1uL;
        shift = (mont_unit)MONT_UNIT_BIT - D;
        tmpc = c -> data + (c -> unit_count - 1);
        r = 0;
        for (x = c -> unit_count - 1; x >= 0; x--)
        {
            rr = *tmpc & mask;
            *tmpc = (*tmpc >> D) | (r << shift);
            --tmpc;
            r = rr;
        }
    }
    mont_bignum_trim_zero(c);
    return MONT_OKAY;
}

void bn_reverse(uint8_t *s, int32_t len)
{
    int ix = 0, iy = len - 1;
    uint8_t t = 0;
    while (ix < iy)
    {
        t     = s[ix];
        s[ix] = s[iy];
        s[iy] = t;
        ++ix;
        --iy;
    }
}

int32_t mont_bignum_bin2bn(mont_bignum *a, const uint8_t *b, int32_t c)
{
    int32_t res = 0;
    mp_zero(a);
    while (c-- > 0)
    {
        if((res = shift_left_by_bits(a, 8)) != MONT_OKAY)
            return res;
        a -> data[0] |= *b++;
        a -> unit_count += 1;
    }
    mont_bignum_trim_zero(a);
    return MONT_OKAY;
}

int32_t mont_bignum_bn2bin(const mont_bignum *a, uint8_t *b)
{
    int32_t x, res;
    mont_bignum  t;
    if((res = mont_bignum_init_copy(&t, a)) != MONT_OKAY)
        return res;
    x = 0;
    while (t.unit_count != 0)
    {
        b[x++] = (uint8_t)(t.data[0] & 255u);
        if((res = shift_right_by_bits(&t, 8, &t, NULL)) != MONT_OKAY)
        {
            mont_bignum_free(&t);
            return res;
        }
    }
    bn_reverse(b, x);
    mont_bignum_free(&t);
    return MONT_OKAY;
}

int32_t mont_mul_arrays(const mont_bignum *a, const mont_bignum *b, mont_bignum *c, int32_t digs)
{
    int32_t res, olduse, pa, ix, iz;
    mont_unit W[MONT_LIMIT];
    mp_double_unit  _W;
    if(c -> alloc_size < digs)
        if((res = mont_bignum_extend(c, digs)) != MONT_OKAY)
            return res;
    pa = (digs < (a -> unit_count + b -> unit_count)) ? digs : (a -> unit_count + b -> unit_count);
    _W = 0;
    for (ix = 0; ix < pa; ix++)
    {
        int32_t tx, ty, iy;
        mont_unit *tmpx, *tmpy;
        ty = ((b -> unit_count - 1) < ix) ? (b -> unit_count - 1) : ix;
        tx = ix - ty;
        tmpx = a -> data + tx;
        tmpy = b -> data + ty;
        iy = ((a -> unit_count - tx) < (ty + 1)) ? (a -> unit_count - tx) : (ty + 1);
        for (iz = 0; iz < iy; ++iz)
            _W += (mp_double_unit)*tmpx++ * (mp_double_unit)*tmpy--;
        W[ix] = (mont_unit)_W & MONT_MASK;
        _W = _W >> (mp_double_unit)MONT_UNIT_BIT;
    }
    olduse  = c -> unit_count;
    c -> unit_count = pa;
    mont_unit *tmpc;
    tmpc = c -> data;
    for (ix = 0; ix < pa; ix++)
        *tmpc++ = W[ix];
    for (; ix < olduse; ix++)
        *tmpc++ = 0;
    mont_bignum_trim_zero(c);
    return MONT_OKAY;
}

int32_t mont_get_bits(const mont_bignum *a)
{
    int32_t r = 0;
    mont_unit q;
    if(a -> unit_count == 0)
        return 0;
    r = (a -> unit_count - 1) * MONT_UNIT_BIT;
    q = a -> data[a -> unit_count - 1];
    while (q > (mont_unit)0)
    {
        ++r;
        q >>= (mont_unit)1;
    }
    return r;
}

int32_t mont_add_unsigned(const mont_bignum *a, const mont_bignum *b, mont_bignum *c)
{
    const mont_bignum *x;
    int32_t olduse, res, min, max;
    if(a -> unit_count > b -> unit_count)
    {
        min = b -> unit_count;
        max = a -> unit_count;
        x = a;
    }
    else
    {
        min = a -> unit_count;
        max = b -> unit_count;
        x = b;
    }
    if(c -> alloc_size < (max + 1))
        if((res = mont_bignum_extend(c, max + 1)) != MONT_OKAY)
            return res;
    olduse = c -> unit_count;
    c -> unit_count = max + 1;
    mont_unit u, *tmpa, *tmpb, *tmpc;
    int32_t i;
    tmpa = a -> data;
    tmpb = b -> data;
    tmpc = c -> data;
    u = 0;
    for (i = 0; i < min; i++)
    {
        *tmpc = *tmpa++ + *tmpb++ + u;
        u = *tmpc >> (mont_unit)MONT_UNIT_BIT;
        *tmpc++ &= MONT_MASK;
    }
    if(min != max)
        for (; i < max; i++)
        {
            *tmpc = x->data[i] + u;
            u = *tmpc >> (mont_unit)MONT_UNIT_BIT;
            *tmpc++ &= MONT_MASK;
        }
    *tmpc++ = u;
    for (i = c -> unit_count; i < olduse; i++)
        *tmpc++ = 0;
    mont_bignum_trim_zero(c);
    return MONT_OKAY;
}

int32_t mont_sub_unsigned(const mont_bignum *a, const mont_bignum *b, mont_bignum *c)
{
    int32_t olduse, res, min, max;
    min = b -> unit_count;
    max = a -> unit_count;
    if(c -> alloc_size < max)
        if((res = mont_bignum_extend(c, max)) != MONT_OKAY)
            return res;
    olduse = c -> unit_count;
    c -> unit_count = max;
    
    
    mont_unit u, *tmpa, *tmpb, *tmpc;
    int32_t i;
    tmpa = a -> data;
    tmpb = b -> data;
    tmpc = c -> data;
    u = 0;
    for (i = 0; i < min; i++)
    {
        *tmpc = (*tmpa++ - *tmpb++) - u;
        u = *tmpc >> (((size_t)CHAR_BIT * sizeof(mont_unit)) - 1u);
        *tmpc++ &= MONT_MASK;
    }
    for (; i < max; i++)
    {
        *tmpc = *tmpa++ - u;
        u = *tmpc >> (((size_t)CHAR_BIT * sizeof(mont_unit)) - 1u);
        *tmpc++ &= MONT_MASK;
    }
    for (i = c -> unit_count; i < olduse; i++)
        *tmpc++ = 0;
    mont_bignum_trim_zero(c);
    return MONT_OKAY;
}

int32_t mont_add(const mont_bignum *a, const mont_bignum *b, mont_bignum *c)
{
    int32_t sa, sb, res;
    sa = a -> is_negative;
    sb = b -> is_negative;
    if(sa == sb)
    {
        c -> is_negative = sa;
        res = mont_add_unsigned(a, b, c);
    }
    else
    {
        if(mont_cmp_mag(a, b) == MONT_CMP_LITTLE)
        {
            c -> is_negative = sb;
            res = mont_sub_unsigned(b, a, c);
        }
        else
        {
            c -> is_negative = sa;
            res = mont_sub_unsigned(a, b, c);
        }
    }
    return res;
}

int32_t mont_sub(const mont_bignum *a, const mont_bignum *b, mont_bignum *c)
{
    int32_t sa, sb, res;
    sa = a -> is_negative;
    sb = b -> is_negative;
    if(sa != sb)
    {
        c -> is_negative = sa;
        res = mont_add_unsigned(a, b, c);
    }
    else
    {
        if(mont_cmp_mag(a, b) != MONT_CMP_LITTLE)
        {
            c -> is_negative = sa;
            res = mont_sub_unsigned(a, b, c);
        }
        else
        {
            c -> is_negative = (sa == MONT_NO) ? MONT_YES : MONT_NO;
            res = mont_sub_unsigned(b, a, c);
        }
    }
    return res;
}


int32_t mont_mul_unit(const mont_bignum *a, mont_unit b, mont_bignum *c)
{
    mont_unit u, *tmpa, *tmpc;
    mp_double_unit  r;
    int32_t ix, res, olduse;
    if(c -> alloc_size < (a -> unit_count + 1))
        if((res = mont_bignum_extend(c, a -> unit_count + 1)) != MONT_OKAY)
            return res;
    olduse = c -> unit_count;
    c -> is_negative = a -> is_negative;
    tmpa = a -> data;
    tmpc = c -> data;
    u = 0;
    for (ix = 0; ix < a -> unit_count; ix++)
    {
        r       = (mp_double_unit)u + ((mp_double_unit)*tmpa++ * (mp_double_unit)b);
        *tmpc++ = (mont_unit)(r & (mp_double_unit)MONT_MASK);
        u       = (mont_unit)(r >> (mp_double_unit)MONT_UNIT_BIT);
    }
    *tmpc++ = u;
    ++ix;
    while (ix++ < olduse)
        *tmpc++ = 0;
    c -> unit_count = a -> unit_count + 1;
    mont_bignum_trim_zero(c);
    return MONT_OKAY;
}

void mont_exchange(mont_bignum *a, mont_bignum *b)
{
    mont_bignum  t;
    t  = *a;
    *a = *b;
    *b = t;
}

int32_t mont_div(const mont_bignum *a, const mont_bignum *b, mont_bignum *c, mont_bignum *d)
{
    mont_bignum  q, x, y, t1, t2;
    int32_t res, n, t, i, norm, neg;
    if(b -> unit_count == 0)
        return MONT_ERROR;
    if(mont_cmp_mag(a, b) == MONT_CMP_LITTLE)
    {
        if(d != NULL)
            res = mont_copy(a, d);
        else
            res = MONT_OKAY;
        if(c != NULL)
            mp_zero(c);
        return res;
    }
    if((res = mont_bignum_init_size(&q, a -> unit_count + 2)) != MONT_OKAY)
        return res;
    q.unit_count = a -> unit_count + 2;
    if((res = mont_bignum_init(&t1)) != MONT_OKAY) goto LBL_Q;
    if((res = mont_bignum_init(&t2)) != MONT_OKAY) goto LBL_T1;
    if((res = mont_bignum_init_copy(&x, a)) != MONT_OKAY) goto LBL_T2;
    if((res = mont_bignum_init_copy(&y, b)) != MONT_OKAY) goto LBL_X;
    neg = (a -> is_negative == b -> is_negative) ? MONT_NO : MONT_YES;
    x.is_negative = y.is_negative = MONT_NO;
    norm = mont_get_bits(&y) % MONT_UNIT_BIT;
    if(norm < (MONT_UNIT_BIT - 1))
    {
        norm = (MONT_UNIT_BIT - 1) - norm;
        if((res = shift_left_by_bits(&x, norm)) != MONT_OKAY) goto LBL_Y;
        if((res = shift_left_by_bits(&y, norm)) != MONT_OKAY) goto LBL_Y;
    }
    else
    {
        norm = 0;
    }
    n = x.unit_count - 1;
    t = y.unit_count - 1;
    if((res = shift_left_by_units(&y, n - t)) != MONT_OKAY) goto LBL_Y;
    while (mont_cmp(&x, &y) != MONT_CMP_LITTLE)
    {
        ++(q.data[n - t]);
        if((res = mont_sub(&x, &y, &x)) != MONT_OKAY) goto LBL_Y;
    }
    shift_right_by_units(&y, n - t);
    for (i = n; i >= (t + 1); i--)
    {
        if(i > x.unit_count)
            continue;
        if(x.data[i] == y.data[t])
            q.data[(i - t) - 1] = ((mont_unit)1 << (mont_unit)MONT_UNIT_BIT) - (mont_unit)1;
        else
        {
            mp_double_unit tmp;
            tmp = (mp_double_unit)x.data[i] << (mp_double_unit)MONT_UNIT_BIT;
            tmp |= (mp_double_unit)x.data[i - 1];
            tmp /= (mp_double_unit)y.data[t];
            if(tmp > (mp_double_unit)MONT_MASK)
                tmp = MONT_MASK;
            q.data[(i - t) - 1] = (mont_unit)(tmp & (mp_double_unit)MONT_MASK);
        }
        q.data[(i - t) - 1] = (q.data[(i - t) - 1] + 1uL) & (mont_unit)MONT_MASK;
        do {
            q.data[(i - t) - 1] = (q.data[(i - t) - 1] - 1uL) & (mont_unit)MONT_MASK;
            mp_zero(&t1);
            t1.data[0] = ((t - 1) < 0) ? 0u : y.data[t - 1];
            t1.data[1] = y.data[t];
            t1.unit_count = 2;
            if((res = mont_mul_unit(&t1, q.data[(i - t) - 1], &t1)) != MONT_OKAY) goto LBL_Y;
            t2.data[0] = ((i - 2) < 0) ? 0u : x.data[i - 2];
            t2.data[1] = ((i - 1) < 0) ? 0u : x.data[i - 1];
            t2.data[2] = x.data[i];
            t2.unit_count = 3;
        } while (mont_cmp_mag(&t1, &t2) == MONT_CMP_GREAT);
        if((res = mont_mul_unit(&y, q.data[(i - t) - 1], &t1)) != MONT_OKAY) goto LBL_Y;
        if((res = shift_left_by_units(&t1, (i - t) - 1)) != MONT_OKAY) goto LBL_Y;
        if((res = mont_sub(&x, &t1, &x)) != MONT_OKAY) goto LBL_Y;
        if(x.is_negative == MONT_YES)
        {
            if((res = mont_copy(&y, &t1)) != MONT_OKAY) goto LBL_Y;
            if((res = shift_left_by_units(&t1, (i - t) - 1)) != MONT_OKAY) goto LBL_Y;
            if((res = mont_add(&x, &t1, &x)) != MONT_OKAY) goto LBL_Y;
            q.data[(i - t) - 1] = (q.data[(i - t) - 1] - 1uL) & MONT_MASK;
        }
    }
    x.is_negative = (x.unit_count == 0) ? MONT_NO : a -> is_negative;
    if(c != NULL)
    {
        mont_bignum_trim_zero(&q);
        mont_exchange(&q, c);
        c -> is_negative = neg;
    }
    if(d != NULL)
    {
        if((res = shift_right_by_bits(&x, norm, &x, NULL)) != MONT_OKAY) goto LBL_Y;
        mont_exchange(&x, d);
    }
    res = MONT_OKAY;
LBL_Y:
    mont_bignum_free(&y);
LBL_X:
    mont_bignum_free(&x);
LBL_T2:
    mont_bignum_free(&t2);
LBL_T1:
    mont_bignum_free(&t1);
LBL_Q:
    mont_bignum_free(&q);
    return res;
}

int32_t mont_mul(const mont_bignum *a, const mont_bignum *b, mont_bignum *c)
{
    int32_t res, neg;
    neg = (a -> is_negative == b -> is_negative) ? MONT_NO : MONT_YES;
    int digs = a -> unit_count + b -> unit_count + 1;
    res = mont_mul_arrays(a, b, c, digs);
    c -> is_negative = (c -> unit_count > 0) ? neg : MONT_NO;
    return res;
}

int32_t mont_mod(const mont_bignum *a, const mont_bignum *b, mont_bignum *c)
{
    mont_bignum  t;
    int32_t res;
    if((res = mont_bignum_init_size(&t, b -> unit_count)) != MONT_OKAY)
        return res;
    if((res = mont_div(a, b, NULL, &t)) != MONT_OKAY)
    {
        mont_bignum_free(&t);
        return res;
    }
    if(t.unit_count == 0 || (t.is_negative == b -> is_negative))
    {
        res = MONT_OKAY;
        mont_exchange(&t, c);
    }
    else
        res = mont_add(b, &t, c);
    mont_bignum_free(&t);
    return res;
}

int32_t mont_mod_mul(const mont_bignum *a, const mont_bignum *b, const mont_bignum *c, mont_bignum *d)
{
    int32_t res;
    mont_bignum  t;
    if((res = mont_bignum_init_size(&t, c -> unit_count)) != MONT_OKAY)
        return res;
    if((res = mont_mul(a, b, &t)) != MONT_OKAY)
    {
        mont_bignum_free(&t);
        return res;
    }
    res = mont_mod(&t, c, d);
    mont_bignum_free(&t);
    return res;
}

int32_t montgomery_init(const mont_bignum *n, mont_unit *rho)
{
    mont_unit x, b;
    b = n->data[0];
    if((b & 1u) == 0u)
        return MONT_ERROR;
    x = (((b + 2u) & 4u) << 1) + b;
    x *= 2u - (b * x);
    x *= 2u - (b * x);
    x *= 2u - (b * x);
    x *= 2u - (b * x);
    *rho = (mont_unit)(((mp_double_unit)1 << (mp_double_unit)MONT_UNIT_BIT) - x) & MONT_MASK;
    return MONT_OKAY;
}

int32_t montgomery_redc(mont_bignum *x, const mont_bignum *n, mont_unit rho)
{
    int32_t ix, res, olduse;
    mp_double_unit W[MONT_LIMIT];
    if(x->unit_count > (int)MONT_LIMIT)
        return MONT_ERROR;
    olduse = x->unit_count;
    if(x->alloc_size < (n->unit_count + 1))
        if((res = mont_bignum_extend(x, n->unit_count + 1)) != MONT_OKAY)
            return res;
    mp_double_unit *_W;
    mont_unit *tmp1;
    _W   = W;
    tmp1 = x->data;
    for (ix = 0; ix < x->unit_count; ix++)
        *_W++ = *tmp1++;
    for (; ix < ((n->unit_count * 2) + 1); ix++)
        *_W++ = 0;
    for (ix = 0; ix < n->unit_count; ix++)
    {
        mont_unit mu;
        mu = ((W[ix] & MONT_MASK) * rho) & MONT_MASK;
        int32_t iy;
        mont_unit *tmpn;
        mp_double_unit *_W;
        tmpn = n->data;
        _W = W + ix;
        for (iy = 0; iy < n->unit_count; iy++)
            *_W++ += (mp_double_unit)mu * (mp_double_unit)*tmpn++;
        W[ix + 1] += W[ix] >> (mp_double_unit)MONT_UNIT_BIT;
    }

    {
        mont_unit *tmpx;
        mp_double_unit *_W, *_W1;
        _W1 = W + ix;
        _W = W + ++ix;
        for (; ix <= ((n->unit_count * 2) + 1); ix++)
            *_W++ += *_W1++ >> (mp_double_unit)MONT_UNIT_BIT;
        tmpx = x->data;
        _W = W + n->unit_count;
        for (ix = 0; ix < (n->unit_count + 1); ix++)
            *tmpx++ = *_W++ & (mp_double_unit)MONT_MASK;
        for (; ix < olduse; ix++)
            *tmpx++ = 0;
    }
    x->unit_count = n->unit_count + 1;
    mont_bignum_trim_zero(x);
    if(mont_cmp_mag(x, n) != MONT_CMP_LITTLE)
        return mont_sub_unsigned(x, n, x);
    return MONT_OKAY;
}

int32_t mp_2expt(mont_bignum *a, int b)
{
    int32_t res;
    mp_zero(a);
    if((res = mont_bignum_extend(a, (b / MONT_UNIT_BIT) + 1)) != MONT_OKAY)
        return res;
    a -> unit_count = (b / MONT_UNIT_BIT) + 1;
    a -> data[b / MONT_UNIT_BIT] = (mont_unit)1 << (mont_unit)(b % MONT_UNIT_BIT);
    return MONT_OKAY;
}

void mp_set(mont_bignum *a, mont_unit b)
{
    mp_zero(a);
    a -> data[0] = b & MONT_MASK;
    a -> unit_count  = (a -> data[0] != 0u) ? 1 : 0;
}

int32_t mp_mul_2(const mont_bignum *a, mont_bignum *b)
{
    int x, res, oldused;
    if(b -> alloc_size < (a -> unit_count + 1))
        if((res = mont_bignum_extend(b, a -> unit_count + 1)) != MONT_OKAY)
            return res;
    oldused = b -> unit_count;
    b -> unit_count = a -> unit_count;
    {
        mont_unit r, rr, *tmpa, *tmpb;
        tmpa = a -> data;
        tmpb = b -> data;
        r = 0;
        for (x = 0; x < a -> unit_count; x++)
        {
            rr = *tmpa >> (mont_unit)(MONT_UNIT_BIT - 1);
            *tmpb++ = ((*tmpa++ << 1uL) | r) & MONT_MASK;
            r = rr;
        }
        if(r != 0u)
        {
            *tmpb = 1;
            ++(b -> unit_count);
        }
        tmpb = b -> data + b -> unit_count;
        for (x = b -> unit_count; x < oldused; x++)
            *tmpb++ = 0;
    }
    b -> is_negative = a -> is_negative;
    return MONT_OKAY;
}

int montgomery_normal(mont_bignum *a, const mont_bignum *b)
{
    int32_t x, bits, res;
    bits = mont_get_bits(b) % MONT_UNIT_BIT;
    if(b -> unit_count > 1) {
        if((res = mp_2expt(a, ((b -> unit_count - 1) * MONT_UNIT_BIT) + bits - 1)) != MONT_OKAY)
            return res;
    }
    else
    {
        mp_set(a, 1uL);
        bits = 1;
    }
    for (x = bits - 1; x < (int)MONT_UNIT_BIT; x++)
    {
        if((res = mp_mul_2(a, a)) != MONT_OKAY)
            return res;
        if(mont_cmp_mag(a, b) != MONT_CMP_LITTLE)
            if((res = mont_sub_unsigned(a, b, a)) != MONT_OKAY)
                return res;
    }
    return MONT_OKAY;
}

int32_t mont_sqr(const mont_bignum *a, mont_bignum *b)
{
    int32_t olduse, res, pa, ix, iz;
    mont_unit   W[MONT_LIMIT], *tmpx;
    mp_double_unit   W1;
    pa = a -> unit_count + a -> unit_count;
    if(b -> alloc_size < pa)
        if((res = mont_bignum_extend(b, pa)) != MONT_OKAY)
            return res;
    W1 = 0;
    for (ix = 0; ix < pa; ix++)
    {
        int32_t tx, ty, iy;
        mp_double_unit  _W;
        mont_unit *tmpy;
        _W = 0;
        ty = ((a -> unit_count - 1) < ix) ? (a -> unit_count - 1) : ix;
        tx = ix - ty;
        tmpx = a -> data + tx;
        tmpy = a -> data + ty;
        iy = ((a -> unit_count - tx) < (ty + 1)) ? (a -> unit_count - tx) : (ty + 1);
        iy = (iy < (((ty - tx) + 1) >> 1)) ? iy : ((ty - tx) + 1) >> 1;
        for (iz = 0; iz < iy; iz++)
            _W += (mp_double_unit)*tmpx++ * (mp_double_unit)*tmpy--;
        _W = _W + _W + W1;
        if(((unsigned)ix & 1u) == 0u)
            _W += (mp_double_unit)a -> data[ix>>1] * (mp_double_unit)a -> data[ix>>1];
        W[ix] = _W & MONT_MASK;
        W1 = _W >> (mp_double_unit)MONT_UNIT_BIT;
    }
    olduse  = b -> unit_count;
    b -> unit_count = a -> unit_count+a -> unit_count;
    {
        mont_unit *tmpb;
        tmpb = b -> data;
        for (ix = 0; ix < pa; ix++)
            *tmpb++ = W[ix] & MONT_MASK;
        for (; ix < olduse; ix++)
            *tmpb++ = 0;
    }
    mont_bignum_trim_zero(b);
    b -> is_negative = MONT_NO;
    return MONT_OKAY;
}

int32_t mont_mod_exp(const mont_bignum *G, const mont_bignum *X, const mont_bignum *P, mont_bignum *Y)
{
    mont_bignum  M[256], res;
    mont_unit buf, mp;
    int32_t err, bitbuf, bitcpy, bitcnt, mode, digidx, x, y, winsize;
    int32_t (*redux)(mont_bignum *x, const mont_bignum *n, mont_unit rho);
    winsize = 5;
    if((err = mont_bignum_init_size(&M[1], P->alloc_size)) != MONT_OKAY)
        return err;
    for (x = 1<<(winsize-1); x < (1 << winsize); x++)
    {
        if((err = mont_bignum_init_size(&M[x], P->alloc_size)) != MONT_OKAY)
        {
            for (y = 1<<(winsize-1); y < x; y++)
                mont_bignum_free(&M[y]);
            mont_bignum_free(&M[1]);
            return err;
        }
    }
    if((err = montgomery_init(P, &mp)) != MONT_OKAY) goto LBL_M;
    redux = montgomery_redc;
    if((err = mont_bignum_init_size(&res, P->alloc_size)) != MONT_OKAY) goto LBL_M;
    if((err = montgomery_normal(&res, P)) != MONT_OKAY) goto LBL_RES;
    if((err = mont_mod_mul(G, &res, P, &M[1])) != MONT_OKAY) goto LBL_RES;
    if((err = mont_copy(&M[1], &M[1 << (winsize - 1)])) != MONT_OKAY) goto LBL_RES;
    for (x = 0; x < (winsize - 1); x++)
    {
        if((err = mont_sqr(&M[1 << (winsize - 1)], &M[1 << (winsize - 1)])) != MONT_OKAY) goto LBL_RES;
        if((err = redux(&M[1 << (winsize - 1)], P, mp)) != MONT_OKAY) goto LBL_RES;
    }
    for (x = (1 << (winsize - 1)) + 1; x < (1 << winsize); x++)
    {
        if((err = mont_mul(&M[x - 1], &M[1], &M[x])) != MONT_OKAY) goto LBL_RES;
        if((err = redux(&M[x], P, mp)) != MONT_OKAY) goto LBL_RES;
    }
    mode   = 0;
    bitcnt = 1;
    buf    = 0;
    digidx = X->unit_count - 1;
    bitcpy = 0;
    bitbuf = 0;
    for (;;)
    {
        if(--bitcnt == 0)
        {
            if(digidx == -1)
                break;
            buf    = X->data[digidx--];
            bitcnt = (int)MONT_UNIT_BIT;
        }
        y     = (mont_unit)(buf >> (MONT_UNIT_BIT - 1)) & 1;
        buf <<= (mont_unit)1;
        if((mode == 0) && (y == 0))
            continue;
        if((mode == 1) && (y == 0))
        {
            if((err = mont_sqr(&res, &res)) != MONT_OKAY) goto LBL_RES;
            if((err = redux(&res, P, mp)) != MONT_OKAY) goto LBL_RES;
            continue;
        }
        bitbuf |= (y << (winsize - ++bitcpy));
        mode    = 2;
        if(bitcpy == winsize)
        {
            for (x = 0; x < winsize; x++)
            {
                if((err = mont_sqr(&res, &res)) != MONT_OKAY) goto LBL_RES;
                if((err = redux(&res, P, mp)) != MONT_OKAY) goto LBL_RES;
            }
            if((err = mont_mul(&res, &M[bitbuf], &res)) != MONT_OKAY) goto LBL_RES;
            if((err = redux(&res, P, mp)) != MONT_OKAY) goto LBL_RES;
            bitcpy = 0;
            bitbuf = 0;
            mode   = 1;
        }
    }
    if((mode == 2) && (bitcpy > 0))
    {
        for (x = 0; x < bitcpy; x++)
        {
            if((err = mont_sqr(&res, &res)) != MONT_OKAY) goto LBL_RES;
            if((err = redux(&res, P, mp)) != MONT_OKAY) goto LBL_RES;
            bitbuf <<= 1;
            if((bitbuf & (1 << winsize)) != 0)
            {
                if((err = mont_mul(&res, &M[1], &res)) != MONT_OKAY) goto LBL_RES;
                if((err = redux(&res, P, mp)) != MONT_OKAY) goto LBL_RES;
            }
        }
    }
    if((err = redux(&res, P, mp)) != MONT_OKAY) goto LBL_RES;
    mont_exchange(&res, Y);
    err = MONT_OKAY;
LBL_RES:
    mont_bignum_free(&res);
LBL_M:
    mont_bignum_free(&M[1]);
    for (x = 1<<(winsize-1); x < (1 << winsize); x++)
        mont_bignum_free(&M[x]);
    return err;
}

int32_t mont_div_by_2(const mont_bignum *a, mont_bignum *b)
{
    int32_t x, res, oldused;
    if(b -> alloc_size < a -> unit_count)
        if((res = mont_bignum_extend(b, a -> unit_count)) != MONT_OKAY)
            return res;
    oldused = b -> unit_count;
    b -> unit_count = a -> unit_count;
    {
        mont_unit r, rr, *tmpa, *tmpb;
        tmpa = a -> data + b -> unit_count - 1;
        tmpb = b -> data + b -> unit_count - 1;
        r = 0;
        for (x = b -> unit_count - 1; x >= 0; x--)
        {
            rr = *tmpa & 1u;
            *tmpb-- = (*tmpa-- >> 1) | (r << (MONT_UNIT_BIT - 1));
            r = rr;
        }
        tmpb = b -> data + b -> unit_count;
        for (x = b -> unit_count; x < oldused; x++)
            *tmpb++ = 0;
    }
    b -> is_negative = a -> is_negative;
    mont_bignum_trim_zero(b);
    return MONT_OKAY;
}

int32_t mont_cmp_by_a_unit(const mont_bignum *a, mont_unit b)
{
    if(a -> is_negative == MONT_YES)
        return MONT_CMP_LITTLE;
    if(a -> unit_count > 1)
        return MONT_CMP_GREAT;
    if(a -> data[0] > b)
        return MONT_CMP_GREAT;
    else if(a -> data[0] < b)
        return MONT_CMP_LITTLE;
    else
        return MONT_CMP_EQUAL;
}

int32_t mont_mod_inv(const mont_bignum *a, const mont_bignum *b, mont_bignum *c)
{
    mont_bignum  x, y, u, v, B, D;
    int32_t res, neg;
    if(mont_iseven(b) == MONT_YES)
        return MONT_ERROR;
    if((res = mont_bignum_init(&x)) != MONT_OKAY) return res;
    if((res = mont_bignum_init(&y)) != MONT_OKAY) return res;
    if((res = mont_bignum_init(&u)) != MONT_OKAY) return res;
    if((res = mont_bignum_init(&v)) != MONT_OKAY) return res;
    if((res = mont_bignum_init(&B)) != MONT_OKAY) return res;
    if((res = mont_bignum_init(&D)) != MONT_OKAY) return res;
    if((res = mont_copy(b, &x)) != MONT_OKAY)   goto LBL_ERR;
    if((res = mont_mod(a, b, &y)) != MONT_OKAY) goto LBL_ERR;
    if((x.unit_count == 0) || (y.unit_count == 0)) {res = MONT_ERROR;goto LBL_ERR;}
    if((res = mont_copy(&x, &u)) != MONT_OKAY) goto LBL_ERR;
    if((res = mont_copy(&y, &v)) != MONT_OKAY) goto LBL_ERR;
    mp_set(&D, 1uL);
top:
    while(mont_iseven(&u) == MONT_YES)
    {
        if((res = mont_div_by_2(&u, &u)) != MONT_OKAY) goto LBL_ERR;
        if(mont_isodd(&B) == MONT_YES)
            if((res = mont_sub(&B, &x, &B)) != MONT_OKAY) goto LBL_ERR;
        if((res = mont_div_by_2(&B, &B)) != MONT_OKAY) goto LBL_ERR;
    }
    
    while(mont_iseven(&v) == MONT_YES)
    {
        if((res = mont_div_by_2(&v, &v)) != MONT_OKAY) goto LBL_ERR;
        if(mont_isodd(&D) == MONT_YES)
            if((res = mont_sub(&D, &x, &D)) != MONT_OKAY) goto LBL_ERR;
        if((res = mont_div_by_2(&D, &D)) != MONT_OKAY) goto LBL_ERR;
    }
    if(mont_cmp(&u, &v) != MONT_CMP_LITTLE)
    {
        if((res = mont_sub(&u, &v, &u)) != MONT_OKAY) goto LBL_ERR;
        if((res = mont_sub(&B, &D, &B)) != MONT_OKAY) goto LBL_ERR;
    }
    else
    {
        if((res = mont_sub(&v, &u, &v)) != MONT_OKAY) goto LBL_ERR;
        if((res = mont_sub(&D, &B, &D)) != MONT_OKAY) goto LBL_ERR;
    }
    if(u.unit_count != 0) goto top;
    if(mont_cmp_by_a_unit(&v, 1uL) != MONT_CMP_EQUAL)
    {
        res = MONT_ERROR;
        goto LBL_ERR;
    }
    neg = a -> is_negative;
    while(D.is_negative == MONT_YES)
        if((res = mont_add(&D, b, &D)) != MONT_OKAY) goto LBL_ERR;
    while(mont_cmp_mag(&D, b) != MONT_CMP_LITTLE)
        if((res = mont_sub(&D, b, &D)) != MONT_OKAY) goto LBL_ERR;
    mont_exchange(&D, c);
    c -> is_negative = neg;
    res = MONT_OKAY;
LBL_ERR:
    mont_bignum_free(&x);
    mont_bignum_free(&y);
    mont_bignum_free(&u);
    mont_bignum_free(&v);
    mont_bignum_free(&B);
    mont_bignum_free(&D);
    return res;
}







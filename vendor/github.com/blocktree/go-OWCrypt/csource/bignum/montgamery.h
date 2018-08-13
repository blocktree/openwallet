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

#ifndef montgamery_h
#define montgamery_h

#include <stdio.h>
#include <stdlib.h>
#include "type.h"

#define MONT_OKAY   0
#define MONT_ERROR -1

typedef uint64_t mont_unit;
typedef unsigned long mp_double_unit __attribute__((mode(TI)));

typedef struct
{
    mont_unit *data;
    int32_t    unit_count;
    int32_t    alloc_size;
    int32_t    is_negative;
}mont_bignum;

int32_t mont_bignum_init(mont_bignum *a);
void mont_bignum_free(mont_bignum *a);

int32_t mont_bignum_bin2bn(mont_bignum *a, const uint8_t *b, int32_t c);
int32_t mont_bignum_bn2bin(const mont_bignum *a, uint8_t *b);

int32_t mont_mod_mul(const mont_bignum *a, const mont_bignum *b, const mont_bignum *c, mont_bignum *d);
int32_t mont_mod_exp(const mont_bignum *G, const mont_bignum *X, const mont_bignum *P, mont_bignum *Y);
int32_t mont_mod_inv(const mont_bignum *a, const mont_bignum *b, mont_bignum *c);

#endif /* montgamery_h */

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

#ifndef ED25519_h
#define ED25519_h

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "sha512.h"
#include "sha512.h"
#include "ecc_set.h"
#include "type.h"

//point = [scalar]*G
//all in little-endian
void ED25519_point_mul_base(uint8_t *scalar, uint8_t *point);
//point2 = point1 + [scalar]*B
//B for basepoint
//all in little-endian
uint8_t ED25519_point_add_mul_base(uint8_t *point1, uint8_t *scalar, uint8_t *point2);

void ED25519_genPubkey(uint8_t *prikey, uint8_t *pubkey);
void ED25519_Sign(uint8_t *prikey, uint8_t *message, uint16_t message_len, uint8_t *sig);
uint16_t ED25519_Verify(uint8_t *pubkey, uint8_t *message, uint16_t message_len, uint8_t *sig);

void ED25519_get_order(uint8_t *order);

#endif /* ED25519_h */

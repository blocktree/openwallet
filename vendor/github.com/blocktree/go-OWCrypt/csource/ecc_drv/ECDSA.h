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

#ifndef ECDSA_h
#define ECDSA_h

#include <stdio.h>
#include "ecc_drv.h"
#include "ecc_set.h"
#include "bigrand.h"
#include "sha256.h"
#include "type.h"

uint16_t ECDSA_genPubkey(ECC_CURVE_PARAM *curveParam, uint8_t *prikey, ECC_POINT *pubkey);
uint16_t ECDSA_sign(ECC_CURVE_PARAM *curveParam, uint8_t *prikey, uint8_t *message, uint16_t message_len, uint8_t *sig);
uint16_t ECDSA_verify(ECC_CURVE_PARAM *curveParam, ECC_POINT *pubkey, uint8_t *message, uint16_t message_len, uint8_t *sig);

#endif /* ECDSA_h */

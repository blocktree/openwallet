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

#ifndef sha3_256_h
#define sha3_256_h
#include <stdio.h>
#include "type.h"
#define sha3_256_hash_size  32
#define sha3_max_permutation_size 25
#define sha3_max_rate_in_qwords 24
typedef struct
{
    /* 1600 bits algorithm hashing state */
    uint64_t hash[sha3_max_permutation_size];
    /* 1536-bit buffer for leftovers */
    uint64_t message[sha3_max_rate_in_qwords];
    /* count of bytes in the message[] buffer */
    unsigned rest;
    /* size of a message block processed at once */
    unsigned block_size;
} SHA3_256_CTX;

/**
 * Initialize context before calculating hash.
 *
 * @param ctx context to initialize
 */
void sha3_256_init(SHA3_256_CTX *ctx);


/**
 * Calculate message hash.
 * Can be called repeatedly with chunks of the message to be hashed.
 *
 * @param ctx the algorithm context containing current hashing state
 * @param msg message chunk
 * @param msglen length of the message chunk
 */
void sha3_2556_update(SHA3_256_CTX *ctx, const uint8_t *msg, uint32_t msglen);

/**
 * Store calculated hash into the given array.
 *
 * @param ctx the algorithm context containing current hashing state
 * @param digest calculated hash in binary form
 */
void sha3_256_final(SHA3_256_CTX *ctx, uint8_t* digest);

/**
 * keccak256 hash.
 *
 * @param msg the message to do hash
 * @param msg_len the length of message
 * @param digest hash result
 */
void sha3_256_hash(const uint8_t *msg,uint32_t msg_len,uint8_t *digest);

#endif /* sha3_256_h */

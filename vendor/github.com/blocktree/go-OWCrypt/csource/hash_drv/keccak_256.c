
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
#include "keccak_256.h"

/* constants */
#define KECCAK256_NumberOfRounds 24

#define KECCAK256_ROTL64(qword, n) ((qword) << (n) ^ ((qword) >> (64 - (n))))

#define KECCAK256_IS_ALIGNED_64(p) (0 == (7 & ((const char*)(p) - (const char*)0)))

#define keccak256_me64_to_le_str(to, from, length) memcpy((to), (from), (length))

# define keccak_le2me_64(x) (x)

/* SHA3 (Keccak) constants for 24 rounds */
static uint64_t keccak_round_constants[KECCAK256_NumberOfRounds] = {
    (0x0000000000000001), (0x0000000000008082), (0x800000000000808A), (0x8000000080008000),
    (0x000000000000808B), (0x0000000080000001), (0x8000000080008081), (0x8000000000008009),
    (0x000000000000008A), (0x0000000000000088), (0x0000000080008009), (0x000000008000000A),
    (0x000000008000808B), (0x800000000000008B), (0x8000000000008089), (0x8000000000008003),
    (0x8000000000008002), (0x8000000000000080), (0x000000000000800A), (0x800000008000000A),
    (0x8000000080008081), (0x8000000000008080), (0x0000000080000001), (0x8000000080008008)
};

/* Initializing a sha3 context for given number of output bits */
static void rhash_keccak_init(KECCAK256_CTX *ctx, unsigned bits)
{
    /* NB: The Keccak capacity parameter = bits * 2 */
    unsigned rate = 1600 - bits * 2;
    
    memset(ctx, 0, sizeof(KECCAK256_CTX));
    ctx->block_size = rate / 8;
    assert(rate <= 1600 && (rate % 64) == 0);
}

/**
 * Initialize context before calculating hash.
 *
 * @param ctx context to initialize
 */
void keccak256_init(KECCAK256_CTX *ctx)
{
    rhash_keccak_init(ctx, 256);
}

#define XORED_A(i) A[(i)] ^ A[(i) + 5] ^ A[(i) + 10] ^ A[(i) + 15] ^ A[(i) + 20]
#define THETA_STEP(i) \
A[(i)]      ^= D[(i)]; \
A[(i) + 5]  ^= D[(i)]; \
A[(i) + 10] ^= D[(i)]; \
A[(i) + 15] ^= D[(i)]; \
A[(i) + 20] ^= D[(i)] \

/* Keccak theta() transformation */
static void keccak_theta(uint64_t *A)
{
    uint64_t D[5];
    D[0] = KECCAK256_ROTL64(XORED_A(1), 1) ^ XORED_A(4);
    D[1] = KECCAK256_ROTL64(XORED_A(2), 1) ^ XORED_A(0);
    D[2] = KECCAK256_ROTL64(XORED_A(3), 1) ^ XORED_A(1);
    D[3] = KECCAK256_ROTL64(XORED_A(4), 1) ^ XORED_A(2);
    D[4] = KECCAK256_ROTL64(XORED_A(0), 1) ^ XORED_A(3);
    THETA_STEP(0);
    THETA_STEP(1);
    THETA_STEP(2);
    THETA_STEP(3);
    THETA_STEP(4);
}

/* Keccak pi() transformation */
static void keccak_pi(uint64_t *A)
{
    uint64_t A1;
    A1 = A[1];
    A[ 1] = A[ 6];
    A[ 6] = A[ 9];
    A[ 9] = A[22];
    A[22] = A[14];
    A[14] = A[20];
    A[20] = A[ 2];
    A[ 2] = A[12];
    A[12] = A[13];
    A[13] = A[19];
    A[19] = A[23];
    A[23] = A[15];
    A[15] = A[ 4];
    A[ 4] = A[24];
    A[24] = A[21];
    A[21] = A[ 8];
    A[ 8] = A[16];
    A[16] = A[ 5];
    A[ 5] = A[ 3];
    A[ 3] = A[18];
    A[18] = A[17];
    A[17] = A[11];
    A[11] = A[ 7];
    A[ 7] = A[10];
    A[10] = A1;
    /* note: A[ 0] is left as is */
}

#define CHI_STEP(i) \
A0 = A[0 + (i)]; \
A1 = A[1 + (i)]; \
A[0 + (i)] ^= ~A1 & A[2 + (i)]; \
A[1 + (i)] ^= ~A[2 + (i)] & A[3 + (i)]; \
A[2 + (i)] ^= ~A[3 + (i)] & A[4 + (i)]; \
A[3 + (i)] ^= ~A[4 + (i)] & A0; \
A[4 + (i)] ^= ~A0 & A1 \

/* Keccak chi() transformation */
static void keccak_chi(uint64_t *A)
{
    uint64_t A0, A1;
    CHI_STEP(0);
    CHI_STEP(5);
    CHI_STEP(10);
    CHI_STEP(15);
    CHI_STEP(20);
}

static void rhash_sha3_permutation(uint64_t *state)
{
    int round;
    for (round = 0; round < KECCAK256_NumberOfRounds; round++)
    {
        keccak_theta(state);
        
        /* apply Keccak rho() transformation */
        state[ 1] = KECCAK256_ROTL64(state[ 1],  1);
        state[ 2] = KECCAK256_ROTL64(state[ 2], 62);
        state[ 3] = KECCAK256_ROTL64(state[ 3], 28);
        state[ 4] = KECCAK256_ROTL64(state[ 4], 27);
        state[ 5] = KECCAK256_ROTL64(state[ 5], 36);
        state[ 6] = KECCAK256_ROTL64(state[ 6], 44);
        state[ 7] = KECCAK256_ROTL64(state[ 7],  6);
        state[ 8] = KECCAK256_ROTL64(state[ 8], 55);
        state[ 9] = KECCAK256_ROTL64(state[ 9], 20);
        state[10] = KECCAK256_ROTL64(state[10],  3);
        state[11] = KECCAK256_ROTL64(state[11], 10);
        state[12] = KECCAK256_ROTL64(state[12], 43);
        state[13] = KECCAK256_ROTL64(state[13], 25);
        state[14] = KECCAK256_ROTL64(state[14], 39);
        state[15] = KECCAK256_ROTL64(state[15], 41);
        state[16] = KECCAK256_ROTL64(state[16], 45);
        state[17] = KECCAK256_ROTL64(state[17], 15);
        state[18] = KECCAK256_ROTL64(state[18], 21);
        state[19] = KECCAK256_ROTL64(state[19],  8);
        state[20] = KECCAK256_ROTL64(state[20], 18);
        state[21] = KECCAK256_ROTL64(state[21],  2);
        state[22] = KECCAK256_ROTL64(state[22], 61);
        state[23] = KECCAK256_ROTL64(state[23], 56);
        state[24] = KECCAK256_ROTL64(state[24], 14);
        
        keccak_pi(state);
        keccak_chi(state);
        
        /* apply iota(state, round) */
        *state ^= keccak_round_constants[round];
    }
}

/**
 * The core transformation. Process the specified block of data.
 *
 * @param hash the algorithm state
 * @param block the message block to process
 * @param block_size the size of the processed block in bytes
 */
static void rhash_sha3_process_block(uint64_t hash[25], const uint64_t *block, uint32_t block_size)
{
    /* expanded loop */
    hash[ 0] ^= keccak_le2me_64(block[ 0]);
    hash[ 1] ^= keccak_le2me_64(block[ 1]);
    hash[ 2] ^= keccak_le2me_64(block[ 2]);
    hash[ 3] ^= keccak_le2me_64(block[ 3]);
    hash[ 4] ^= keccak_le2me_64(block[ 4]);
    hash[ 5] ^= keccak_le2me_64(block[ 5]);
    hash[ 6] ^= keccak_le2me_64(block[ 6]);
    hash[ 7] ^= keccak_le2me_64(block[ 7]);
    hash[ 8] ^= keccak_le2me_64(block[ 8]);
    /* if not sha3-512 */
    if (block_size > 72) {
        hash[ 9] ^= keccak_le2me_64(block[ 9]);
        hash[10] ^= keccak_le2me_64(block[10]);
        hash[11] ^= keccak_le2me_64(block[11]);
        hash[12] ^= keccak_le2me_64(block[12]);
        /* if not sha3-384 */
        if (block_size > 104) {
            hash[13] ^= keccak_le2me_64(block[13]);
            hash[14] ^= keccak_le2me_64(block[14]);
            hash[15] ^= keccak_le2me_64(block[15]);
            hash[16] ^= keccak_le2me_64(block[16]);
            /* if not sha3-256 */
            if (block_size > 136) {
                hash[17] ^= keccak_le2me_64(block[17]);
            }
        }
    }
    /* make a permutation of the hash */
    rhash_sha3_permutation(hash);
}

#define KECCAK256_FINALIZED 0x80000000

/**
 * Calculate message hash.
 * Can be called repeatedly with chunks of the message to be hashed.
 *
 * @param ctx the algorithm context containing current hashing state
 * @param msg message chunk
 * @param msg_len length of the message chunk
 */
void keccak256_update(KECCAK256_CTX *ctx, const uint8_t *msg, uint32_t msg_len)
{
    size_t index = (size_t)ctx->rest;
    size_t block_size = (size_t)ctx->block_size;
    
    if (ctx->rest & KECCAK256_FINALIZED) return; /* too late for additional input */
    ctx->rest = (unsigned)((ctx->rest + msg_len) % block_size);
    
    /* fill partial block */
    if (index) {
        size_t left = block_size - index;
        memcpy((char*)ctx->message + index, msg, (msg_len < left ? msg_len : left));
        if (msg_len < left) return;
        
        /* process partial block */
        rhash_sha3_process_block(ctx->hash, ctx->message, (uint32_t)block_size);
        msg  += left;
        msg_len -= left;
    }
    while (msg_len >= block_size) {
        uint64_t* aligned_message_block;
        if (KECCAK256_IS_ALIGNED_64(msg)) {
            /* the most common case is processing of an already aligned message
             without copying it */
            aligned_message_block = (uint64_t*)msg;
        } else {
            memcpy(ctx->message, msg, block_size);
            aligned_message_block = ctx->message;
        }
        
        rhash_sha3_process_block(ctx->hash, aligned_message_block, (uint32_t)block_size);
        msg  += block_size;
        msg_len -= block_size;
    }
    if (msg_len) {
        memcpy(ctx->message, msg, msg_len); /* save leftovers */
    }
}

/**
 * Store calculated hash into the given array.
 *
 * @param ctx the algorithm context containing current hashing state
 * @param digest calculated hash in binary form
 */
void keccak256_final(KECCAK256_CTX *ctx, uint8_t* digest)
{
    size_t digest_length = 100 - ctx->block_size / 2;
    const size_t block_size = ctx->block_size;
    if (!(ctx->rest & KECCAK256_FINALIZED))
    {
        /* clear the rest of the data queue */
        memset((char*)ctx->message + ctx->rest, 0, block_size - ctx->rest);
        ((char*)ctx->message)[ctx->rest] |= 0x01;
        ((char*)ctx->message)[block_size - 1] |= 0x80;
        
        /* process final block */
        rhash_sha3_process_block(ctx->hash, ctx->message, (uint32_t)block_size);
        ctx->rest = KECCAK256_FINALIZED; /* mark context as finalized */
    }
    assert(block_size > digest_length);
    if (digest) keccak256_me64_to_le_str(digest, ctx->hash, digest_length);
}

/**
 * keccak256 hash.
 *
 * @param msg the message to do hash
 * @param msg_len the length of message
 * @param digest hash result
 */
void keccak256_hash(const uint8_t *msg,uint32_t msg_len,uint8_t *digest)
{
    KECCAK256_CTX ctx;
    keccak256_init(&ctx);
    keccak256_update(&ctx, msg, msg_len);
    keccak256_final(&ctx, digest);
}

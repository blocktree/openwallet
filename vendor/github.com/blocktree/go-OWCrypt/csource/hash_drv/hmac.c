
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
#include "hmac.h"

/*
 @function:paddding key,make sure the length of key ids same as block size and XOR with ipad or opad
 @parameter[in]:K pointer to key to be padded
 @paramter[in]:Klen,the byte length of K
 @paramter[in]:pad pointer ipad or opad
 @paramter[out]:K padding result
 @return:NULL
 @notice:(1)the input and output can be the same buffer
 */
static void padkey(uint8_t *K,uint8_t *pad,uint16_t len,uint8_t *out)
{
    uint16_t i;
    if(!len)
        return;
    if(!K || !pad || !out)
        return;
    for(i=0;i<len;i++)
    {
        out[i] = K[i]^pad[i];
    }
    return;
}
/*
 @function:compute massage authentication code
 @paramter[in]:K pointer to key
 @paramter[in]:Klen,the byte length of K
 @paramter[in]:M pointer to message to be authenticated
 @paramter[in]:Mlen,the byte length of M
 @paramter[out]:out pointer to HMAC result
 @paramter[in]:Hash_Alg,hash algorithm flag.if Hash_Alg = MD5_ALG,Choose MD5 algorithm;if Hash_Alg = SHA1_ALG,choose SHA1 algorithm;
 Hash_Alg=SHA256_ALG,Choose SHA256 algorithm;if Hash_Alg=SHA512_ALG, Choose SHA512 algorithm;if Hash_Alg=SM3_ALG, Choose SM3 algorithm;if HAsh_Alg=BLAKE2B_ALG, choose BLAKE2B algorithm;if HAsh_Alg=BLAKE2S,choose BLAKE2S algorithm;default:not support.
 @return:NULL
 @notice:if Hash_Alg=MD5_ALG,the space size of out is 16 byte; if Hash_Alg=SHA1_ALG,the space size of out is 20 byte; if Hash_Alg=SHA256_ALG,the space size of out is 32 byte;if Hash_Alg=SHA512_ALG,the space size of out is 64 byte; if Hash_Alg=SM3_ALG,the space size of out is 32 byte;
 */
void HMAC(uint8_t *K,uint16_t Klen,uint8_t *M,uint16_t Mlen,uint8_t *out,uint8_t Hash_Alg)
{
    uint16_t blockByte;
    uint8_t *KCopy=NULL,*ipad=NULL,*opad=NULL,*temp_result=NULL,*keypad=NULL;
    SHA1_CTX *sha1_ctx=NULL;
    MD5_CTX *md5_ctx=NULL;
    SHA256_CTX *sha256_ctx=NULL;
    SHA512_CTX *sha512_ctx=NULL;
    SM3_CTX *sm3_ctx=NULL;
    RIPEMD160_CTX *ripemd160_ctx=NULL;
    BLAKE2B_CTX *blake2b_ctx=NULL;
    BLAKE2S_CTX *blake2s_ctx=NULL;
    if(!K || !M || !out)
        return;
    if(!Klen || !Mlen)
        return;
    if((Hash_Alg == MD5_ALG) || (Hash_Alg==SHA1_ALG)||(Hash_Alg==SHA256_ALG)||(Hash_Alg == SM3_ALG)||(Hash_Alg == RIPEMD160) || (Hash_Alg == BLAKE2S))
    {
        blockByte = 512 >> 3;
    }
    else if((Hash_Alg == SHA512_ALG) ||(Hash_Alg == BLAKE2B))
    {
        blockByte = 1024 >> 3;
    }
    else
    {
        return;
    }
    KCopy=calloc(blockByte,sizeof(uint8_t));
    ipad = calloc(blockByte,sizeof(uint8_t));
    opad = calloc(blockByte,sizeof(uint8_t));
    keypad = calloc(blockByte,sizeof(uint8_t));
    memset(ipad,0x36,blockByte);
    memset(opad,0x5c,blockByte);
    switch(Hash_Alg)
    {
        case RIPEMD160:
            ripemd160_ctx=calloc(1,sizeof( RIPEMD160_CTX));
            temp_result = calloc(20,sizeof(uint8_t));
            if(Klen > blockByte)
            {
                ripemd160_hash(K, Klen,KCopy);
            }
            else
            {
                memcpy(KCopy,K,Klen);
            }
            padkey(KCopy,ipad,blockByte,keypad);
            ripemd160_init(ripemd160_ctx);
            ripemd160_update(ripemd160_ctx,keypad,blockByte);
            ripemd160_update(ripemd160_ctx,M,Mlen);
            ripemd160_final(ripemd160_ctx,temp_result);
            padkey(KCopy,opad,blockByte,keypad);
            ripemd160_init(ripemd160_ctx);
            ripemd160_update(ripemd160_ctx,keypad,blockByte);
            ripemd160_update(ripemd160_ctx,temp_result,20);
            ripemd160_final(ripemd160_ctx,out);
            free(ripemd160_ctx);
            break;
        case MD5_ALG:
            md5_ctx=calloc(1,sizeof( MD5_CTX));
            temp_result = calloc(16,sizeof(uint8_t));
            if(Klen > blockByte)
            {
                md5_hash(K, Klen,KCopy);
            }
            else
            {
                memcpy(KCopy,K,Klen);
            }
            padkey(KCopy,ipad,blockByte,keypad);
            md5_init (md5_ctx);
            md5_update(md5_ctx, keypad, blockByte);
            md5_update(md5_ctx, M, Mlen);
            md5_final (md5_ctx,temp_result);
            padkey(KCopy,opad,blockByte,keypad);
            md5_init (md5_ctx);
            md5_update(md5_ctx, keypad, blockByte);
            md5_update(md5_ctx, temp_result, 16);
            md5_final (md5_ctx,out);
            free(md5_ctx);
            break;
        case SHA1_ALG:
            sha1_ctx=calloc(1,sizeof( SHA1_CTX));
            temp_result = calloc(20,sizeof(uint8_t));
            if(Klen > blockByte)
            {
                sha1_hash(K, Klen,KCopy);
            }
            else
            {
                memcpy(KCopy,K,Klen);
            }
            padkey(KCopy,ipad,blockByte,keypad);
            sha1_init (sha1_ctx);
            sha1_update(sha1_ctx, keypad, blockByte);
            sha1_update(sha1_ctx, M, Mlen);
            sha1_final (sha1_ctx,temp_result);
            padkey(KCopy,opad,blockByte,keypad);
            sha1_init (sha1_ctx);
            sha1_update(sha1_ctx, keypad, blockByte);
            sha1_update(sha1_ctx, temp_result, 20);
            sha1_final (sha1_ctx,out);
            free(sha1_ctx);
            break;
        case SHA256_ALG:
            sha256_ctx=calloc(1,sizeof(SHA256_CTX));
            temp_result = calloc(32,sizeof(uint8_t));
            if(Klen > blockByte)
            {
                sha256_hash(K, Klen,KCopy);
            }
            else
            {
                memcpy(KCopy,K,Klen);
            }
            padkey(KCopy,ipad,blockByte,keypad);
            sha256_init (sha256_ctx);
            sha256_update(sha256_ctx, keypad, blockByte);
            sha256_update(sha256_ctx, M, Mlen);
            sha256_final (sha256_ctx,temp_result);
            padkey(KCopy,opad,blockByte,keypad);
            sha256_init (sha256_ctx);
            sha256_update(sha256_ctx, keypad, blockByte);
            sha256_update(sha256_ctx, temp_result, 32);
            sha256_final (sha256_ctx,out);
            free(sha256_ctx);
            break;
        case SHA512_ALG:
            sha512_ctx=calloc(1,sizeof( SHA512_CTX));
            temp_result = calloc(64,sizeof(uint8_t));
            if(Klen > blockByte)
            {
                sha512_hash(K, Klen,KCopy);
            }
            else
            {
                memcpy(KCopy,K,Klen);
            }
            padkey(KCopy,ipad,blockByte,keypad);
            sha512_init (sha512_ctx);
            sha512_update(sha512_ctx, keypad, blockByte);
            sha512_update(sha512_ctx, M, Mlen);
            sha512_final (sha512_ctx,temp_result);
            padkey(KCopy,opad,blockByte,keypad);
            sha512_init (sha512_ctx);
            sha512_update(sha512_ctx, keypad, blockByte);
            sha512_update(sha512_ctx, temp_result, 64);
            sha512_final (sha512_ctx,out);
            free(sha512_ctx);
            break;
        case SM3_ALG:
            sm3_ctx=calloc(1,sizeof(SM3_CTX));
            temp_result = calloc(32,sizeof(uint8_t));
            if(Klen > blockByte)
            {
                sm3_hash(K, Klen,KCopy);
            }
            else
            {
                memcpy(KCopy,K,Klen);
            }
            padkey(KCopy,ipad,blockByte,keypad);
            sm3_init (sm3_ctx);
            sm3_update(sm3_ctx, keypad, blockByte);
            sm3_update(sm3_ctx, M, Mlen);
            sm3_final (sm3_ctx,temp_result);
            padkey(KCopy,opad,blockByte,keypad);
            sm3_init (sm3_ctx);
            sm3_update(sm3_ctx, keypad, blockByte);
            sm3_update(sm3_ctx, temp_result, 32);
            sm3_final (sm3_ctx,out);
            free(sm3_ctx);
            break;
        case BLAKE2B:
            blake2b_ctx=calloc(1,sizeof(BLAKE2B_CTX));
            temp_result = calloc(BLAKE2B_DIGEST_LENGTH,sizeof(uint8_t));
            if(Klen > blockByte)
            {
                //sm3_hash(K, Klen,KCopy);
                blake2b(K, Klen,NULL,0, BLAKE2B_DIGEST_LENGTH, KCopy);
            }
            else
            {
                memcpy(KCopy,K,Klen);
            }
            padkey(KCopy,ipad,blockByte,keypad);
            blake2b_init(blake2b_ctx, NULL,0,BLAKE2B_DIGEST_LENGTH);
            blake2b_update(blake2b_ctx, keypad, blockByte);
            blake2b_update(blake2b_ctx, M, Mlen);
            blake2b_final(blake2b_ctx,out, BLAKE2B_DIGEST_LENGTH);
            padkey(KCopy,opad,blockByte,keypad);
            blake2b_init(blake2b_ctx, NULL,0,BLAKE2B_DIGEST_LENGTH);
            blake2b_update(blake2b_ctx, keypad, blockByte);
            blake2b_update(blake2b_ctx, temp_result, BLAKE2B_DIGEST_LENGTH);
            blake2b_final(blake2b_ctx,out, BLAKE2B_DIGEST_LENGTH);
            free(blake2b_ctx);
            break;
        case BLAKE2S:
            blake2s_ctx=calloc(1,sizeof(BLAKE2S_CTX));
            temp_result = calloc(BLAKE2S_DIGEST_LENGTH,sizeof(uint8_t));
            if(Klen > blockByte)
            {
                //sm3_hash(K, Klen,KCopy);
                blake2s(K, Klen,NULL,0, BLAKE2S_DIGEST_LENGTH, KCopy);
            }
            else
            {
                memcpy(KCopy,K,Klen);
            }
            padkey(KCopy,ipad,blockByte,keypad);
            blake2s_init(blake2s_ctx, NULL,0,BLAKE2S_DIGEST_LENGTH);
            blake2s_update(blake2s_ctx, keypad, blockByte);
            blake2s_update(blake2s_ctx, M, Mlen);
            blake2s_final(blake2s_ctx,out, BLAKE2S_DIGEST_LENGTH);
            padkey(KCopy,opad,blockByte,keypad);
            blake2s_init(blake2s_ctx, NULL,0,BLAKE2S_DIGEST_LENGTH);
            blake2s_update(blake2s_ctx, keypad, blockByte);
            blake2s_update(blake2s_ctx, temp_result, BLAKE2S_DIGEST_LENGTH);
            blake2s_final(blake2s_ctx,out, BLAKE2S_DIGEST_LENGTH);
            free(blake2s_ctx);
            break;
        default:
            break;
    }
    free(KCopy);
    free(ipad);
    free(opad);
    free(keypad);
    free(temp_result);
    return;
}

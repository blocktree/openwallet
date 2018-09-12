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

#include "ecc_drv.h"

/** 1 for legal;0 for illegal **/

/*
 @function:transfer byte string to point
 @paramter[in]:src pointer to source address(byte string)
 @paramter[out]:dst pointer to destination address(point)
 @return: 1 denotes success; 0 denotes fail
 */
static uint8_t byte_to_point(uint8_t *src,ECC_POINT *dst)
{
    if(!src || !dst)
    {
        return 0;
    }
    memcpy(dst->x,src,ECC_LEN);
    memcpy(dst->y,src+ECC_LEN,ECC_LEN);
    //make sure the input point is not infinity
    dst->infinity = 0;
    return 1;
}
/*
 @function:transfer point to byte string
 @paramter[in]:src pointer to source address(point)
 @paramter[out]:dst pointer to destination address(byte string)
 @return: 1 denotes success; 0 denotes fail
 */
static uint8_t point_to_byte(ECC_POINT *src,uint8_t *dst)
{
    if(!dst || !src)
    {
        return 0;
    }
    memcpy(dst,src->x,ECC_LEN);
    memcpy(dst + ECC_LEN,src->y,ECC_LEN);
    return 1;
}
uint8_t is_prikey_legal(ECC_CURVE_PARAM *curveParam, uint8_t *prikey)
{
    if(is_all_zero(prikey, ECC_LEN) || memcmp(prikey, curveParam -> n, ECC_LEN) >= 0)
        return 0;
    else
        return 1;
}

uint8_t is_neg_y(uint8_t *y1, uint8_t *y2, uint8_t *p)
{
    uint8_t *tmp = NULL;
    tmp = calloc(ECC_LEN, sizeof(uint8_t));
    
    bignum_mod_add(y1, y2, p, tmp);
    
    if(is_all_zero(tmp, ECC_LEN))
    {
        free(tmp);
        return 1;
    }
    else
    {
        free(tmp);
        return 0;
    }
}

uint8_t get_lambda(ECC_CURVE_PARAM *curveParam, ECC_POINT *point1, ECC_POINT *point2, uint8_t *lambda)
{
    uint8_t *tmp1 = NULL, *tmp2 = NULL;

    if(memcmp(point1 -> x, point2 -> x, ECC_LEN)) //x1 != x2
    {
        tmp1 = calloc(ECC_LEN, sizeof(uint8_t));
        tmp2 = calloc(ECC_LEN, sizeof(uint8_t));
        
        bignum_mod_sub(point2 -> y, point1 -> y, curveParam -> p, tmp1);
        bignum_mod_sub(point2 -> x, point1 -> x, curveParam -> p, tmp2);
        
        bignum_mod_inv(tmp2, curveParam -> p, tmp2);
        bignum_mod_mul(tmp1, tmp2, curveParam -> p, lambda);
        
        free(tmp1);
        free(tmp2);
    }
    else //x1 == x2
    {
        if(is_neg_y(point1 -> y, point2 -> y, curveParam -> p)) //y1 == -y2
        {
            return 1;
        }
        else //y1 == y2
        {
            tmp1 = calloc(ECC_LEN, sizeof(uint8_t));
            tmp2 = calloc(ECC_LEN, sizeof(uint8_t));
            
            bignum_mod_mul(point1 -> x, point1 -> x, curveParam -> p, tmp1);
            bignum_mod_add(tmp1, tmp1, curveParam -> p, tmp2);
            bignum_mod_add(tmp1, tmp2, curveParam -> p, tmp1);
            bignum_mod_add(tmp1, curveParam -> a, curveParam -> p, tmp1);
            bignum_mod_add(point1 -> y, point1 -> y, curveParam -> p, tmp2);
            bignum_mod_inv(tmp2, curveParam -> p, tmp2);
            bignum_mod_mul(tmp1, tmp2, curveParam -> p, lambda);
            
            free(tmp1);
            free(tmp2);
        }
    }
    
    return 0;
}


//point = point1 + point2
//return 1 : infinity point
uint8_t point_add(ECC_CURVE_PARAM *curveParam, ECC_POINT *point1, ECC_POINT *point2, ECC_POINT *point)
{
    uint8_t *lambda = NULL, *tmp = NULL;
    ECC_POINT *point_tmp = NULL;
    
    if(point1 -> infinity && !point2 -> infinity)
    {
        memcpy((uint8_t *)point, (uint8_t *)point2, sizeof(ECC_POINT));
        return 0;
    }
    if(!point1 -> infinity && point2 -> infinity)
    {
        memcpy((uint8_t *)point, (uint8_t *)point1, sizeof(ECC_POINT));
        return 0;
    }
    if(point1 -> infinity && point2 -> infinity)
    {
        point -> infinity = 1;
        return 1;
    }
    
    lambda = calloc(ECC_LEN, sizeof(uint8_t));
    
    if(get_lambda(curveParam, point1, point2, lambda))
    {
        free(lambda);
        point -> infinity = 1;
        return 1;
    }
    tmp = calloc(ECC_LEN, sizeof(uint8_t));
    point_tmp = calloc(1, sizeof(ECC_POINT));
    
    bignum_mod_mul(lambda, lambda, curveParam -> p, tmp);
    bignum_mod_sub(tmp, point1 -> x, curveParam -> p, tmp);
    bignum_mod_sub(tmp, point2 -> x, curveParam -> p, point_tmp -> x);
    
    bignum_mod_sub(point1 -> x, point_tmp -> x, curveParam -> p, tmp);
    bignum_mod_mul(lambda, tmp, curveParam -> p, tmp);
    bignum_mod_sub(tmp, point1 -> y, curveParam -> p, point_tmp -> y);
    
    memcpy((uint8_t *)point, (uint8_t *)point_tmp, sizeof(ECC_POINT));
    
    free(lambda);
    free(tmp);
    free(point_tmp);
    return 0;
}

//point_out= [k]point_in
//return 1 : infinity point
//二进制展开
uint8_t point_mul(ECC_CURVE_PARAM *curveParam, ECC_POINT *point_in, uint8_t *k, ECC_POINT *point_out)
{
    uint16_t bit_len = 0;
    int16_t i = 0;
    ECC_POINT *point_tmp = NULL;
    
    point_tmp = calloc(1, sizeof(ECC_POINT));
    
    memcpy((uint8_t *)point_tmp, (uint8_t *)point_in, sizeof(ECC_POINT));
    
    bit_len = get_bit_len(k, ECC_LEN);
    
    if(bit_len == 0)
        return 0;
    
    point_out -> infinity = 1;
    
    for(i = bit_len - 1; i >= 0; i --)
    {
        point_add(curveParam, point_out, point_out, point_out);
        if(get_bit_value(k, ECC_LEN, i))
            point_add(curveParam, point_tmp, point_out, point_out);
    }
    return point_out -> infinity;
}

//1 for legal
//0 for illegal
uint8_t is_pubkey_legal(ECC_CURVE_PARAM *curveParam, ECC_POINT *point)
{
    uint8_t *tmp1 = NULL, *tmp2 = NULL;
    ECC_POINT *point_tmp = NULL;
    
    if(memcmp(point -> x, curveParam -> p, ECC_LEN) >= 0 || memcmp(point -> y, curveParam -> p, ECC_LEN) >= 0)
        return 0;
    
    tmp1 = calloc(ECC_LEN, sizeof(uint8_t));
    tmp2 = calloc(ECC_LEN, sizeof(uint8_t));
    
    bignum_mod_mul(point -> x, point -> x, curveParam -> p, tmp1);
    bignum_mod_mul(point -> x, tmp1, curveParam -> p, tmp2);
    bignum_mod_mul(curveParam -> a, point -> x, curveParam -> p, tmp1);
    bignum_mod_add(tmp1, tmp2, curveParam -> p, tmp1);
    bignum_mod_add(tmp1, curveParam -> b, curveParam -> p, tmp1);
    
    bignum_mod_mul(point -> y, point -> y, curveParam -> p, tmp2);
    
    if(memcmp(tmp1, tmp2, ECC_LEN))
    {
        free(tmp1);
        free(tmp2);
        return 0;
    }
    
    point_tmp = calloc(1, sizeof(ECC_POINT));
    
    if(!point_mul(curveParam, point, curveParam -> n, point_tmp))
    {
        free(tmp1);
        free(tmp2);
        free(point_tmp);
        return 0;
    }
    
    free(tmp1);
    free(tmp2);
    free(point_tmp);
    
    return 1;
}
/*
 @function:(Point)outpoint_buf = (Point)inputpoint1_buf +[k](Point)inputpoint2_buf
 @paramter[in]:curveParam pointer to curve elliptic paremters
 @paramter[in]:inputpoint1_buf pointer to one point on the curve elliptic(stroreed by byte string)
 @paramter[in]:Q pointer to another point on the elliptic(stored by byte string)
 @paramter[in]:k pointer to the multiplicator
 @paramter[out]:outpoint_buf pointer to the result((Point)outpoint_buf:=(Point)inputpoint1_buf +[k](Point)inputpoint2_buf)
 */
uint8_t point_mul_add(ECC_CURVE_PARAM *curveParam,uint8_t *inputpoint1_buf,uint8_t *inputpoint2_buf,uint8_t *k,uint8_t *outpoint_buf)
{
    uint16_t ret;
    ECC_POINT *P=NULL,*Q=NULL,*T=NULL;
    P = calloc(1,sizeof(ECC_POINT));
    Q = calloc(1,sizeof(ECC_POINT));
    T = calloc(1,sizeof(ECC_POINT));
    byte_to_point(inputpoint2_buf,Q);
    ret=point_mul(curveParam, Q, k, T);
    if(ret)
    {
        return ret;
    }
    byte_to_point(inputpoint1_buf,P);
    ret=point_add(curveParam, P, T, Q);
    if(ret)
    {
        return ret;
    }
    point_to_byte(Q,outpoint_buf);
    free(P);
    free(Q);
    free(T);
    return 0;
}


/*
 @function:点的压缩
 @paramter[in]:point_buf,待压缩的点
 @paramter[in]:point_buf_len表示point_buf的字节长度
 @paramter[in]:x,点压缩后的横坐标（长度为ECC_LEN+1 字节）
 @return：1，压缩失败；0:压缩成功
 */
uint8_t point_compress(uint8_t *point_buf,uint16_t point_buf_len,uint8_t *x)
{
    if(point_buf_len ==((ECC_LEN<<1) + 1))
    {
        if(point_buf[0]!=0x04)
            return 0;
    }
    else if(point_buf_len == (ECC_LEN<<1))
    {
        ;
    }
    else
    {
        return 0;
    }
    if(point_buf_len == ((ECC_LEN<<1)+1))
    {
        if(point_buf[(ECC_LEN << 1)]&0x01)
        {
            x[0]=0x03;
            memcpy(x + 1,point_buf+1,(point_buf_len-1)>>1);
        }
        else
        {
            x[0]=0x02;
            memcpy(x+1,point_buf + 1,(point_buf_len-1)>>1);
        }
    }
    else
    {
        if(point_buf[(ECC_LEN << 1)-1]&0x01)
        {
            x[0]=0x03;
            memcpy(x + 1,point_buf,point_buf_len>>1);
        }
        else
        {
            x[0]=0x02;
            memcpy(x+1,point_buf,point_buf_len>>1);
        }
    }
    return 1;
}
/*
 @function:点的解压缩：根据曲线参数curveParam和x坐标，求解y坐标(满足曲线方程y^2=x^3+a*x+b)
 @paramter[in]:curveParam,椭圆曲线方程参数
 @paramter[in]:x,曲线上点的横坐标（第一个字节为0x02或0x03.0x02表示y为偶数；0x03表示y为奇数）
 @paramter[in]:x_len表示x的字节长度（一个字节的表示符 + ECC_LEN 字节的私钥）
 @paramter[out]:point_buf,待求解的曲线上的点（含0x04）
 @return:1,表示输入的数据格式错误或者求解y时，平方根不存在;0:表示解压缩成功
 @note：(1)输入的x坐标一定带有标示字节（第一个字节）0x02:表示y为偶数；0x03表示y为奇数.(2)目前支持（p =3(mod4)和p=5(mod8)两种情况）
 */

uint8_t point_decompress(ECC_CURVE_PARAM *curveParam, uint8_t *x,uint16_t x_len,uint8_t *point_buf)
{
    uint8_t *tmp1 = NULL, *tmp2 = NULL,*tmp3=NULL,*tmp4=NULL;
    tmp1 = calloc(ECC_LEN, sizeof(uint8_t));
    tmp2 = calloc(ECC_LEN, sizeof(uint8_t));
    tmp3 = calloc(ECC_LEN, sizeof(uint8_t));
    tmp4 = calloc(ECC_LEN, sizeof(uint8_t));
    if(x_len != (ECC_LEN + 1))
    {
        return 0;
    }
    if((x[0] != 0x02)&&(x[0] != 0x03))
    {
        return 0;
    }
    //求解tmp1 = x^2
    bignum_mod_mul(x+1, x+1, curveParam -> p, tmp1);
    //求解tmp2 = x^3
    bignum_mod_mul(x+1, tmp1, curveParam -> p, tmp2);
    //求解 tmp1 = a*x (mod q)
    bignum_mod_mul(curveParam -> a, x + 1, curveParam -> p, tmp1);
    //求解 tmp1 = x^3 + a*x
    bignum_mod_add(tmp1, tmp2, curveParam -> p, tmp1);
    //求解 tmp1 = x^3 + a*x +b
    bignum_mod_add(tmp1, curveParam -> b, curveParam -> p, tmp1);
    //下面求解tmp1的平方根
    
    //curveParam->p =3(mod 4)
    if((curveParam->p[ECC_LEN-1]&0x03)==3)
    {
        memset(tmp2,0,ECC_LEN);
        tmp2[ECC_LEN-1]=0x03;
        //tmp3=(p-3)/4
        bignum_sub(curveParam->p, tmp2, ECC_LEN,tmp3);
        bignum_shr_1bit(tmp3, ECC_LEN);
        bignum_shr_1bit(tmp3, ECC_LEN);
         memset(tmp2,0,ECC_LEN);
        tmp2[ECC_LEN-1]=0x01;
        //计算tmp3 = tmp3 + 1
        bignum_add(tmp3, tmp2, ECC_LEN, tmp3);
        //计算tmp2=tmp1^(tmp3)
        bignum_mod_exp(tmp1, tmp3, curveParam->p, tmp2);
        //计算 tmp3 = tmp2^2
        bignum_mod_mul(tmp2, tmp2, curveParam->p, tmp3);
        //check whether tmp1 is equal to tmp3.if it is,tmp3 is the result we need;otherwise,there is no square root.
        if(bignum_cmp(tmp1, ECC_LEN,tmp3,ECC_LEN)==0)
        {
            if(x[0]==0x02)
            {
                if(tmp2[ECC_LEN-1]&0x01)
                {
                    bignum_sub(curveParam->p, tmp2, ECC_LEN, point_buf + ECC_LEN + 1);
                }
                else
                {
                    memcpy(point_buf + ECC_LEN + 1,tmp2,ECC_LEN);
                    
                }
            }
           else if(x[0]==0x03)
            {
                if(tmp2[ECC_LEN-1]&0x01)
                {
                  memcpy(point_buf + ECC_LEN + 1,tmp2,ECC_LEN);
                }
                else
                {
                    bignum_sub(curveParam->p, tmp2, ECC_LEN, point_buf + ECC_LEN + 1);
                }
            }
            else
            {
                return 0;
            }
        }
        else
        {
            return 0;
        }
    }
    //curveParam->p = 5(mod 8)
    else if((curveParam->p[ECC_LEN-1]&7)==5)
    {
        memset(tmp2,0,ECC_LEN);
        //tmp4 = (p-5)/8
        tmp2[ECC_LEN-1]=0x05;
        bignum_sub(curveParam->p, tmp2, ECC_LEN,tmp4);
        bignum_shr_1bit(tmp4, ECC_LEN);
        bignum_shr_1bit(tmp4, ECC_LEN);
        bignum_shr_1bit(tmp4, ECC_LEN);
        //tmp3=2*tmp4
        bignum_add(tmp4, tmp4, ECC_LEN, tmp3);
        //tmp3 = tmp3 + 1
        memset(tmp2,0,ECC_LEN);
        tmp2[ECC_LEN]=0x01;
        bignum_add(tmp3, tmp2, ECC_LEN, tmp3);
        //tmp2=tmp1^tmp3
        bignum_mod_exp(tmp1, tmp3, curveParam->p, tmp2);
        bignum_mod(tmp2, curveParam->p, tmp3);
        
        memset(tmp2, 0, ECC_LEN);
        tmp2[ECC_LEN-1]=0x01;
        if( bignum_cmp(tmp3, ECC_LEN,tmp2,ECC_LEN)==0)
        {
            memset(tmp2,0,ECC_LEN);
            tmp2[ECC_LEN-1]=0x01;
            bignum_add(tmp4, tmp2, ECC_LEN, tmp4);
            bignum_mod_exp(tmp1, tmp4, curveParam->p, tmp2);
            if(x[0]==0x02)
            {
                if(tmp2[ECC_LEN-1]&0x01)
                {
                    bignum_sub(curveParam->p, tmp2, ECC_LEN, point_buf + ECC_LEN + 1);
                }
                else
                {
                    memcpy(point_buf + ECC_LEN + 1,tmp2,ECC_LEN);
                }
            }
            else if(x[0]==0x03)
            {
                if(tmp2[ECC_LEN-1]&0x01)
                {
                    memcpy(point_buf + ECC_LEN + 1,tmp2,ECC_LEN);
                }
                else
                {
                    bignum_sub(curveParam->p, tmp2, ECC_LEN, point_buf + ECC_LEN + 1);
                }
            }
            else
            {
                return 0;
            }
        }
        else
        {
            bignum_sub(curveParam->p, tmp2, ECC_LEN, tmp2);
            if(bignum_cmp(tmp3, ECC_LEN,tmp2,ECC_LEN)==0)
            {
                memset(tmp2,0,ECC_LEN);
                tmp2[ECC_LEN-1]=0x04;
                bignum_mod_mul(tmp1,tmp2,curveParam ->p, tmp3);
                bignum_mod_exp(tmp3, tmp4, curveParam->p, tmp2);
                memset(tmp3,0,ECC_LEN);
                tmp3[ECC_LEN] = 0x02;
                bignum_mod_mul(tmp1,tmp3,curveParam ->p, tmp4);
                bignum_mod_mul(tmp4,tmp2,curveParam ->p, tmp3);
                if(x[0]==0x02)
                {
                    if(tmp3[ECC_LEN-1]&0x01)
                    {
                       bignum_sub(curveParam->p, tmp3, ECC_LEN, point_buf + ECC_LEN + 1);
                    }
                    else
                    {
                        memcpy(point_buf + ECC_LEN + 1,tmp3,ECC_LEN);
                    }
                }
                else if(x[0]==0x03)
                {
                    if(tmp3[ECC_LEN-1]&0x01)
                    {
                        memcpy(point_buf + ECC_LEN + 1,tmp3,ECC_LEN);
                    }
                    else
                    {
                        bignum_sub(curveParam->p, tmp3, ECC_LEN, point_buf + ECC_LEN + 1);
                    }
                }
            }
            else
            {
                return 0;
            }
        }
    }
    //暂时不支持这种模式
    else if((curveParam->p[ECC_LEN-1]&7)==1)
    {
        ;
    }
    else
    {
        return 0;
    }
    point_buf[0]=0x04;
    memcpy(point_buf + 1,x+1,ECC_LEN);
    free(tmp1);
    free(tmp2);
    free(tmp3);
    free(tmp4);
    return 1;
}

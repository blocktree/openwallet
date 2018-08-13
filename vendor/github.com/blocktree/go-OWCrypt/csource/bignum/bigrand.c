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

#include "bigrand.h"

void bigrand_get_rand_range(uint8_t *bigrand, uint8_t *range, uint16_t len)
{
    uint32_t tmp = 0;
    uint16_t i = 0;
    
    srand((uint32_t)time(NULL));
    
    for(i = 0; i < len / 4; i ++)
    {
        tmp = rand();
        memcpy(bigrand + i * 4, (uint8_t *)&tmp, 4);
    }
    
    if((len - i * 4) % 4)
    {
        tmp = rand();
        memcpy(bigrand + i * 4, (uint8_t *)&tmp, (len - i * 4) % 4);
    }
    
    while(*bigrand >= *range)
        *bigrand -= *range;
}

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
#ifndef bigrand_h
#define bigrand_h

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include "type.h"

void bigrand_get_rand_range(uint8_t *rand, uint8_t *range, uint16_t len);

#endif /* bigrand_h */

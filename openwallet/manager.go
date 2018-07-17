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

package openwallet

import (
	"reflect"
	"github.com/ethereum/go-ethereum/event"
	"sync"
	"github.com/ethereum/go-ethereum/accounts"
)

// Manager openWallet作为客户端使用时，提供账户管理功能
type Manager struct {
	backends map[reflect.Type][]accounts.Backend // Index of backends currently registered
	updaters []event.Subscription       // Wallet update subscriptions for all backends
	//updates  chan WalletEvent           // Subscription sink for backend wallet changes
	accounts  []*AssetsAccount                   // Cache of all wallets from all registered backends

	feed event.Feed // Wallet feed notifying of arrivals/departures

	quit chan chan error
	lock sync.RWMutex
}

// NewManager creates a generic account manager to sign transaction via various
// supported backends.
//func NewManager(backends ...Backend) *Manager {
//	// Retrieve the initial list of wallets from the backends and sort by URL
//	var wallets []Wallet
//	for _, backend := range backends {
//		wallets = merge(wallets, backend.Wallets()...)
//	}
//	// Subscribe to wallet notifications from all backends
//	updates := make(chan WalletEvent, 4*len(backends))
//
//	subs := make([]event.Subscription, len(backends))
//	for i, backend := range backends {
//		subs[i] = backend.Subscribe(updates)
//	}
//	// Assemble the account manager and return
//	am := &Manager{
//		backends: make(map[reflect.Type][]Backend),
//		updaters: subs,
//		updates:  updates,
//		wallets:  wallets,
//		quit:     make(chan chan error),
//	}
//	for _, backend := range backends {
//		kind := reflect.TypeOf(backend)
//		am.backends[kind] = append(am.backends[kind], backend)
//	}
//	go am.update()
//
//	return am
//}
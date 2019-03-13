/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package walletnode

import (
	"fmt"
	s "strings"
)

// API for Walletnode Management
type WalletnodeManagerInterface interface {
	GetWalletnodeStatus(symbol string) error
	StartWalletnode(symbol string) error
	StopWalletnode(symbol string) error
	RestartWalletnode(symbol string) error

	// 获取全节点配置
	GetWalletnodeConfig(symbol string) *WalletnodeConfig
	// 更新全节点配置
	// UpdateWalletnodeConfig(nc *WalletnodeConfig, symbol string) error
	// 从钱包节点备份到本地
	CopyFromContainer(symbol, src, dst string) error
	// 从本地恢复钱包到节点
	CopyToContainer(symbol, src, dst string)
}

// func NewWalletnodeManager() WalletnodeManagerInterface {
// 	var wm WalletnodeManagerInterface
// 	wm = &WalletnodeManager{}
// 	return wm
// }

// For WMD Interface:NodeManager ()
type NodeManager struct {
	// GetNodeStatus   func(symbol string) error
	// CreateNodeFlow  func(symbol string) error
	// StartNodeFlow   func(symbol string) error
	// StopNodeFlow    func(symbol string) error
	// RestartNodeFlow func(symbol string) error
	// RemoveNodeFlow  func(string) error
}

func (nm *NodeManager) GetNodeStatus(symbol string) error {

	wn := WalletnodeManager{}

	if status, err := wn.GetWalletnodeStatus(symbol); err != nil {
		return err
	} else {
		fmt.Printf("%s walletnode status: %s\n", s.ToUpper(symbol), status)
	}

	return nil
}

// Create a new container for wallet fullnode
//
// Workflow:
//		// 步骤一: 判定本地 .ini 文件是否存在
//		if  .ini 不存在，创建一个默认的 {
//			1> 询问用户配置参数
//			2> 创建初始 .ini 文件
//		} else {									// .ini 存在
//			1> 无操作，进入下一步
//		}
//
//		// 步骤二：判断是否需要创建节点容器
//		if 容器不存在 or 不正常 {
//			1> 删除后，或直接创建一个新的(需：)
//			2> 导出 container 数据(IP, status)
//		} else {									// 正常
//			1> 导出 container 数据(IP, status)
//		}
//
//		// 步骤三
//		1> 根据导出的 container 数据，更新配置文件中关于 container 的项（重复更新，方便用户改错后自动恢复）
func (w *NodeManager) CreateNodeFlow(symbol string) error {

	// 一:
	if err := CheckAndCreateConfig(symbol); err != nil {
		return err
	}

	// 二:
	wn := WalletnodeManager{}
	if err := wn.CheckAdnCreateContainer(symbol); err != nil {
		return err
	}

	// 三:
	// Update config
	if err := updateConfig(symbol); err != nil {
		return err
	}

	return nil
}

func (nm *NodeManager) StartNodeFlow(symbol string) error {

	wn := WalletnodeManager{}

	if err := wn.StartWalletnode(symbol); err != nil {
		return err
	} else {
		fmt.Printf("%s walletnode start in success!\n", symbol)
	}

	return nil
}

func (nm *NodeManager) StopNodeFlow(symbol string) error {

	wn := WalletnodeManager{}

	if err := wn.StopWalletnode(symbol); err != nil {
		return err
	} else {
		fmt.Printf("%s walletnode stop in success!\n", symbol)
	}
	return nil
}

func (nm *NodeManager) RestartNodeFlow(symbol string) error {

	wn := WalletnodeManager{}

	if err := wn.RestartWalletnode(symbol); err != nil {
		return err
	} else {
		fmt.Printf("%s walletnode restart in success!\n", symbol)
	}

	return nil
}

func (nm *NodeManager) RemoveNodeFlow(symbol string) error {
	wn := WalletnodeManager{}

	if err := wn.RemoveWalletnode(symbol); err != nil {
		return err
	} else {
		fmt.Printf("%s walletnode remove in success!\n", symbol)
	}

	return nil
}

func (nm *NodeManager) LogsNodeFlow(symbol string) error {
	wn := WalletnodeManager{}

	if err := wn.LogsWalletnode(symbol); err != nil {
		return err
	} else {
		fmt.Printf("%s walletnode logs in success!\n", symbol)
	}

	return nil
}

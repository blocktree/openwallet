package walletnode

import (
	"fmt"
	"log"
	"path/filepath"
	s "strings"

	bconfig "github.com/astaxie/beego/config"
	"github.com/pkg/errors"
)

// Load settings for global from local conf/<Symbol>.ini
//
//		- will change Global VAR: WNConfig
func loadConfig(symbol string) error {

	var c bconfig.Configer

	configFilePath, _ := filepath.Abs("conf")
	configFileName := s.ToUpper(symbol) + ".ini"
	absFile := filepath.Join(configFilePath, configFileName)

	c, err := bconfig.NewConfig("ini", absFile)
	if err != nil {
		log.Println(err)
		return errors.New(fmt.Sprintf("Load Config Failed-> %s", err))
	}

	WNConfig.RPCUser = c.String("rpcUser")
	WNConfig.RPCPassword = c.String("rpcPassword")
	WNConfig.TestNet = c.String("isTestNet")

	WNConfig.walletnodePrefix = c.String("walletnode::Prefix")
	WNConfig.walletnodeServerType = c.String("walletnode::ServerType")
	WNConfig.walletnodeServerAddr = c.String("walletnode::ServerAddr")
	WNConfig.walletnodeServerPort = c.String("walletnode::ServerPort")
	WNConfig.walletnodeStartNodeCMD = c.String("walletnode::StartNodeCMD")
	WNConfig.walletnodeStopNodeCMD = c.String("walletnode::StopNodeCMD")
	WNConfig.walletnodeMainNetDataPath = c.String("walletnode::mainNetDataPath")
	WNConfig.walletnodeTestNetDataPath = c.String("walletnode::testNetDataPath")
	// WNConfig.walletnodeServerSocket = c.String("walletnode::WalletnodeServerSocket")

	return nil
}

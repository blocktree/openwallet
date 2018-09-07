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
	WNConfig = &WalletnodeConfig{}

	configFilePath, _ := filepath.Abs("conf")
	configFileName := s.ToUpper(symbol) + ".ini"
	absFile := filepath.Join(configFilePath, configFileName)

	c, err := bconfig.NewConfig("ini", absFile)
	if err != nil {
		log.Println(err)
		return errors.New(fmt.Sprintf("Load Config Failed-> %s", err))
	}

	WNConfig.mainNetDataPath = c.String("mainNetDataPath")
	WNConfig.testNetDataPath = c.String("testNetDataPath")
	WNConfig.RPCUser = c.String("rpcUser")
	WNConfig.RPCPassword = c.String("rpcPassword")
	WNConfig.TestNet = c.String("isTestNet")

	WNConfig.walletnodePrefix = c.String("walletnode::WalletnodePrefix")
	WNConfig.walletnodeServerType = c.String("walletnode::WalletnodeServerType")
	WNConfig.walletnodeServerAddr = c.String("walletnode::WalletnodeServerAddr")
	WNConfig.walletnodeServerPort = c.String("walletnode::WalletnodeServerPort")
	WNConfig.walletnodeServerSocket = c.String("walletnode::WalletnodeServerSocket")

	return nil
}

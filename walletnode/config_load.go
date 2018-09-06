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
	WNConfig.rpcUser = c.String("rpcUser")
	WNConfig.rpcPassword = c.String("rpcPassword")
	WNConfig.isTestNet = c.String("isTestNet")

	WNConfig.walletnodeServerType = c.String("walletnode::walletnodeServerType")
	WNConfig.walletnodeServerAddr = c.String("walletnode::walletnodeServerAddr")
	WNConfig.walletnodeServerSocket = c.String("walletnode::walletnodeServerSocket")

	return nil
}

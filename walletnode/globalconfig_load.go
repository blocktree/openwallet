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
//		- will change VAR
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

	mainNetDataPath = c.String("mainNetDataPath")
	testNetDataPath = c.String("testNetDataPath")
	rpcUser = c.String("rpcUser")
	rpcPassword = c.String("rpcPassword")
	isTestNet = c.String("isTestNet")
	serverAddr = c.String("serverAddr")
	serverPort = c.String("serverPort")

	return nil
}

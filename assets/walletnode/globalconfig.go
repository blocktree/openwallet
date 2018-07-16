package walletnode

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/pkg/errors"
	"path/filepath"
	s "strings"
)

var (
	isTestNet = true // default in TestNet

	// Connection of Docker master server
	serverAddr = "127.0.0.1"
	serverPort = 2735
)

// Load settings for global from local conf/<Symbol>.ini
func loadConfig(symbol string) error {
	var (
		c   config.Configer
		err error
	)
	configFilePath, _ := filepath.Abs("conf")
	configFileName := s.ToUpper(symbol) + ".ini"

	absFile := filepath.Join(configFilePath, configFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New(fmt.Sprintf("Load Config Failed: %s", err))
	}

	if v, err := c.Bool("isTestNet"); err != nil {
		return errors.New(fmt.Sprintf("Load Config Failed: %s", err))
	} else {
		isTestNet = v
	}

	return nil
}

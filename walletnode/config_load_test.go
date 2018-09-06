package walletnode

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	if err := loadConfig("bopo"); err != nil {
		t.Errorf("TestLoadConfig: %+v\n", err)
	}
}

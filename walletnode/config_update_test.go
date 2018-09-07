package walletnode

import (
	"testing"
)

func TestUpdateConfig(t *testing.T) {
	if err := loadConfig("bopo"); err != nil {
		t.Errorf("\tTestUpdateConfig: %+v\n", err)
	}

	if err := updateConfig("bopo"); err != nil {
		t.Errorf("\tTestUpdateConfig: %+v\n", err)
	}
}

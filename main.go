package OpenWallet

import (
	"github.com/ethereum/go-ethereum/common"
	"fmt"
	"github.com/asdine/storm"
)

func main() {

	address := common.HexToAddress("0xb9b94b8a1453becda8bb18677439bd23c554445a")
	fmt.Printf("address: %s",address.Hex())

	db, _ := storm.Open("my.db")
	defer db.Close()
}

package file

import (
	"testing"
	"log"
)

func TestCopyFile(t *testing.T) {
	Delete("/myspace/dest/wallet.txt")
	err := Copy("/myspace/source/text-1232323.txt", "/myspace/dest/wallet.txt")
	if err != nil {
		log.Fatal(err)
	}
}
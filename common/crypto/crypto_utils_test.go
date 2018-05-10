package crypto

import (
	"fmt"
	"testing"
)

func TestHmacSHA1(t *testing.T) {
	var (
		appkey     = "2d68067484a20f1a346b3cf28a898ed7f5736f5bacf0fe60449da95efdb97ad4"
		appsecret  = "0dd1e322907ad7f55deaa35fec2aac97cae7931454d734364bc63f3e9b9f993a"
		timestamp  = "1506565393"
		period     = "3600"
		ciphertext string
	)

	ciphertext = HmacSHA1(appsecret, appkey+timestamp+period)

	fmt.Println("ciphertext = ", ciphertext)
}

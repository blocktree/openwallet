package owtp

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var (
	httpHost *OWTPNode
	httpClient *OWTPNode
	httpURL = "127.0.0.1:8422"
	httpHostPrv = "FSomdQBZYzgu9YYuuSr3qXd8sP1sgQyk4rhLFo6gyi32"
	httpHostNodeID = "54dZTdotBmE9geGJmJcj7Qzm6fzNrEUJ2NcDwZYp2QEp"
)

func TestHTTPHostRun(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)
	cert, _ := NewCertificate(httpHostPrv, "aes")
	httpHost = NewOWTPNode(cert, 0, 0)
	fmt.Printf("nodeID = %s \n", httpHost.NodeID())
	config := make(map[string]string)
	config["address"] = httpURL
	config["connectType"] = HTTP
	config["enableSignature"] = "1"
	httpHost.HandleFunc("getInfo", getInfo)
	httpHost.Listen(config)

	<- endRunning
}

func TestHTTPClientCall(t *testing.T) {

	config := make(map[string]string)
	config["address"] = httpURL
	config["connectType"] = HTTP
	config["enableSignature"] = "1"
	httpClient := RandomOWTPNode()
	err := httpClient.Connect(httpHostNodeID, config)
	if err != nil {
		t.Errorf("Connect unexcepted error: %v", err)
		return
	}
	//err = httpClient.KeyAgreement(httpHostNodeID, "aes")
	//if err != nil {
	//	t.Errorf("KeyAgreement unexcepted error: %v", err)
	//	return
	//}

	params := map[string]interface{}{
		"name": "chance",
		"age": 18,
	}

	err = httpClient.Call(httpHostNodeID, "getInfo", params, false, func(resp Response) {

		result := resp.JsonData()
		symbols := result.Get("symbols")

		fmt.Printf("symbols: %v\n", symbols)
	})

	if err != nil {
		t.Errorf("unexcepted error: %v", err)
		return
	}
}

func TestHTTPKeyAgreement(t *testing.T) {

	var (
		//endRunning = make(chan bool, 1)
		url = "127.0.0.1:8422"
	)
	host := RandomOWTPNode()
	config := make(map[string]string)
	config["address"] = url
	config["connectType"] = HTTP
	host.HandleFunc("getInfo", getInfo)
	host.Listen(config)

	time.Sleep(2 * time.Second)

	client := RandomOWTPNode("aes")
	client.Connect(host.NodeID(), config)


	//cert, _ := NewCertificate(RandomPrivateKey(), "aes")
	//
	//pubkey, _ := cert.KeyPair()

	//pubkey := "2ESGLPkKwK1htLBAY259ARugtwBPzDV3H51QEYKuZqVp"
	//
	//fmt.Printf("pubkey: %v \n", pubkey)
	//
	//_, tmpPubkeyInitiator := owcrypt.KeyAgreement_initiator_step1(owcrypt.ECC_CURVE_SM2_STANDARD)
	//
	//param := map[string]interface{}{
	//	"pubkey":      pubkey,
	//	"tmpPubkey":   base58.Encode(tmpPubkeyInitiator),
	//	"consultType": "aes",
	//}
	//
	//err := client.Call(host.NodeID(), KeyAgreementMethod, param, true, func(resp Response) {
	//
	//	result := resp.JsonData()
	//
	//	//响应方协商结果
	//	pubkeyOther := result.Get("pubkeyOther").String()
	//	tmpPubkeyOther := result.Get("tmpPubkeyOther").String()
	//	sb := result.Get("sb").String()
	//
	//	fmt.Printf("pubkeyOther: %s\n", pubkeyOther)
	//	fmt.Printf("tmpPubkeyOther: %s\n", tmpPubkeyOther)
	//	fmt.Printf("sb: %s\n", sb)
	//})
	//
	////result, err := client.Call(KeyAgreementMethod, param)
	//if err != nil {
	//	t.Errorf("unexcepted error: %v", err)
	//	return
	//}

	//<- endRunning

	time.Sleep(5 * time.Second)
}


func TestConcurrentHTTPConnect(t *testing.T) {

	var wait sync.WaitGroup


	config := make(map[string]string)
	config["address"] = httpURL
	config["connectType"] = HTTP

	for i := 0; i < 100; i++ {
		wait.Add(100)
		go func() {

			httpClient := RandomOWTPNode()
			err := httpClient.Connect(httpHostNodeID, config)
			if err != nil {
				t.Errorf("Connect unexcepted error: %v", err)
				return
			}
			//err = httpClient.KeyAgreement(httpHostNodeID, "aes")
			//if err != nil {
			//	t.Errorf("KeyAgreement unexcepted error: %v", err)
			//	return
			//}

			params := map[string]interface{}{
				"name": "chance",
				"age": 18,
			}

			for i := 0; i<100;i++ {
				err = httpClient.Call(httpHostNodeID, "getInfo", params, false, func(resp Response) {

					result := resp.JsonData()
					symbols := result.Get("symbols")

					fmt.Printf("symbols: %v\n", symbols)
					wait.Done()
				})

				if err != nil {
					t.Errorf("unexcepted error: %v", err)
					return
				}

				time.Sleep(500)
			}
		}()
	}

	wait.Wait()

}
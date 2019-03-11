package owtp

import (
	"encoding/hex"
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/session"
	"github.com/blocktree/go-owcrypt"
	"github.com/imroc/req"
	"github.com/mr-tron/base58/base58"
	"sync"
	"testing"
	"time"
)

var (
	httpHost       *OWTPNode
	httpClient     *OWTPNode
	httpURL        = "0.0.0.0:8422"
	httpHostPrv    = "FSomdQBZYzgu9YYuuSr3qXd8sP1sgQyk4rhLFo6gyi32"
	httpHostNodeID = "54dZTdotBmE9geGJmJcj7Qzm6fzNrEUJ2NcDwZYp2QEp"
	globalSessions *SessionManager
)

func testSetupGlobalSession() {
	globalSessions, _ = NewSessionManager("memory", &session.ManagerConfig{
		Gclifetime: 3600,
	})
	go globalSessions.GC()
}

func init() {
	testSetupGlobalSession()
}

func testMakeHTTPCall(httpClient *OWTPNode) {

	config := ConnectConfig{}
	config.Address = httpURL
	config.ConnectType = HTTP

	params := map[string]interface{}{
		"name": "chance",
		"age":  18,
	}
	//err = httpClient.Connect(httpHostNodeID, config)
	err := httpClient.ConnectAndCall(httpHostNodeID, config,"getInfo", params, false, func(resp Response) {
		if resp.Status == StatusSuccess {
			result := resp.JsonData()
			symbols := result.Get("symbols")
			fmt.Printf("symbols: %v\n", symbols)
		} else {
			log.Error(resp)
		}

	})
	//err := httpClient.Call(httpHostNodeID, "getInfo", params, true, func(resp Response) {
	//	if resp.Status == StatusSuccess {
	//		result := resp.JsonData()
	//		symbols := result.Get("symbols")
	//		fmt.Printf("symbols: %v\n", symbols)
	//	} else {
	//		log.Error(resp)
	//	}
	//
	//})

	//err := httpClient.ConnectAndCall(httpHostNodeID, config, "getInfo", params, true, func(resp Response) {
	//
	//	result := resp.JsonData()
	//	symbols := result.Get("symbols")
	//	fmt.Printf("symbols: %v\n", symbols)
	//})

	if err != nil {
		log.Error("unexcepted error: %v", err)
		return
	}
}

func TestHTTPHostRun(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)

	cert, _ := NewCertificate(httpHostPrv, "aes")
	httpHost = NewOWTPNode(cert, 0, 0)
	httpHost.SetPeerstore(globalSessions)
	fmt.Printf("nodeID = %s \n", httpHost.NodeID())
	config := ConnectConfig{}
	config.Address = httpURL
	config.ConnectType = HTTP
	//config["enableSignature"] = "1"
	httpHost.HandleFunc("getInfo", getInfo)
	httpHost.HandlePrepareFunc(func(ctx *Context) {
		log.Notice("remoteAddress:", ctx.RemoteAddress)
		log.Notice("prepare")
		//ctx.ResponseStopRun(nil, StatusSuccess, "success")
	})
	httpHost.HandleFinishFunc(func(ctx *Context) {
		username := ctx.GetSession("username")
		log.Notice("username:", username)
		log.Notice("finish")
	})
	httpHost.Listen(config)

	<-endRunning
}

func TestHTTPClientCall(t *testing.T) {

	//config := ConnectConfig{}
	//config.Address = httpURL
	//config.ConnectType = HTTP
	//config["enableSignature"] = "1"
	cert, _ := NewCertificate("E3cQTqKZfVVL6cQvyrSgbjmkVnnbkBuoqt7ed9wQLjgz", "aes")
	//httpClient := RandomOWTPNode()
	httpClient := NewNode(NodeConfig{
		Cert: cert,
	})
	httpClient.SetPeerstore(globalSessions)
	prv, pub := httpClient.Certificate().KeyPair()
	log.Info("pub:", pub)
	log.Info("prv:", prv)
	//err := httpClient.Connect(httpHostNodeID, config)
	//if err != nil {
	//	t.Errorf("Connect unexcepted error: %v", err)
	//	return
	//}
	err := httpClient.KeyAgreement(httpHostNodeID, "aes")
	if err != nil {
		t.Errorf("KeyAgreement unexcepted error: %v", err)
		return
	}

	testMakeHTTPCall(httpClient)
}


func TestHTTPClientContinueCall(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
	)

	config := ConnectConfig{}
	config.Address = httpURL
	config.ConnectType = HTTP
	//config["enableSignature"] = "1"
	cert, _ := NewCertificate("E3cQTqKZfVVL6cQvyrSgbjmkVnnbkBuoqt7ed9wQLjgz", "aes")
	//httpClient := RandomOWTPNode()
	httpClient := NewNode(NodeConfig{
		Cert: cert,
	})
	httpClient.SetPeerstore(globalSessions)
	prv, pub := httpClient.Certificate().KeyPair()
	log.Info("pub:", pub)
	log.Info("prv:", prv)
	//err := httpClient.Connect(httpHostNodeID, config)
	//if err != nil {
	//	t.Errorf("Connect unexcepted error: %v", err)
	//	return
	//}
	//err = httpClient.KeyAgreement(httpHostNodeID, "aes")
	//if err != nil {
	//	t.Errorf("KeyAgreement unexcepted error: %v", err)
	//	return
	//}

	for i := 0;i<500000;i++ {
		testMakeHTTPCall(httpClient)

	}

	<-endRunning
}

func TestHTTPKeyAgreement(t *testing.T) {

	var (
		//endRunning = make(chan bool, 1)
		url = "127.0.0.1:8422"
	)
	host := RandomOWTPNode()
	config := ConnectConfig{}
	config.Address = url
	config.ConnectType = HTTP
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

	config := ConnectConfig{}
	config.Address = httpURL
	config.ConnectType = HTTP
	//config["enableSignature"] = "1"
	for i := 0; i < 100; i++ {
		wait.Add(100)
		go func() {

			httpClient := RandomOWTPNode()
			err := httpClient.Connect(httpHostNodeID, config)
			if err != nil {
				t.Errorf("Connect unexcepted error: %v", err)
				return
			}
			err = httpClient.KeyAgreement(httpHostNodeID, "aes")
			if err != nil {
				t.Errorf("KeyAgreement unexcepted error: %v", err)
				return
			}

			params := map[string]interface{}{
				"name": "chance",
				"age":  18,
			}

			for i := 0; i < 100; i++ {
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

				//time.Sleep(500)
			}
		}()
	}

	wait.Wait()

}

func TestHTTPNormalCall(t *testing.T) {

	b, _ := hex.DecodeString("1234abef")
	a := base58.Encode(b)
	authHeader := req.Header{
		"Accept": "application/json",
		"a":      a,
	}

	hash := owcrypt.Hash(b, 0, owcrypt.HASH_ALG_SHA256)
	base := base58.Encode(hash)
	log.Info("nodeID:", base)
	res, err := req.Post("http://"+httpURL, authHeader)
	if err != nil {
		t.Errorf("Connect unexcepted error: %v", err)
		return
	}

	log.Infof("res: +%v", res)
}

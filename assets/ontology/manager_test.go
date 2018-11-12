package ontology

var (
	tw *WalletManager
)

func init() {

	tw = NewWalletManager()

	tw.Config.ServerAPI = "http://127.0.0.1:20336"
	tw.Config.RpcUser = ""
	tw.Config.RpcPassword = ""
	token := BasicAuth(tw.Config.RpcUser, tw.Config.RpcPassword)
	tw.WalletClient = NewClient(tw.Config.ServerAPI, token, false)

	//explorerURL := "http://192.168.32.107:20003/insight-api/"
	//tw.ExplorerClient = NewExplorer(explorerURL, true)

	tw.Config.RPCServerType = RPCServerCore
}
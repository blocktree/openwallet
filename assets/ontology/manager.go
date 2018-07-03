package ontology

import "github.com/ontio/ontology-go-sdk"

func Helloworld() {

	sdk := ontology_go_sdk.NewOntologySdk()
	sdk.Rpc.SetAddress("http://192.168.2.224:10056")
	sdk.Rpc.GetVersion()
}

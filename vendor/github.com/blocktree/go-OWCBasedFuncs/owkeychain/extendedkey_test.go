package owkeychain

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/blocktree/go-OWCrypt"
)

//normal private key extend based on secp256k1 private key
func Test_GenPrivateChild_fromPrivate_secp256k1_normal(t *testing.T) {

	passFlag := true
	//test cases based on secp256k1
	//set root private key
	rootPri := [32]byte{0x9e, 0xa1, 0x9e, 0x6e, 0xc2, 0x59, 0xf7, 0x85, 0x4e, 0xe4, 0x1b, 0x53, 0x07, 0xcf, 0xc4, 0xb8, 0xf4, 0x47, 0x75, 0x34, 0x20, 0x5e, 0xc9, 0x83, 0xc4, 0xd3, 0xa9, 0xb5, 0x6c, 0x0b, 0x27, 0x0c}
	rootChainCode := [32]byte{0xab, 0xc9, 0xcc, 0x46, 0xa8, 0x16, 0x6d, 0x81, 0x55, 0xac, 0x1e, 0xd1, 0x2b, 0xe4, 0x11, 0xcd, 0x21, 0x3a, 0x3e, 0x28, 0xe4, 0xef, 0x46, 0x46, 0xfe, 0x03, 0xd7, 0x00, 0x2f, 0xef, 0x15, 0x2c}
	rootParentFP := [4]byte{0, 0, 0, 0}

	rooPriKey := NewExtendedKey(rootPri[:], rootChainCode[:], rootParentFP[:], 0, 0, true, owcrypt.ECC_CURVE_SECP256K1)

	//print root private key
	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	fmt.Println("父私钥 -----> 普通子私钥")
	fmt.Println("root private key data:")
	fmt.Println("key:", hex.EncodeToString(rooPriKey.key))
	fmt.Println("chaincode:", hex.EncodeToString(rooPriKey.chainCode))
	fmt.Println("parent FP:", hex.EncodeToString(rooPriKey.parentFP))
	fmt.Println("dpth:", rooPriKey.depth)
	fmt.Println("serializes", rooPriKey.serializes)
	fmt.Println("private flag:", rooPriKey.isPrivate)

	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")

	//normal extend , serializes = 0x0
	serialize := uint32(0)
	//expect data
	expectChildPri := "f938a2e7fef45315b9b0c31b4db08e23a84b362e71876e7fc1880b2ea94e38f1"
	expectChildChainCode := "a9e25b8ef131d1180292e8b7ef967347004ed436abf02ea14929325952f72809"
	expectChildParentFP := "fb080f46"
	expectChildDpth := uint8(1)
	expectChildSerialize := serialize
	expectChildPriFlag := true

	childPriKey, err := rooPriKey.GenPrivateChild(serialize)

	if err != nil {
		t.Error("父私钥向子私钥扩展出错")
	} else {
		//check the result
		if expectChildPri != hex.EncodeToString(childPriKey.key) {
			t.Error("扩展的子私钥数据错误")
			passFlag = false
		}

		if expectChildChainCode != hex.EncodeToString(childPriKey.chainCode) {
			t.Error("扩展的子私钥链码数据错误")
			passFlag = false
		}

		if expectChildParentFP != hex.EncodeToString(childPriKey.parentFP) {
			t.Error("扩展的子私钥父指纹数据错误")
			passFlag = false
		}

		if expectChildDpth != childPriKey.depth {
			t.Error("扩展的子私钥深度数据错误")
			passFlag = false
		}

		if expectChildSerialize != childPriKey.serializes {
			t.Error("扩展的子私钥索引号数据错误")
			passFlag = false
		}

		if expectChildPriFlag != childPriKey.isPrivate {
			t.Error("扩展的子私钥公私钥标记数据错误")
			passFlag = false
		}
		//print child private key
		if passFlag {
			fmt.Println("child private key data:")
			fmt.Println("key:", hex.EncodeToString(childPriKey.key))
			fmt.Println("chaincode:", hex.EncodeToString(childPriKey.chainCode))
			fmt.Println("parent FP:", hex.EncodeToString(childPriKey.parentFP))
			fmt.Println("dpth:", childPriKey.depth)
			fmt.Println("serializes", childPriKey.serializes)
			fmt.Println("private flag:", childPriKey.isPrivate)
		}

		fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	}

}

//normal public key extend based on secp256k1 private key
func Test_GenPublicChild_fromPrivate_secp256k1_normal(t *testing.T) {
	passFlag := true

	//test cases based on secp256k1
	//set root private key
	rootPri := [32]byte{0x9e, 0xa1, 0x9e, 0x6e, 0xc2, 0x59, 0xf7, 0x85, 0x4e, 0xe4, 0x1b, 0x53, 0x07, 0xcf, 0xc4, 0xb8, 0xf4, 0x47, 0x75, 0x34, 0x20, 0x5e, 0xc9, 0x83, 0xc4, 0xd3, 0xa9, 0xb5, 0x6c, 0x0b, 0x27, 0x0c}
	rootChainCode := [32]byte{0xab, 0xc9, 0xcc, 0x46, 0xa8, 0x16, 0x6d, 0x81, 0x55, 0xac, 0x1e, 0xd1, 0x2b, 0xe4, 0x11, 0xcd, 0x21, 0x3a, 0x3e, 0x28, 0xe4, 0xef, 0x46, 0x46, 0xfe, 0x03, 0xd7, 0x00, 0x2f, 0xef, 0x15, 0x2c}
	rootParentFP := [4]byte{0, 0, 0, 0}

	rooPriKey := NewExtendedKey(rootPri[:], rootChainCode[:], rootParentFP[:], 0, 0, true, owcrypt.ECC_CURVE_SECP256K1)

	//print root private key
	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	fmt.Println("父私钥 -----> 普通子公钥")
	fmt.Println("root private key data:")
	fmt.Println("key:", hex.EncodeToString(rooPriKey.key))
	fmt.Println("chaincode:", hex.EncodeToString(rooPriKey.chainCode))
	fmt.Println("parent FP:", hex.EncodeToString(rooPriKey.parentFP))
	fmt.Println("dpth:", rooPriKey.depth)
	fmt.Println("serializes", rooPriKey.serializes)
	fmt.Println("private flag:", rooPriKey.isPrivate)

	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")

	//normal extend , serializes = 0x0
	serialize := uint32(0)
	//expect data
	expectChildPub := "0347e1f04775f36482cf78ea6d028ac71ab423199e37e04cbb448f31f973a63bba"
	expectChildChainCode := "a9e25b8ef131d1180292e8b7ef967347004ed436abf02ea14929325952f72809"
	expectChildParentFP := "fb080f46"
	expectChildDpth := uint8(1)
	expectChildSerialize := serialize
	expectChildPriFlag := false

	childPubKey, err := rooPriKey.GenPublicChild(serialize)

	if err != nil {
		t.Error("父私钥向子私钥扩展出错")
	} else {
		//check the result
		if expectChildPub != hex.EncodeToString(childPubKey.key) {
			t.Error("扩展的子私钥数据错误")
			passFlag = false
		}

		if expectChildChainCode != hex.EncodeToString(childPubKey.chainCode) {
			t.Error("扩展的子私钥链码数据错误")
			passFlag = false
		}

		if expectChildParentFP != hex.EncodeToString(childPubKey.parentFP) {
			t.Error("扩展的子私钥父指纹数据错误")
			passFlag = false
		}

		if expectChildDpth != childPubKey.depth {
			t.Error("扩展的子私钥深度数据错误")
			passFlag = false
		}

		if expectChildSerialize != childPubKey.serializes {
			t.Error("扩展的子私钥索引号数据错误")
			passFlag = false
		}

		if expectChildPriFlag != childPubKey.isPrivate {
			t.Error("扩展的子私钥公私钥标记数据错误")
			passFlag = false
		}
		//print child private key
		if passFlag {
			fmt.Println("child public key data:")
			fmt.Println("key:", hex.EncodeToString(childPubKey.key))
			fmt.Println("chaincode:", hex.EncodeToString(childPubKey.chainCode))
			fmt.Println("parent FP:", hex.EncodeToString(childPubKey.parentFP))
			fmt.Println("dpth:", childPubKey.depth)
			fmt.Println("serializes", childPubKey.serializes)
			fmt.Println("private flag:", childPubKey.isPrivate)
		}

		fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	}
}

//HD private key extend based on secp256k1 private key
func Test_GenPrivateChild_fromPrivate_secp256k1_hd(t *testing.T) {
	passFlag := true
	//test cases based on secp256k1
	//set root private key
	rootPri := [32]byte{0x9e, 0xa1, 0x9e, 0x6e, 0xc2, 0x59, 0xf7, 0x85, 0x4e, 0xe4, 0x1b, 0x53, 0x07, 0xcf, 0xc4, 0xb8, 0xf4, 0x47, 0x75, 0x34, 0x20, 0x5e, 0xc9, 0x83, 0xc4, 0xd3, 0xa9, 0xb5, 0x6c, 0x0b, 0x27, 0x0c}
	rootChainCode := [32]byte{0xab, 0xc9, 0xcc, 0x46, 0xa8, 0x16, 0x6d, 0x81, 0x55, 0xac, 0x1e, 0xd1, 0x2b, 0xe4, 0x11, 0xcd, 0x21, 0x3a, 0x3e, 0x28, 0xe4, 0xef, 0x46, 0x46, 0xfe, 0x03, 0xd7, 0x00, 0x2f, 0xef, 0x15, 0x2c}
	rootParentFP := [4]byte{0, 0, 0, 0}

	rooPriKey := NewExtendedKey(rootPri[:], rootChainCode[:], rootParentFP[:], 0, 0, true, owcrypt.ECC_CURVE_SECP256K1)

	//print root private key
	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	fmt.Println("父私钥 -----> 强化子私钥")
	fmt.Println("root private key data:")
	fmt.Println("key:", hex.EncodeToString(rooPriKey.key))
	fmt.Println("chaincode:", hex.EncodeToString(rooPriKey.chainCode))
	fmt.Println("parent FP:", hex.EncodeToString(rooPriKey.parentFP))
	fmt.Println("dpth:", rooPriKey.depth)
	fmt.Println("serializes", rooPriKey.serializes)
	fmt.Println("private flag:", rooPriKey.isPrivate)

	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")

	//HD extend , serializes = 0x80000000
	serialize := uint32(0x80000000)
	//expect data
	expectChildPri := "54d5bf2f8f82107ef1cd6a55c10a852643bde4653bfd4faa8d5ca1c14d9e7120"
	expectChildChainCode := "8f1b8987c2d267d31980afce919f0276cb93e98972d6e27d3ba79550f729cce1"
	expectChildParentFP := "fb080f46"
	expectChildDpth := uint8(1)
	expectChildSerialize := serialize
	expectChildPriFlag := true

	childPriKey, err := rooPriKey.GenPrivateChild(serialize)

	if err != nil {
		t.Error("父私钥向子私钥扩展出错")
	} else {
		//check the result
		if expectChildPri != hex.EncodeToString(childPriKey.key) {
			t.Error("扩展的子私钥数据错误")
			passFlag = false
		}

		if expectChildChainCode != hex.EncodeToString(childPriKey.chainCode) {
			t.Error("扩展的子私钥链码数据错误")
			passFlag = false
		}

		if expectChildParentFP != hex.EncodeToString(childPriKey.parentFP) {
			t.Error("扩展的子私钥父指纹数据错误")
			passFlag = false
		}

		if expectChildDpth != childPriKey.depth {
			t.Error("扩展的子私钥深度数据错误")
			passFlag = false
		}

		if expectChildSerialize != childPriKey.serializes {
			t.Error("扩展的子私钥索引号数据错误")
			passFlag = false
		}

		if expectChildPriFlag != childPriKey.isPrivate {
			t.Error("扩展的子私钥公私钥标记数据错误")
			passFlag = false
		}
		//print child private key
		if passFlag {
			fmt.Println("child private key data:")
			fmt.Println("key:", hex.EncodeToString(childPriKey.key))
			fmt.Println("chaincode:", hex.EncodeToString(childPriKey.chainCode))
			fmt.Println("parent FP:", hex.EncodeToString(childPriKey.parentFP))
			fmt.Println("dpth:", childPriKey.depth)
			fmt.Println("serializes", childPriKey.serializes)
			fmt.Println("private flag:", childPriKey.isPrivate)
		}

		fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	}

}

//HD public key extend based on secp256k1 private key
func Test_GenPublicChild_fromPrivate_secp256k1_hd(t *testing.T) {

	passFlag := true
	//test cases based on secp256k1
	//set root private key
	rootPri := [32]byte{0x9e, 0xa1, 0x9e, 0x6e, 0xc2, 0x59, 0xf7, 0x85, 0x4e, 0xe4, 0x1b, 0x53, 0x07, 0xcf, 0xc4, 0xb8, 0xf4, 0x47, 0x75, 0x34, 0x20, 0x5e, 0xc9, 0x83, 0xc4, 0xd3, 0xa9, 0xb5, 0x6c, 0x0b, 0x27, 0x0c}
	rootChainCode := [32]byte{0xab, 0xc9, 0xcc, 0x46, 0xa8, 0x16, 0x6d, 0x81, 0x55, 0xac, 0x1e, 0xd1, 0x2b, 0xe4, 0x11, 0xcd, 0x21, 0x3a, 0x3e, 0x28, 0xe4, 0xef, 0x46, 0x46, 0xfe, 0x03, 0xd7, 0x00, 0x2f, 0xef, 0x15, 0x2c}
	rootParentFP := [4]byte{0, 0, 0, 0}

	rooPriKey := NewExtendedKey(rootPri[:], rootChainCode[:], rootParentFP[:], 0, 0, true, owcrypt.ECC_CURVE_SECP256K1)

	//print root private key
	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	fmt.Println("父私钥 -----> 强化子公钥")
	fmt.Println("root private key data:")
	fmt.Println("key:", hex.EncodeToString(rooPriKey.key))
	fmt.Println("chaincode:", hex.EncodeToString(rooPriKey.chainCode))
	fmt.Println("parent FP:", hex.EncodeToString(rooPriKey.parentFP))
	fmt.Println("dpth:", rooPriKey.depth)
	fmt.Println("serializes", rooPriKey.serializes)
	fmt.Println("private flag:", rooPriKey.isPrivate)

	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")

	//normal extend , serializes = 0x80000000
	serialize := uint32(0x80000000)
	//expect data
	expectChildPub := "03f6130e91673fea46204c6f05115c501421fd1ff038ab8c3371e8e81a6060a8e4"
	expectChildChainCode := "8f1b8987c2d267d31980afce919f0276cb93e98972d6e27d3ba79550f729cce1"
	expectChildParentFP := "fb080f46"
	expectChildDpth := uint8(1)
	expectChildSerialize := serialize
	expectChildPriFlag := false

	childPubKey, err := rooPriKey.GenPublicChild(serialize)

	if err != nil {
		t.Error("父私钥向子私钥扩展出错")
	} else {
		//check the result
		if expectChildPub != hex.EncodeToString(childPubKey.key) {
			t.Error("扩展的子私钥数据错误")
			passFlag = false
		}

		if expectChildChainCode != hex.EncodeToString(childPubKey.chainCode) {
			t.Error("扩展的子私钥链码数据错误")
			passFlag = false
		}

		if expectChildParentFP != hex.EncodeToString(childPubKey.parentFP) {
			t.Error("扩展的子私钥父指纹数据错误")
			passFlag = false
		}

		if expectChildDpth != childPubKey.depth {
			t.Error("扩展的子私钥深度数据错误")
			passFlag = false
		}

		if expectChildSerialize != childPubKey.serializes {
			t.Error("扩展的子私钥索引号数据错误")
			passFlag = false
		}

		if expectChildPriFlag != childPubKey.isPrivate {
			t.Error("扩展的子私钥公私钥标记数据错误")
			passFlag = false
		}
		//print child private key
		if passFlag {
			fmt.Println("child public key data:")
			fmt.Println("key:", hex.EncodeToString(childPubKey.key))
			fmt.Println("chaincode:", hex.EncodeToString(childPubKey.chainCode))
			fmt.Println("parent FP:", hex.EncodeToString(childPubKey.parentFP))
			fmt.Println("dpth:", childPubKey.depth)
			fmt.Println("serializes", childPubKey.serializes)
			fmt.Println("private flag:", childPubKey.isPrivate)
		}

		fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	}
}

//normal private key extend based on secp256k1 public key
func Test_GenPrivateChild_fromPublic_secp256k1_normal(t *testing.T) {

	//test cases based on secp256k1
	//set root public key
	rootPub := [33]byte{0x02, 0x83, 0x84, 0x5E, 0x2B, 0x12, 0x0F, 0xB5, 0x96, 0x4E, 0x9F, 0x48, 0x7F, 0x1C, 0x87, 0xE8, 0x7A, 0xE9, 0xF7, 0xD1, 0x6F, 0x0A, 0x5E, 0x77, 0x13, 0x14, 0x7F, 0x9E, 0x84, 0xF4, 0xAD, 0x10, 0x06}
	rootChainCode := [32]byte{0xab, 0xc9, 0xcc, 0x46, 0xa8, 0x16, 0x6d, 0x81, 0x55, 0xac, 0x1e, 0xd1, 0x2b, 0xe4, 0x11, 0xcd, 0x21, 0x3a, 0x3e, 0x28, 0xe4, 0xef, 0x46, 0x46, 0xfe, 0x03, 0xd7, 0x00, 0x2f, 0xef, 0x15, 0x2c}
	rootParentFP := [4]byte{0, 0, 0, 0}

	rooPubKey := NewExtendedKey(rootPub[:], rootChainCode[:], rootParentFP[:], 0, 0, false, owcrypt.ECC_CURVE_SECP256K1)
	serialize := uint32(0)
	childPriKey, err := rooPubKey.GenPrivateChild(serialize)
	fmt.Println("父公钥 -----> 普通子私钥")
	if childPriKey != nil {
		t.Error("未能检出父公钥派生子私钥的非法操作")
	}
	if err != ErrNotPrivExtKey {
		t.Error("抛出了错误的异常")
	}

}

//HD private key extend based on secp256k1 public key
func Test_GenPrivateChild_fromPublic_secp256k1_HD(t *testing.T) {

	//test cases based on secp256k1
	//set root public key
	rootPub := [33]byte{0x02, 0x83, 0x84, 0x5E, 0x2B, 0x12, 0x0F, 0xB5, 0x96, 0x4E, 0x9F, 0x48, 0x7F, 0x1C, 0x87, 0xE8, 0x7A, 0xE9, 0xF7, 0xD1, 0x6F, 0x0A, 0x5E, 0x77, 0x13, 0x14, 0x7F, 0x9E, 0x84, 0xF4, 0xAD, 0x10, 0x06}
	rootChainCode := [32]byte{0xab, 0xc9, 0xcc, 0x46, 0xa8, 0x16, 0x6d, 0x81, 0x55, 0xac, 0x1e, 0xd1, 0x2b, 0xe4, 0x11, 0xcd, 0x21, 0x3a, 0x3e, 0x28, 0xe4, 0xef, 0x46, 0x46, 0xfe, 0x03, 0xd7, 0x00, 0x2f, 0xef, 0x15, 0x2c}
	rootParentFP := [4]byte{0, 0, 0, 0}

	rooPubKey := NewExtendedKey(rootPub[:], rootChainCode[:], rootParentFP[:], 0, 0, false, owcrypt.ECC_CURVE_SECP256K1)
	serialize := uint32(0x80000000)
	childPriKey, err := rooPubKey.GenPrivateChild(serialize)
	fmt.Println("父公钥 -----> 强化子私钥")
	if childPriKey != nil {
		t.Error("未能检出父公钥派生子私钥的非法操作")
	}
	if err != ErrNotPrivExtKey {
		t.Error("抛出了错误的异常")
	}

}

//normal public key extend based on secp256k1 public key
func Test_GenPublicChild_fromPublic_secp256k1_normal(t *testing.T) {

	passFlag := true
	//test cases based on secp256k1
	//set root public key
	rootPub := [33]byte{0x02, 0x83, 0x84, 0x5E, 0x2B, 0x12, 0x0F, 0xB5, 0x96, 0x4E, 0x9F, 0x48, 0x7F, 0x1C, 0x87, 0xE8, 0x7A, 0xE9, 0xF7, 0xD1, 0x6F, 0x0A, 0x5E, 0x77, 0x13, 0x14, 0x7F, 0x9E, 0x84, 0xF4, 0xAD, 0x10, 0x06}
	rootChainCode := [32]byte{0xab, 0xc9, 0xcc, 0x46, 0xa8, 0x16, 0x6d, 0x81, 0x55, 0xac, 0x1e, 0xd1, 0x2b, 0xe4, 0x11, 0xcd, 0x21, 0x3a, 0x3e, 0x28, 0xe4, 0xef, 0x46, 0x46, 0xfe, 0x03, 0xd7, 0x00, 0x2f, 0xef, 0x15, 0x2c}
	rootParentFP := [4]byte{0, 0, 0, 0}

	rooPubKey := NewExtendedKey(rootPub[:], rootChainCode[:], rootParentFP[:], 0, 0, false, owcrypt.ECC_CURVE_SECP256K1)

	//print root private key
	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	fmt.Println("父公钥 -----> 普通子公钥")
	fmt.Println("root private key data:")
	fmt.Println("key:", hex.EncodeToString(rooPubKey.key))
	fmt.Println("chaincode:", hex.EncodeToString(rooPubKey.chainCode))
	fmt.Println("parent FP:", hex.EncodeToString(rooPubKey.parentFP))
	fmt.Println("dpth:", rooPubKey.depth)
	fmt.Println("serializes", rooPubKey.serializes)
	fmt.Println("private flag:", rooPubKey.isPrivate)

	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	serialize := uint32(0)
	//expect data
	expectChildPub := "0347e1f04775f36482cf78ea6d028ac71ab423199e37e04cbb448f31f973a63bba"
	expectChildChainCode := "a9e25b8ef131d1180292e8b7ef967347004ed436abf02ea14929325952f72809"
	expectChildParentFP := "fb080f46"
	expectChildDpth := uint8(1)
	expectChildSerialize := serialize
	expectChildPriFlag := false
	childPubKey, err := rooPubKey.GenPublicChild(serialize)

	if err != nil {
		t.Error("父公钥向普通子公钥扩展错误")
	} else {
		if expectChildPub != hex.EncodeToString(childPubKey.key) {
			t.Error("扩展的子公钥数据错误")
			passFlag = false
		}
		if expectChildChainCode != hex.EncodeToString(childPubKey.chainCode) {
			t.Error("扩展的子公钥链码数据错误")
			passFlag = false
		}
		if expectChildParentFP != hex.EncodeToString(childPubKey.parentFP) {
			t.Error("扩展的子公钥父指纹数据错误")
			passFlag = false
		}
		if expectChildDpth != childPubKey.depth {
			t.Error("扩展的子公钥深度数据错误")
			passFlag = false
		}
		if expectChildSerialize != childPubKey.serializes {
			t.Error("扩展的子公钥索引号数据错误")
			passFlag = false
		}
		if expectChildPriFlag != childPubKey.isPrivate {
			t.Error("扩展的子公钥私钥标记数据错误")
			passFlag = false
		}
	}
	if passFlag {
		fmt.Println("child public key data:")
		fmt.Println("key:", hex.EncodeToString(childPubKey.key))
		fmt.Println("chaincode:", hex.EncodeToString(childPubKey.chainCode))
		fmt.Println("parent FP:", hex.EncodeToString(childPubKey.parentFP))
		fmt.Println("dpth:", childPubKey.depth)
		fmt.Println("serializes", childPubKey.serializes)
		fmt.Println("private flag:", childPubKey.isPrivate)
		fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	}

}

//HD public key extend based on secp256k1 public key
func Test_GenPublicChild_fromPublic_secp256k1_HD(t *testing.T) {

	//test cases based on secp256k1
	//set root public key
	rootPub := [33]byte{0x02, 0x83, 0x84, 0x5E, 0x2B, 0x12, 0x0F, 0xB5, 0x96, 0x4E, 0x9F, 0x48, 0x7F, 0x1C, 0x87, 0xE8, 0x7A, 0xE9, 0xF7, 0xD1, 0x6F, 0x0A, 0x5E, 0x77, 0x13, 0x14, 0x7F, 0x9E, 0x84, 0xF4, 0xAD, 0x10, 0x06}
	rootChainCode := [32]byte{0xab, 0xc9, 0xcc, 0x46, 0xa8, 0x16, 0x6d, 0x81, 0x55, 0xac, 0x1e, 0xd1, 0x2b, 0xe4, 0x11, 0xcd, 0x21, 0x3a, 0x3e, 0x28, 0xe4, 0xef, 0x46, 0x46, 0xfe, 0x03, 0xd7, 0x00, 0x2f, 0xef, 0x15, 0x2c}
	rootParentFP := [4]byte{0, 0, 0, 0}

	rooPubKey := NewExtendedKey(rootPub[:], rootChainCode[:], rootParentFP[:], 0, 0, false, owcrypt.ECC_CURVE_SECP256K1)
	serialize := uint32(0x80000000)

	childPubKey, err := rooPubKey.GenPublicChild(serialize)
	fmt.Println("父公钥 -----> 强化子公钥")
	if childPubKey != nil {
		t.Error("未检出父公钥扩展强化子密钥的非法操作")
	}
	if err != ErrDeriveHardFromPublic {
		t.Error("抛出了错误的异常")
	}
}

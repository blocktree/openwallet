package bliss

import (
	"io"

	dcrcrypto "github.com/HcashOrg/hcd/crypto"
	"github.com/HcashOrg/bliss/poly"
)

type DSA interface {

	// ----------------------------------------------------------------------------
	// Private keys
	//
	// NewPrivateKey instantiates a new private key for the given data
	NewPrivateKey(s1, s2, a *poly.PolyArray) dcrcrypto.PrivateKey

	// PrivKeyFromBytes calculates the public key from serialized bytes,
	// and returns both it and the private key.
	PrivKeyFromBytes(pk []byte) (dcrcrypto.PrivateKey, dcrcrypto.PublicKey)

	// PrivKeyBytesLen returns the length of a serialized private key.
	PrivKeyBytesLen() int

	// ----------------------------------------------------------------------------
	// Public keys
	//
	// NewPublicKey instantiates a new public key (point) for the given data.
	NewPublicKey(a *poly.PolyArray) dcrcrypto.PublicKey

	// ParsePubKey parses a serialized public key for the given
	// curve and returns a public key.
	ParsePubKey(pubKeyStr []byte) (dcrcrypto.PublicKey, error)

	// PubKeyBytesLen returns the length of the default serialization
	// method for a public key.
	PubKeyBytesLen() int

	// ----------------------------------------------------------------------------
	// Signatures
	//
	// NewSignature instantiates a new signature
	NewSignature(z1, z2 *poly.PolyArray, c []uint32) dcrcrypto.Signature

	// ParseDERSignature parses a DER encoded signature .
	// If the method doesn't support DER signatures, it
	// just parses with the default method.
	ParseDERSignature(sigStr []byte) (dcrcrypto.Signature, error)

	// ParseSignature a default encoded signature
	ParseSignature(sigStr []byte) (dcrcrypto.Signature, error)

	// RecoverCompact recovers a public key from an encoded signature
	// and message, then verifies the signature against the public
	// key.
	RecoverCompact(signature, hash []byte) (dcrcrypto.PublicKey, bool, error)

	// ----------------------------------------------------------------------------
	// Bliss
	//
	// GenerateKey generates a new private and public keypair from the
	// given reader.
	GenerateKey(rand io.Reader) (dcrcrypto.PrivateKey, dcrcrypto.PublicKey, error)

	// Sign produces an Bliss signature using a private key and a message.
	Sign(priv dcrcrypto.PrivateKey, hash []byte) (dcrcrypto.Signature, error)

	// Verify verifies an Bliss signature against a given message and
	// public key.
	Verify(pub dcrcrypto.PublicKey, hash []byte, sig dcrcrypto.Signature) bool
}

const (
	BSTypeBliss = 4

	BlissVersion = 1

	BlissPubKeyLen = 897

	BlissPrivKeyLen = 385
)

// Secp256k1 is the secp256k1 curve and ECDSA system used in Bitcoin.
var Bliss = newBlissDSA()

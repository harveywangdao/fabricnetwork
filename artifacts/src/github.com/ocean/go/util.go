package main

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"regexp"
)

func GetNewAddress() (string, string, string) {
	priv, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", ""
	}

	wif, err := btcutil.NewWIF(priv, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", "", ""
	}

	pubKeyBytes := wif.SerializePubKey()

	address, err := btcutil.NewAddressPubKey(pubKeyBytes, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", ""
	}

	return wif.String(), hex.EncodeToString(pubKeyBytes), address.EncodeAddress()
}

func Verify(pubKeyHexStr, originStr, signHexStr string) (bool, error) {
	if pubKeyHexStr == "" || originStr == "" || signHexStr == "" {
		return false, nil
	}

	// Decode hex-encoded serialized public key.
	pubKeyBytes, err := hex.DecodeString(pubKeyHexStr)
	if err != nil {
		return false, err
	}

	pubKey, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
	if err != nil {
		return false, err
	}

	// Decode hex-encoded serialized signature.
	sigBytes, err := hex.DecodeString(signHexStr)
	if err != nil {
		return false, err
	}

	signature, err := btcec.ParseSignature(sigBytes, btcec.S256())
	if err != nil {
		return false, err
	}

	// Verify the signature for the message using the public key.
	originHash := chainhash.HashB([]byte(originStr))

	return signature.Verify(originHash, pubKey), nil
}

func Sign(privKeyWif string, originData []byte) (string, error) {
	wif, err := btcutil.DecodeWIF(privKeyWif)
	if err != nil {
		return "", err
	}

	// Sign a message using the private key.
	originHash := chainhash.HashB(originData)

	signature, err := wif.PrivKey.Sign(originHash)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(signature.Serialize()), nil
}

//get address by public key
func GetAddress(pubKeyHexStr string) string {
	if pubKeyHexStr == "" {
		return ""
	}

	// Decode hex-encoded serialized public key.
	pubKeyBytes, err := hex.DecodeString(pubKeyHexStr)
	if err != nil {
		return ""
	}

	address, err := btcutil.NewAddressPubKey(pubKeyBytes, &chaincfg.MainNetParams)
	if err != nil {
		return ""
	}

	return address.EncodeAddress()
}

func GetPubKeyByPrivKey(privKeyWif string) (string, error) {
	wif, err := btcutil.DecodeWIF(privKeyWif)
	if err != nil {
		return "", err
	}

	pubKeyBytes := wif.SerializePubKey()

	return hex.EncodeToString(pubKeyBytes), nil
}

func IsGtZeroInteger(s string) bool {
	pattern := `^\+?[1-9]\d*$`
	match, _ := regexp.MatchString(pattern, s)
	return match
}

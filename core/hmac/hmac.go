// Package hmac provides functions for validating hmac signatures.
package hmac

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"strings"
)

var (
	errSigCheck          = errors.New("payload signature check failed")
	errSigMissing        = errors.New("signature is missing")
	errAllowedVerMissing = errors.New("allowed version is missing")
	errSigParse          = errors.New("signature parsing failed")
	errSigVer            = errors.New("signature version invalid")
	errSigDecode         = errors.New("signature decoding failed")
)

// ValidateSignature validates the signature for the given message based on
// the allowed version such as sha1 or v1 depending on how provider chose.
func ValidateSignature(allowedVersion string, signature string, payload, secretToken []byte) error {
	messageMAC, hashFunc, err := messageMAC(allowedVersion, signature)
	if err != nil {
		return err
	}
	if !checkMAC(payload, messageMAC, secretToken, hashFunc) {
		return errSigCheck
	}

	return nil
}

// messageMAC returns the hex-decoded HMAC tag from the signature and its
// corresponding hash function.
func messageMAC(allowedVersion string, signature string) ([]byte, func() hash.Hash, error) {
	if signature == "" {
		return nil, nil, errSigMissing
	}
	if allowedVersion == "" {
		return nil, nil, errAllowedVerMissing
	}
	sigParts := strings.SplitN(signature, "=", 2)
	if len(sigParts) != 2 {
		return nil, nil, fmt.Errorf("%w: %q", errSigParse, signature)
	}

	if sigParts[0] != allowedVersion {
		return nil, nil, fmt.Errorf("%w: %q", errSigVer, sigParts[0])
	}

	var hashFunc func() hash.Hash
	switch sigParts[0] {
	case "v1":
		fallthrough
	default:
		hashFunc = sha256.New
	}

	buf, err := hex.DecodeString(sigParts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("%w %q: %v", errSigDecode, signature, err)
	}

	return buf, hashFunc, nil
}

// genMAC generates the HMAC signature for a message provided the secret key
// and hashFunc.
func genMAC(message, key []byte, hashFunc func() hash.Hash) []byte {
	mac := hmac.New(hashFunc, key)
	mac.Write(message)

	return mac.Sum(nil)
}

// checkMAC reports whether messageMAC is a valid HMAC tag for message.
func checkMAC(message, messageMAC, key []byte, hashFunc func() hash.Hash) bool {
	expectedMAC := genMAC(message, key, hashFunc)

	return hmac.Equal(messageMAC, expectedMAC)
}

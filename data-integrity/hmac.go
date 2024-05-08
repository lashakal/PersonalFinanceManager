package data_integrity

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	_ "log"
	"os"
)

// function to get  HMAC secret key from an environment variable
func getHMACSecretKey() []byte {
	// get the secret key from an environment variable
	key := os.Getenv("HMAC_KEY")

	return []byte(key)
}

// generate an HMAC for the given message and key [sha256]
func generateHMAC(message []byte) string {
	key := getHMACSecretKey()
	hmacHash := hmac.New(sha256.New, key)
	hmacHash.Write(message)
	hashedMessage := hmacHash.Sum(nil)
	return hex.EncodeToString(hashedMessage)
}

// add an HMAC to the message and return the combined payload
func AddHMAC(message []byte) []byte {
	hmacValue := generateHMAC(message)
	return append(message, []byte(hmacValue)...)
}

// verify the integrity of the message and returns true if it's valid
func VerifyHMAC(payload []byte) (bool, []byte) {
	if len(payload) < 64 {
		// The payload must have at least 64 characters for the HMAC
		return false, nil
	}

	message := payload[:len(payload)-64]
	receivedHMAC := payload[len(payload)-64:]

	// recalculate HMAC to compare them later
	calculatedHMAC := generateHMAC(message)

	// compare recalculated and received HMACs
	// return comparison boolean value, as well as the message itself
	return hmac.Equal([]byte(calculatedHMAC), receivedHMAC), message
}

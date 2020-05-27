package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"
	"unicode"

	"github.com/ebfe/scard"
	keycard "github.com/status-im/keycard-go"
	"github.com/status-im/keycard-go/apdu"
	"github.com/status-im/keycard-go/derivationpath"
	kio "github.com/status-im/keycard-go/io"
	ktypes "github.com/status-im/keycard-go/types"
	"golang.org/x/crypto/ssh/terminal"
)

func createKeycardCmdSet(card *scard.Card) *keycard.CommandSet {
	channel := kio.NewNormalChannel(card)
	return keycard.NewCommandSet(channel)
}

func selectKeycard(cmdSet *keycard.CommandSet) {
	if err := cmdSet.Select(); err != nil {
		fail("Error: Select failed")
	}

	if !cmdSet.ApplicationInfo.Installed {
		fail("Error: Initialization failed")
	}

	if !cmdSet.ApplicationInfo.Initialized {
		fail("Error: Card not initialized")
	}

	if len(cmdSet.ApplicationInfo.KeyUID) == 0 {
		fail("Error: No wallet found")
	}
}

func getPairingPath() string {
	home, _ := os.UserHomeDir()
	return path.Join(home, ".keycard-crypt-pairings")
}

func requestPassword(prompt string) string {
	fmt.Print(prompt)
	password, _ := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	return strings.TrimSpace(string(password))
}

func allDigits(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}

	return true
}

func authenticatePIN(cmdSet *keycard.CommandSet) {
	pin := requestPassword("Enter your PIN: ")

	if len(pin) == 6 && allDigits(pin) {
		if err := cmdSet.VerifyPIN(pin); err != nil {
			fail("Error: Wrong PIN or blocked card")
		}
	} else {
		fail("Error: The PIN must be 6 digits")
	}
}

func readPairings() map[string]*ktypes.PairingInfo {
	pairings := make(map[string]*ktypes.PairingInfo)
	dat, err := ioutil.ReadFile(getPairingPath())

	if err != nil {
		return pairings
	}

	for i := 0; i < len(dat); i += 49 {
		instanceID := string(dat[i : i+16])
		key := dat[i+16 : i+48]
		idx := dat[i+48]

		pairings[instanceID] = &ktypes.PairingInfo{Key: key, Index: int(idx)}
	}

	return pairings
}

func writePairing(instanceID []byte, pairing *ktypes.PairingInfo) {
	f, err := os.OpenFile(getPairingPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fail("Error: Can't write pairing")
	}

	defer f.Close()

	f.Write(instanceID)
	f.Write(pairing.Key)
	f.Write([]byte{byte(pairing.Index)})
}

func setupPairing(cmdSet *keycard.CommandSet) {
	pairings := readPairings()
	pairing, ok := pairings[string(cmdSet.ApplicationInfo.InstanceUID)]

	if ok {
		cmdSet.PairingInfo = pairing
	} else {
		password := requestPassword("Enter the pairing password: ")

		if err := cmdSet.Pair(password); err != nil {
			fail("Error: Wrong password")
		}

		writePairing(cmdSet.ApplicationInfo.InstanceUID, cmdSet.PairingInfo)
	}
}

func authentication(cmdSet *keycard.CommandSet) {
	setupPairing(cmdSet)

	if err := cmdSet.OpenSecureChannel(); err != nil {
		fail("Error: Couldn't open secure channel")
	}

	authenticatePIN(cmdSet)
}

func getEncryptionKey(cmdSet *keycard.CommandSet, keyPath []byte) []byte {
	encodedPath, _ := derivationpath.EncodeFromBytes(keyPath)

	fmt.Println("using derivation path", encodedPath)

	data, err := cmdSet.ExportKey(true, false, false, encodedPath)

	if err != nil {
		fail("Error: Couldn't generate export key")
	}

	key, _ := apdu.FindTag(data, apdu.Tag{0xa1}, apdu.Tag{0x81})

	return key
}

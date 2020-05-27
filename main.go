package main

import (
	"flag"
	"fmt"
	"os"
)

func fail(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func main() {
	keepOriginals := flag.Bool("keep-originals", false, "keep original files")
	flag.Parse()
	clArgs := flag.Args()

	fmt.Println(clArgs)

	if len(clArgs) < 2 {
		fail("Error: args missing")
	}

	var cmd CommandFunc

	switch clArgs[0] {
	case "encrypt":
		cmd = encryptFile
	case "decrypt":
		cmd = decryptFile
	default:
		fail("Error: Wrong command")
	}

	ctx := createContext()
	defer releaseContext(ctx)

	card := connectCard(ctx)
	defer disconnectCard(card)
	getCardStatus(card)

	cmdSet := createKeycardCmdSet(card)
	selectKeycard(cmdSet)
	authentication(cmdSet)

	processFiles(cmdSet, clArgs[1:], cmd, *keepOriginals)
}

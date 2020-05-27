package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/minio/sio"
	keycard "github.com/status-im/keycard-go"
)

type CommandFunc func(*keycard.CommandSet, string) bool

var baseDerivationPath = []byte{
	0x80, 0x00, 0x00, 0x2b,
	0x80, 0x00, 0x00, 0x3c,
	0x80, 0x00, 0x06, 0x2d,
	0x80, 0x00, 0x00, 0x01,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
}

var magicNumber = []byte{0x4b, 0x45}

func plainFileName(file string) string {
	if strings.HasSuffix(file, ".kef") {
		return strings.TrimSuffix(file, ".kef")
	}

	return file + ".plain"
}

func generateDerivationPath() []byte {
	randPart := baseDerivationPath[16:]
	rand.Read(randPart)

	for i := 0; i < 20; i += 4 {
		randPart[i] |= 0x80
	}

	randPart[20] &= 0x7f

	return baseDerivationPath
}

func encryptFile(cmdSet *keycard.CommandSet, file string) bool {
	derPath := generateDerivationPath()
	derPathLen := []byte{byte(len(derPath))}
	key := getEncryptionKey(cmdSet, derPath)
	dstFilePath := file + ".kef"

	src, err := os.Open(file)
	if err != nil {
		fmt.Println("Error: Couldn't open ", file)
		return false
	}

	defer src.Close()

	dst, err := os.Create(dstFilePath)
	if err != nil {
		fmt.Println("Error: Couldn't create ", dstFilePath)
		return false
	}

	defer dst.Close()

	dst.Write(magicNumber)
	dst.Write(cmdSet.ApplicationInfo.KeyUID)
	dst.Write(derPathLen)
	dst.Write(derPath)

	if _, err := sio.Encrypt(dst, src, sio.Config{Key: key}); err != nil {
		fmt.Println("Error: Couldn't encrypt ", file)
		return false
	}

	return true
}

func decryptFile(cmdSet *keycard.CommandSet, file string) bool {
	src, err := os.Open(file)
	if err != nil {
		fmt.Println("Error: Couldn't open ", file)
		return false
	}

	defer src.Close()

	header := make([]byte, 35)
	src.Read(header)

	if magNum := header[0:2]; bytes.Compare(magNum, magicNumber) != 0 {
		fmt.Println("Errror: Invalid file type for ", file)
		return false
	}

	if keyUID := header[2:34]; bytes.Compare(keyUID, cmdSet.ApplicationInfo.KeyUID) != 0 {
		fmt.Println("Errror: Wrong card inserted for ", file)
		return false
	}

	derPath := make([]byte, header[34])
	src.Read(derPath)

	key := getEncryptionKey(cmdSet, derPath)

	dst, err := os.Create(plainFileName(file))
	if err != nil {
		fmt.Println("Error: Couldn't create ", plainFileName(file))
		return false
	}

	defer dst.Close()

	if _, err := sio.Decrypt(dst, src, sio.Config{Key: key}); err != nil {
		fmt.Println("Error: Failed to decrypt ", file)
		return false
	}

	return true
}

func processFiles(cmdSet *keycard.CommandSet, files []string, fn CommandFunc, keepOriginals bool) {
	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err == nil {
			if fileInfo.Mode().IsDir() {
				dirFiles, _ := ioutil.ReadDir(file)
				for _, dirFile := range dirFiles {
					if !strings.HasPrefix(dirFile.Name(), ".") {
						dirFilePath := path.Join(file, dirFile.Name())
						processFiles(cmdSet, []string{dirFilePath}, fn, keepOriginals)
					}
				}
			} else if fileInfo.Mode().IsRegular() {
				if fn(cmdSet, file) && !keepOriginals {
					os.Remove(file)
				}
			}
		}
	}
}

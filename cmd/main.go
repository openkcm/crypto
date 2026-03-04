package main

import (
	"crypto/aes"
	"fmt"
	"time"

	"github.com/openkcm/krypton/internal/memvault"
)

func main() {
	// nonce := []byte("unique_nonce")
	// clearText := []byte("...encrypted_data...")

	vaultMainKey, err := memvault.NewWithSecret([]byte("passphrasewhichneedstobe32bytes!"))
	if err != nil {
		panic(err)
	}

	process(vaultMainKey.Bytes())
	_, err = aes.NewCipher(vaultMainKey.Bytes())
	if err != nil {
		panic(err)
	}

	err = vaultMainKey.Wipe()
	if err != nil {
		panic(err)
	}

	fmt.Println("wipe finished")

	for {
		time.Sleep(10 * time.Second)
	}
}

func process(a []byte) {
	for i := range a {
		fmt.Println("xxx", a[i])
	}
}

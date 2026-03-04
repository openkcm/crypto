package main

import (
	"fmt"

	"github.com/openkcm/krypton/internal/memvault"
)

func main() {
	// nonce := []byte("unique_nonce")
	// clearText := []byte("...encrypted_data...")

	vaultMasterKey, err := memvault.NewWithSecret([]byte("passphrasewhichneedstobe32bytes!"))
	if err != nil {
		panic(err)
	}

	process(vaultMasterKey.Bytes())
	// _, err = aes.NewCipher(vaultMasterKey.Bytes())
	// if err != nil {
	// 	panic(err)
	// }

	err = vaultMasterKey.Wipe()
	if err != nil {
		panic(err)
	}

	fmt.Println("wipe finished")

	for {
	}
}

func process(a []byte) {
	for i := range a {
		fmt.Println("xxx", a[i])
	}
}

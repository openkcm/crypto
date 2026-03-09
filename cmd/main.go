package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/awnumar/memcall"
	"github.com/awnumar/memguard"
	"github.com/openkcm/krypton/internal/memvault"
)

func init() {
	// memcall.Configure(memguard.DefaultConfig) // Enables all defenses
	memcall.DisableCoreDumps()
	// RLIMIT_CORE=0 explicitly
	memguard.CatchInterrupt()
}

func main() {
	// nonce := []byte("unique_nonce")
	// clearText := []byte("...clearText...")
	//
	vault, err := memvault.NewWithSize(10)
	if err != nil {
		panic(err)
	}
	data := vault.Bytes()
	for i := range data {
		data[i] = byte('a' + i)
	}

	buf := memguard.NewBuffer(10)
	data2 := buf.Data()
	for i := range data2 {
		data2[i] = byte('A' + i)
	}

	// secret.Do(func() {
	// 	err := memvault.Run(context.Background(), func(ctx context.Context, vaultMainKey *memvault.MemVault) error {
	// 		err := vaultMainKey.WithSecret([]byte("passphrasewhichneedstobe32bytes!"))
	// 		if err != nil {
	// 			return err
	// 		}
	//
	// 		block, err := aes.NewCipher(vaultMainKey.Bytes())
	// 		if err != nil {
	// 			return err
	// 		}
	//
	// 		aesGCM, err := cipher.NewGCM(block)
	// 		if err != nil {
	// 			return err
	// 		}
	//
	// 		return memvault.Run(ctx, func(ctx context.Context, encryptedStore *memvault.MemVault) error {
	// 			err = encryptedStore.WithSize(31)
	// 			if err != nil {
	// 				return err
	// 			}
	//
	// 			aesGCM.Seal(encryptedStore.Bytes()[:0], nonce, clearText, nil)
	//
	// 			return memvault.Run(ctx, func(ctx context.Context, decryptedStore *memvault.MemVault) error {
	// 				err := decryptedStore.WithSize(len(clearText))
	// 				if err != nil {
	// 					return err
	// 				}
	//
	// 				_, err = aesGCM.Open(decryptedStore.Bytes()[:0], nonce, encryptedStore.Bytes(), nil)
	// 				if err != nil {
	// 					return err
	// 				}
	//
	// 				return nil
	// 			})
	// 		})
	// 	})
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// })
	//
	for {
		runtime.GC()
		fmt.Println("GC finished")
		time.Sleep(10 * time.Second)
	}
}

func process(a []byte) {
	for i := range a {
		fmt.Println("xxx", a[i])
	}
}

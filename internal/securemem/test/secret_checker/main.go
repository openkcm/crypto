package main

import (
	"context"
	"crypto/aes"
	"fmt"
	"os"
	"time"

	"github.com/openkcm/krypton/internal/securemem"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC RECOVERED: %v\n", r)
		}
	}()

	resp, err := securemem.Run(context.Background(), func(ctx context.Context, hr *securemem.HandlerRequest) error {
		b := []byte("MYSECRETKEY123458901234567890123")
		secret, err := hr.PersistentVault().Reserve("secret", len(b))
		if err != nil {
			return err
		}
		copy(secret, b)

		tmpSecret, err := hr.TmpVault().Reserve("tmp_secret", len(b))
		if err != nil {
			return err
		}
		copy(tmpSecret, b)

		_, err = aes.NewCipher(b)
		if err != nil {
			panic(fmt.Sprintf("Failed to create AES cipher: %v", err))
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Checking for secrets in MemVault...")
	_, ok := resp.MemVault().Get("secret")
	if !ok {
		fmt.Println("❌ SECRET NOT FOUND IN MEMVAULT")
	} else {
		fmt.Println("✅ SECRET FOUND IN MEMVAULT")
	}

	exposedSecret := "EXPOSED_SECRET123456789012345678"
	_, err = aes.NewCipher([]byte(exposedSecret))
	if err != nil {
		panic(fmt.Sprintf("Failed to create AES cipher with exposed secret: %v", err))
	}

	isCreated := false
	for {
		if !isCreated {
			_, err := os.Create("start")
			if err != nil {
				fmt.Printf("Error creating file: %v\n", err)
				continue
			}
			isCreated = true
		}
		time.Sleep(10 * time.Second)
	}
}

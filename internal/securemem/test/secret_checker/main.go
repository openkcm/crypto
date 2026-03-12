package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
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

		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Checking for secrets in MemVault...")
	_, ok := resp.MemVault().Get("secret")
	if !ok {
		fmt.Println("❌ Secret not found in MemVault")
	} else {
		fmt.Println("✅ Secret found in MemVault")
	}

	isCreated := false
	for {
		runtime.GC()

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

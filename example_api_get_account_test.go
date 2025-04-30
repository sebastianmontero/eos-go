package eos_test

import (
	"context"
	"encoding/json"
	"fmt"

	eos "github.com/sebastianmontero/eos-go"
)

func ExampleAPI_GetAccount() {
	api, err := eos.New(getAPIURL())
	if err != nil {
		panic(fmt.Errorf("new api: %w", err))
	}

	account := eos.AccountName("eos.rex")
	info, err := api.GetAccount(context.Background(), account)
	if err != nil {
		if err == eos.ErrNotFound {
			fmt.Printf("unknown account: %s", account)
			return
		}

		panic(fmt.Errorf("get account: %w", err))
	}

	bytes, err := json.Marshal(info)
	if err != nil {
		panic(fmt.Errorf("json marshal response: %w", err))
	}

	fmt.Println(string(bytes))
}

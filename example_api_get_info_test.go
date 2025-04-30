package eos_test

import (
	"context"
	"encoding/json"
	"fmt"

	eos "github.com/sebastianmontero/eos-go"
)

func ExampleAPI_GetInfo() {
	api, err := eos.New(getAPIURL())
	if err != nil {
		panic(fmt.Errorf("new api: %w", err))
	}

	info, err := api.GetInfo(context.Background())
	if err != nil {
		panic(fmt.Errorf("get info: %w", err))
	}

	bytes, err := json.Marshal(info)
	if err != nil {
		panic(fmt.Errorf("json marshal response: %w", err))
	}

	fmt.Println(string(bytes))
}

package main

import (
	"context"
	"testing"
)

func TestGetApplicationSecrets(t *testing.T) {
	secret, err := GetApplicationSecrets(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("We found a secret")
	t.Log(secret)
}

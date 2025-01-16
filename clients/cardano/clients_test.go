package cardano

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"testing"
)

func TestCardanoClient_GetBlock(t *testing.T) {
	_, err := NewCardanoClient(context.Background(), hclog.NewNullLogger())
	if err != nil {
		t.Fatal(err)
	}
}

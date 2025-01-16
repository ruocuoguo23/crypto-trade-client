package cardano

import (
	"context"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/hashicorp/go-hclog"
	"net/http"
	"time"
)

type AdaClient struct {
	ctx               context.Context
	client            *client.APIClient
	log               hclog.Logger
	networkIdentifier *types.NetworkIdentifier
}

func NewCardanoClient(ctx context.Context, logger hclog.Logger) (*AdaClient, error) {
	log := logger.Named("cardano-wallet-svc")
	// cardano rosetta client
	cfg := client.NewConfiguration(
		"https://cardano-mainnet.gateway.tatum.io",
		"cardano",
		&http.Client{Timeout: 30 * time.Second})

	cfg.AddDefaultHeader("x-api-key", "t-6747e0b03c7ce02d00954859-c32a2ec9df8a43929da819c8")

	apiClient := client.NewAPIClient(cfg)

	// request network identifier
	networkListReq := &types.MetadataRequest{
		Metadata: map[string]interface{}{},
	}
	networkListRsp, rosettaErr, err := apiClient.NetworkAPI.NetworkList(ctx, networkListReq)
	if err != nil {
		return nil, err
	}
	if rosettaErr != nil {
		return nil, err
	}
	if len(networkListRsp.NetworkIdentifiers) != 1 {
		return nil, fmt.Errorf("there is no cardano network")
	}

	adaClient := &AdaClient{
		ctx:               ctx,
		client:            apiClient,
		log:               log,
		networkIdentifier: networkListRsp.NetworkIdentifiers[0],
	}

	return adaClient, nil
}

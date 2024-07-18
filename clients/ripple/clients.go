package ripple

import "github.com/hashicorp/go-hclog"

type XrpClient struct {
	*XrpRpc
}

// NewXrpClient creates a new Ripple client with the given endpoint and private key
func NewXrpClient(endpoint string) (*XrpClient, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name: "ripple-client",
	})
	client, err := NewXrpRpc(endpoint, logger)
	if err != nil {
		return nil, err
	}

	return &XrpClient{client}, nil
}

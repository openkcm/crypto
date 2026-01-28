package kmiptest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/openkcm/krypton/kmip/kmipserver"
	"github.com/openkcm/krypton/kmip/payloads"
)

func TestClientServer(t *testing.T) {
	client := NewClientAndServer(t, kmipserver.NewBatchExecutor())
	resp, err := client.Request(context.Background(), &payloads.DiscoverVersionsRequestPayload{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

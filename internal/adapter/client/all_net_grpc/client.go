// Package all_net_grpc is an alternative AllNetClient that talks to a
// sega-idac-client gRPC server instead of scraping All.Net's HTML directly. It
// implements the same port.AllNetClient interface as the HTTP client, so it is a
// drop-in replacement selected in main by SEGA_IDAC_GRPC_ADDR.
package all_net_grpc

import (
	"context"

	allnetv1 "github.com/Duke40199/sega-idac-client/gen/allnetv1"
	"google.golang.org/grpc"

	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
)

type AllNetClient struct {
	client allnetv1.AllNetServiceClient
}

// NewAllNetClient builds a gRPC-backed AllNetClient over an existing connection.
func NewAllNetClient(conn grpc.ClientConnInterface) port.AllNetClient {
	return &AllNetClient{client: allnetv1.NewAllNetServiceClient(conn)}
}

func (c *AllNetClient) GetListStore(ctx context.Context, gameCode, languageCode, areaCode string) ([]entity.StoreLocation, string, error) {
	resp, err := c.client.GetListStore(ctx, &allnetv1.GetListStoreRequest{
		GameCode:     gameCode,
		LanguageCode: languageCode,
		AreaCode:     areaCode,
	})
	if err != nil {
		return nil, "", err
	}
	stores := make([]entity.StoreLocation, 0, len(resp.GetStores()))
	for _, s := range resp.GetStores() {
		store := entity.StoreLocation{
			ID:             s.GetId(),
			Name:           s.GetName(),
			Address:        s.GetAddress(),
			SegaAreaCode:   s.GetSegaAreaCode(),
			AllNetAreaCode: s.GetAllNetAreaCode(),
		}
		if ts := s.GetCreatedAt(); ts != nil {
			store.CreatedAt = ts.AsTime()
		}
		if ts := s.GetUpdatedAt(); ts != nil {
			store.UpdatedAt = ts.AsTime()
		}
		stores = append(stores, store)
	}
	return stores, resp.GetAreaName(), nil
}

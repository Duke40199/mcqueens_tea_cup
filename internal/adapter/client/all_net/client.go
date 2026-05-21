package all_net

import (
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/port"
)

type AllNetClient struct {
	AllNetCfg *config.AllNetClientConfig
}

func NewAllNetClient(segaIdacCfg *config.SegaClientConfig) port.SegaIDACClient {
	return &AllNetClient{
		SegaIdacCfg: segaIdacCfg,
	}
}

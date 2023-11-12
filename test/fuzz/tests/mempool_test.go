//go:build gofuzz || go1.20

package tests

import (
	"testing"

	abciclient "github.com/cometbft/cometbft/abci/client"
	"github.com/cometbft/cometbft/abci/example/kvstore"
	"github.com/cometbft/cometbft/config"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	mempool "github.com/cometbft/cometbft/mempool"
	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
)

func FuzzMempool(f *testing.F) {
	app := kvstore.NewInMemoryApplication()
	mtx := new(cmtsync.Mutex)
	conn := abciclient.NewLocalClient(mtx, app)
	err := conn.Start()
	if err != nil {
		panic(err)
	}

	cfg := config.DefaultMempoolConfig()
	cfg.Broadcast = false

	txDecoder := types.DefaultTxDecoder
	mp := mempool.NewCListMempool(cfg, conn, 0, types.DefaultTxDecoder)

	f.Fuzz(func(t *testing.T, data []byte) {
		tx, err := txDecoder(data)
		require.NoError(t, err)
		_, _ = mp.CheckTx(tx)
	})
}

package evm_wallet

import (
	"context"

	evmWalletModel "github.com/zd4r/frame-claim/internal/model/evm_wallet"
)

type evmWalletStore interface {
	GetList(ctx context.Context) ([]evmWalletModel.EvmWallet, error)
}

type passphraseStore interface {
	Set(val []byte)
	Get() string
}

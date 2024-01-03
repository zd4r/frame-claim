package evm_wallet

import (
	"context"
	"fmt"

	evmWalletModel "github.com/zd4r/frame-claim/internal/model/evm_wallet"
	"github.com/zd4r/frame-claim/internal/storage/sqlite"
)

type SQLite struct {
	db *sqlite.Storage
}

func New(db *sqlite.Storage) *SQLite {
	return &SQLite{db}
}

func (s *SQLite) GetList(ctx context.Context) ([]evmWalletModel.EvmWallet, error) {
	const op = "store.wallet.sqlite.GetList"

	stmt, err := s.db.Prepare("SELECT id, name, address FROM evm_wallet")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var wallets []evmWalletModel.EvmWallet
	for rows.Next() {
		var wallet evmWalletModel.EvmWallet

		if err := rows.Scan(&wallet.ID, &wallet.Name, &wallet.Address); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		wallets = append(wallets, wallet)
	}

	return wallets, nil
}

package evm_wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	evmWalletModel "github.com/zd4r/frame-claim/internal/model/evm_wallet"
)

type Service struct {
	walletStore     evmWalletStore
	keyStore        *keystore.KeyStore
	passphraseStore passphraseStore
}

func New(walletStore evmWalletStore, keyStore *keystore.KeyStore, passphraseStore passphraseStore) *Service {
	return &Service{
		walletStore:     walletStore,
		keyStore:        keyStore,
		passphraseStore: passphraseStore,
	}
}

func (s *Service) GetList(ctx context.Context) ([]evmWalletModel.EvmWallet, error) {
	return s.walletStore.GetList(ctx)
}

func (s *Service) Sigh(address string, msg []byte) ([]byte, error) {
	acc, err := s.keyStore.Find(accounts.Account{Address: common.HexToAddress(address)})
	if err != nil {
		return nil, err
	}

	return s.keyStore.SignHashWithPassphrase(acc, s.passphraseStore.Get(), msg)
}

func (s *Service) CheckAccess() error {
	for _, account := range s.keyStore.Accounts() {
		if err := s.keyStore.TimedUnlock(
			account,
			s.passphraseStore.Get(),
			1*time.Microsecond,
		); err != nil {
			return fmt.Errorf("failed to unlock keystore: %w", err)
		}
	}

	return nil
}

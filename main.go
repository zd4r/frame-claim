package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ethereum/go-ethereum/crypto"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	frameClient "github.com/zd4r/frame-claim/internal/client/frame"
	evmWalletService "github.com/zd4r/frame-claim/internal/service/evm_wallet"
	"github.com/zd4r/frame-claim/internal/storage/sqlite"
	evmWalletStore "github.com/zd4r/frame-claim/internal/store/evm_wallet"
	"github.com/zd4r/frame-claim/internal/store/passphrase"
	"golang.org/x/term"
)

const (
	keystoreDirPath = "/.ethereum/keystore"
	databasePath    = "/.ethereum/wallet-manager.db"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	cxt := context.Background()

	// init global passphrase store
	pp := passphrase.New()

	// set passphrase
	password, err := readPassphrase()
	if err != nil {
		log.Fatal(err)
	}
	pp.Set(password)

	// get $HOME
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// init keystore
	ks := keystore.NewKeyStore(
		fmt.Sprintf("%s%s", homeDir, keystoreDirPath), // TODO: change path to $HOME/.ethereum/keystore
		keystore.StandardScryptN,
		keystore.StandardScryptP,
	)

	// init bot data storage
	storage, err := sqlite.NewWithContext(cxt, fmt.Sprintf("%s%s", homeDir, databasePath))
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Stop()

	// init wallet service
	evmWalletSrv := evmWalletService.New(
		evmWalletStore.New(storage),
		ks,
		pp,
	)
	if err := evmWalletSrv.CheckAccess(); err != nil {
		log.Fatal(err)
	}

	// addresses
	wallets, err := evmWalletSrv.GetList(cxt)
	if err != nil {
		log.Fatal(err)
	}

	// frame client
	client := frameClient.New()

	// table
	table := tabwriter.NewWriter(os.Stdout, 1, 1, 3, ' ', 0)
	fmt.Fprintln(table, "name\t"+"address\t"+"totalAllocation\t"+"hasClaimedPoints\t"+"pointsClaimed\t"+"error\t")

	// progress bar
	bar := progressbar.Default(int64(len(wallets)))

	// fill table
	var totalPoints int
	for _, wallet := range wallets {
		func() {
			defer bar.Add(1)

			// sign msg
			data := fmt.Sprintf(
				"You are claiming the Frame Chapter One Airdrop with the following address: %s",
				strings.ToLower(wallet.Address),
			)
			msg := crypto.Keccak256(
				[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)),
			)
			signature, err := evmWalletSrv.Sigh(wallet.Address, msg)
			signature[64] += 27

			// auth
			authReqBody := &frameClient.AuthenticateReqBody{
				Address:   wallet.Address,
				Signature: fmt.Sprintf("0x%x", signature),
			}
			authRespBody, err := client.Authenticate(authReqBody)
			if err != nil {
				fmt.Fprintf(table,
					"%s\t%s\t-\t-\t-\tauthenticate: %s\t\n",
					wallet.Name,
					wallet.GetShortAddress(),
					err.Error(),
				)
				return
			}

			if authRespBody.UserInfo.HasClaimedPoints || authRespBody.UserInfo.TotalAllocation == 0 {
				totalPoints += authRespBody.UserInfo.TotalAllocation
				fmt.Fprintf(table,
					"%s\t%s\t%d\t%v\t%s\t-\t\n",
					wallet.Name,
					wallet.GetShortAddress(),
					authRespBody.UserInfo.TotalAllocation,
					authRespBody.UserInfo.HasClaimedPoints,
					authRespBody.UserInfo.PointsClaimed,
				)
				return
			}

			// claim
			_, err = client.Claim(authRespBody.Token)
			if err != nil {
				fmt.Fprintf(table,
					"%s\t%s\t%d\t%v\t%s\tclaim: %s\t\n",
					wallet.Name,
					wallet.GetShortAddress(),
					authRespBody.UserInfo.TotalAllocation,
					authRespBody.UserInfo.HasClaimedPoints,
					authRespBody.UserInfo.PointsClaimed,
					err.Error(),
				)
				return
			}

			// check if claimed
			userReqBody, err := client.User(authRespBody.Token)
			if err != nil {
				fmt.Fprintf(table,
					"%s\t%s\t%d\t%v\t%s\tuser: %s\t\n",
					wallet.Name,
					wallet.GetShortAddress(),
					authRespBody.UserInfo.TotalAllocation,
					authRespBody.UserInfo.HasClaimedPoints,
					authRespBody.UserInfo.PointsClaimed,
					err.Error(),
				)
				return
			}
			totalPoints += authRespBody.UserInfo.TotalAllocation

			// add row
			fmt.Fprintf(table,
				"%s\t%s\t%d\t%v\t%s\t-\t\n",
				wallet.Name,
				wallet.GetShortAddress(),
				userReqBody.TotalAllocation,
				userReqBody.HasClaimedPoints,
				userReqBody.PointsClaimed,
			)
		}()
	}

	fmt.Println()

	table.Flush()

	fmt.Printf("total: %d\n", totalPoints)
}

func readPassphrase() ([]byte, error) {
	defer fmt.Printf("\n\n")

	fmt.Print("password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to term.ReadPassword: %w", err)
	}

	return password, nil
}

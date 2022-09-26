package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/osmosis-labs/osmosis/v11/app"
	"github.com/pomifer/scriptor/cmd/history"
	"github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
)

var defaultKeyHome string

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	defaultKeyHome = filepath.Join(userHomeDir, ".tx-gen")
}

func main() {

	encodingConfig := app.MakeEncodingConfig()

	// cfg := sdk.GetConfig()
	// cfg.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	// cfg.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	// cfg.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	// cfg.Seal()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(defaultKeyHome).
		WithViper("SCRIPTOR")

	rootCmd := &cobra.Command{
		Use:   "scriptor",
		Short: "run scripts for osmosis",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			node, err := cmd.Flags().GetString("node")
			if err != nil {
				return err
			}

			rpcClient, err := client.NewClientFromNode("tcp://65.108.196.173:26657")
			if err != nil {
				log.Fatal(err)
			}

			node = strings.Replace(node, "26657", "9090", 1)
			node = strings.Replace(node, "tcp://", "", 1)

			grpcClient, err := grpc.Dial(node, grpc.WithInsecure())
			if err != nil {
				return err
			}

			initClientCtx = initClientCtx.WithClient(rpcClient).
				WithGRPCClient(grpcClient)
			// WithKeyring(kr)

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			return nil
		},
	}

	rootCmd.PersistentFlags().String(flags.FlagNode, "tcp://65.108.196.173:26657", "")
	rootCmd.PersistentFlags().String(flags.FlagChainID, "osmosis-1", "")

	rootCmd.PersistentFlags().String(flags.FlagHome, defaultKeyHome, "The application home directory")
	rootCmd.PersistentFlags().String(cli.OutputFlag, "text", "Output format (text|json)")

	rootCmd.AddCommand(
		history.HistoryCmd(),
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &client.Context{})

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}

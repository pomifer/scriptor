package history

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
)

func BlockSummaryCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "summary",
		Short: "query historical info for some blocks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			start, end, err := readStartEndFlags(cmd, cctx)
			if err != nil {
				return err
			}

			valMap, err := queryValidators(cmd.Context(), cctx)
			if err != nil {
				return err
			}

			// summaries := []BlockSummary{}
			for i := start; i < end; i++ {
				res, err := cctx.Client.Block(cmd.Context(), &i)
				if err != nil {
					return err
				}

				valAddr, err := sdk.ConsAddressFromHex(res.Block.ProposerAddress.String())
				if err != nil {
					return err
				}

				valName, has := valMap[valAddr.String()]
				if !has {
					valName = valAddr.String()
				}

				bs := BlockSummary{
					Height:   res.Block.Height,
					Proposer: valName,
					Time:     res.Block.Time,
				}

				decoder := cctx.TxConfig.TxDecoder()
				for _, rawtx := range res.Block.Data.Txs {
					sdkTx, err := decoder(rawtx)
					if err != nil {
						return err
					}
					tx := ParseTx(sdkTx)
					bs.Txs = append(bs.Txs, tx)
				}

				fmt.Println(bs.String())
			}

			return nil
		},
	}

	return command
}

// returns a map of all the validators where proposer address:name
func queryValidators(ctx context.Context, cctx client.Context) (map[string]string, error) {
	sqc := stakingtypes.NewQueryClient(cctx.GRPCClient)

	valResp, err := sqc.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
		Pagination: &query.PageRequest{
			Offset:     0,
			Limit:      1000,
			CountTotal: true,
		},
	})
	if err != nil {
		return nil, err
	}

	out := make(map[string]string)

	for _, v := range valResp.Validators {
		var pubKey cryptotypes.PubKey
		err = cctx.InterfaceRegistry.UnpackAny(v.ConsensusPubkey, &pubKey)
		if err != nil {
			return nil, err
		}
		conAddr, err := sdk.ConsAddressFromHex(pubKey.Address().String())
		if err != nil {
			return nil, err
		}
		out[conAddr.String()] = v.Description.Moniker
	}
	return out, nil
}

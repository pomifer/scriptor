package history

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

type Delegatorrr struct {
	addy              string
	first_block       int64
	last_block        int64
	max_delegated     float64
	current_delegated float64
}

// func (u *Delegatorrr) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(&struct {
// 		Name string `json:"name"`
// 	}{
// 		Name: "customized" + u.addy,
// 	})
// }

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

			delegateMap := make(map[string]Delegatorrr)
			// summaries := []BlockSummary{}
			for i := start; i < end; i++ {
				// res, err := cctx.Client.Block(cmd.Context(), &i)
				// if err != nil {
				// 	return err
				// }
				fmt.Println("------------------------------------------------------------------------", i)
				res, err := cctx.Client.Block(cmd.Context(), &i)
				if err != nil {
					return err
				}

				for j, tx := range res.Block.Txs {
					//fmt.Println("###########################", j)
					j = j
					stx, err := cctx.TxConfig.TxDecoder()(tx)

					if err != nil {
						// threw this error so I'm just going to ignore it
						// unable to resolve type URL /osmosis.lockup.MsgUnlockPeriodLock: tx parse error
						fmt.Println(err)
						continue
					}

					for _, msg := range stx.GetMsgs() {
						// for new delegations
						if sdk.MsgTypeURL(msg) == "/cosmos.staking.v1beta1.MsgDelegate" {

							delegateMsg := msg.(*stakingtypes.MsgDelegate)

							if delegateMsg.ValidatorAddress == "osmovaloper1ls4kmz5v7ytwcqmmchkex970565j8q3d5s6gdw" {
								fmt.Println("found a pomifer delegation")
								var currentDelegatorrr Delegatorrr
								currentDelegatorrr = delegateMap[delegateMsg.DelegatorAddress]
								if currentDelegatorrr.addy == "" {
									currentDelegatorrr.addy = delegateMsg.DelegatorAddress
									currentDelegatorrr.first_block = i
									currentDelegatorrr.last_block = 999999999
									currentDelegatorrr.max_delegated = 0
									currentDelegatorrr.current_delegated = 0
								}

								var newAmount float64 = 0
								if delegateMsg.Amount.Denom == "uosmo" {
									newAmount = delegateMsg.Amount.Amount.ToDec().MustFloat64() / float64(1000000)
								} else {
									fmt.Println("not uosmo")
									fmt.Println(delegateMsg.Amount.Denom)
									newAmount = delegateMsg.Amount.Amount.ToDec().MustFloat64()

								}
								currentDelegatorrr.current_delegated = currentDelegatorrr.current_delegated + newAmount
								if currentDelegatorrr.max_delegated < currentDelegatorrr.current_delegated {
									currentDelegatorrr.max_delegated = currentDelegatorrr.current_delegated
								}
								delegateMap[delegateMsg.DelegatorAddress] = currentDelegatorrr
								fmt.Println("new amount is: ", newAmount, " current amount is: ", currentDelegatorrr.current_delegated, " max amount: ", currentDelegatorrr.max_delegated, " to account: ", currentDelegatorrr.addy)

							}
						}

						//for redelegations to pomifer
						if sdk.MsgTypeURL(msg) == "/cosmos.staking.v1beta1.MsgBeginRedelegate" {
							redelegateMsg := msg.(*stakingtypes.MsgBeginRedelegate)
							if redelegateMsg.ValidatorDstAddress == "osmovaloper1ls4kmz5v7ytwcqmmchkex970565j8q3d5s6gdw" {
								fmt.Println("found a pomifer redelegation")
								var currentDelegatorrr Delegatorrr
								currentDelegatorrr = delegateMap[redelegateMsg.DelegatorAddress]
								if currentDelegatorrr.addy == "" {
									currentDelegatorrr.addy = redelegateMsg.DelegatorAddress
									currentDelegatorrr.first_block = i
									currentDelegatorrr.last_block = 999999999
									currentDelegatorrr.max_delegated = 0
									currentDelegatorrr.current_delegated = 0
								}

								var newAmount float64 = 0
								if redelegateMsg.Amount.Denom == "uosmo" {
									newAmount = redelegateMsg.Amount.Amount.ToDec().MustFloat64() / float64(1000000)
								} else {
									fmt.Println("not uosmo")
									fmt.Println(redelegateMsg.Amount.Denom)
									newAmount = redelegateMsg.Amount.Amount.ToDec().MustFloat64()

								}
								currentDelegatorrr.current_delegated = currentDelegatorrr.current_delegated + newAmount
								if currentDelegatorrr.max_delegated < currentDelegatorrr.current_delegated {
									currentDelegatorrr.max_delegated = currentDelegatorrr.current_delegated
								}
								delegateMap[redelegateMsg.DelegatorAddress] = currentDelegatorrr
								fmt.Println("new amount is: ", newAmount, " current amount is: ", currentDelegatorrr.current_delegated, " max amount: ", currentDelegatorrr.max_delegated, " to account: ", currentDelegatorrr.addy)

							}
						}

						//for redelegations from pomifer
						if sdk.MsgTypeURL(msg) == "/cosmos.staking.v1beta1.MsgBeginRedelegate" {
							redelegateMsg := msg.(*stakingtypes.MsgBeginRedelegate)
							if redelegateMsg.ValidatorSrcAddress == "osmovaloper1ls4kmz5v7ytwcqmmchkex970565j8q3d5s6gdw" {
								fmt.Println("found a redelegation from pomifer")
								var currentDelegatorrr Delegatorrr
								currentDelegatorrr = delegateMap[redelegateMsg.DelegatorAddress]
								var newAmount float64 = 0
								if redelegateMsg.Amount.Denom == "uosmo" {
									newAmount = redelegateMsg.Amount.Amount.ToDec().MustFloat64() / float64(1000000)
								} else {
									fmt.Println("not uosmo")
									fmt.Println(redelegateMsg.Amount.Denom)
									newAmount = redelegateMsg.Amount.Amount.ToDec().MustFloat64()

								}
								currentDelegatorrr.current_delegated = currentDelegatorrr.current_delegated - newAmount
								if currentDelegatorrr.max_delegated < currentDelegatorrr.current_delegated {
									currentDelegatorrr.max_delegated = currentDelegatorrr.current_delegated
								}
								delegateMap[redelegateMsg.DelegatorAddress] = currentDelegatorrr
								fmt.Println("new amount is: ", newAmount, " current amount is: ", currentDelegatorrr.current_delegated, " max amount: ", currentDelegatorrr.max_delegated, " to account: ", currentDelegatorrr.addy)

							}
						}

						//for undelegations from pomifer
						if sdk.MsgTypeURL(msg) == "/cosmos.staking.v1beta1.MsgUndelegate" {
							undelegateMsg := msg.(*stakingtypes.MsgUndelegate)
							if undelegateMsg.ValidatorAddress == "osmovaloper1ls4kmz5v7ytwcqmmchkex970565j8q3d5s6gdw" {
								fmt.Println("found a pomifer undelegation")
								var currentDelegatorrr Delegatorrr
								currentDelegatorrr = delegateMap[undelegateMsg.DelegatorAddress]
								var newAmount float64 = 0
								if undelegateMsg.Amount.Denom == "uosmo" {
									newAmount = undelegateMsg.Amount.Amount.ToDec().MustFloat64() / float64(1000000)
								} else {
									fmt.Println("not uosmo")
									fmt.Println(undelegateMsg.Amount.Denom)
									newAmount = undelegateMsg.Amount.Amount.ToDec().MustFloat64()

								}
								currentDelegatorrr.current_delegated = currentDelegatorrr.current_delegated - newAmount
								if currentDelegatorrr.max_delegated < currentDelegatorrr.current_delegated {
									currentDelegatorrr.max_delegated = currentDelegatorrr.current_delegated
								}
								delegateMap[undelegateMsg.DelegatorAddress] = currentDelegatorrr
								fmt.Println("new amount is: ", newAmount, " current amount is: ", currentDelegatorrr.current_delegated, " max amount: ", currentDelegatorrr.max_delegated, " to account: ", currentDelegatorrr.addy)

							}
						}

					}
				}

			}

			// rawJson, err := json.Marshal(delegateMap)
			// if err != nil {
			// 	return err
			// } else {
			// 	fmt.Println("this is the raw json we are saving: ", string(rawJson))
			// }
			dt := time.Now()
			var fnstring = "delegateMap-" + dt.String() + ".json"
			file, err := os.OpenFile(fnstring, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			defer file.Close()

			enc := json.NewEncoder(file)
			var heady = "address,current_delegated,max_delegated,first_block,last_block"
			err = enc.Encode(heady)
			for k, e := range delegateMap {
				fmt.Println("key ", k, " element : ", e)
				var encodeme = string(e.addy) + "," + fmt.Sprintf("%f", e.current_delegated) + "," + fmt.Sprintf("%f", e.max_delegated) + "," + strconv.FormatInt(e.first_block, 10) + "," + strconv.FormatInt(e.last_block, 10)
				err = enc.Encode(encodeme)
				if err != nil {
					return err
				}
			}

			// valAddr, err := sdk.ConsAddressFromHex(res.Block.ProposerAddress.String())
			// if err != nil {
			// 	return err
			// }

			// valName, has := valMap[valAddr.String()]
			// if !has {
			// 	valName = valAddr.String()
			// }

			// bs := BlockSummary{
			// 	Height:   res.Block.Height,
			// 	Proposer: valName,
			// 	Time:     res.Block.Time,
			// }

			// decoder := cctx.TxConfig.TxDecoder()
			// for _, rawtx := range res.Block.Data.Txs {
			// 	sdkTx, err := decoder(rawtx)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	tx := ParseTx(sdkTx)
			// 	bs.Txs = append(bs.Txs, tx)
			// }

			// fmt.Println(bs.String())

			return nil
		},
	}

	return command
}

func ConsolidateAttributes(events []abcitypes.EventAttribute) string {
	out := []string{}
	for _, ev := range events {
		out = append(out, string(ev.Key))
		out = append(out, string(ev.Value))
	}
	return strings.Join(out, " ")
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

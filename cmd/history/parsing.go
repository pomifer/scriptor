package history

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gammtypes "github.com/osmosis-labs/osmosis/v9/x/gamm/types"
)

type Tx struct {
	Valid       bool
	MessageLogs []string `json:"logs"`
}

type BlockSummary struct {
	Height   int64     `json:"height"`
	Proposer string    `json:"proposer"`
	Time     time.Time `json:"time"`
	Txs      []Tx      `json:"txs"`
}

func (bs BlockSummary) String() string {
	bz, err := json.MarshalIndent(bs, "", "    ")
	if err != nil {
		fmt.Println(err)
		return "error parsing block summary json"
	}
	return string(bz)
}

func ParseTx(tx sdk.Tx) Tx {
	out := Tx{}
	for _, msg := range tx.GetMsgs() {
		l, has := ParseMsg(msg)
		if !has {
			continue
		}
		out.MessageLogs = append(out.MessageLogs, l)
	}
	return out
}

func ParseMsg(msg sdk.Msg) (string, bool) {
	switch m := msg.(type) {
	case *gammtypes.MsgSwapExactAmountIn:
		return fmt.Sprintf(
			"Exact amount in from %s, token in %s out %s minout %d",
			m.Sender,
			m.TokenIn.String(),
			m.TokenOutDenom(),
			m.TokenOutMinAmount.Int64()), true
	default:
		return "", false
	}
}

package types

import "github.com/cosmos/cosmos-sdk/types/query"

// Payload contains the payload data that is sent from Hasura
type Payload struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            PayloadArgs            `json:"input"`
}

// GetAddress returns the address associated with this payload, if any
func (p *Payload) GetAddress() string {
	return p.Input.Address
}

// GetDenom returns the denom associated with this payload, if any
func (p *Payload) GetDenom() string {
	return p.Input.Denom
}

// GetPagination returns the pagination asasociated with this payload, if any
func (p *Payload) GetPagination() *query.PageRequest {
	return &query.PageRequest{
		Offset:     p.Input.Offset,
		Limit:      p.Input.Limit,
		CountTotal: p.Input.CountTotal,
	}
}

type PayloadArgs struct {
	Address    string `json:"address"`
	Denom      string `json:"denom"`
	Height     int64  `json:"height"`
	Offset     uint64 `json:"offset"`
	Limit      uint64 `json:"limit"`
	CountTotal bool   `json:"count_total"`
}

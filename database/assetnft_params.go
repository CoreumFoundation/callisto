package database

import (
	"encoding/json"
	"fmt"

	dbtypes "github.com/forbole/callisto/v4/database/types"
	"github.com/forbole/callisto/v4/types"

	assetnfttypes "github.com/tokenize-x/tx-chain/v6/x/asset/nft/types"
)

// SaveAssetNFTParams allows to store the given params into the database.
func (db *Db) SaveAssetNFTParams(params types.AssetNFTParams) error {
	paramsBz, err := json.Marshal(&params.Params)
	if err != nil {
		return fmt.Errorf("error while marshaling assetnft params: %s", err)
	}

	stmt := `
INSERT INTO assetnft_params (params, height) 
VALUES ($1, $2)
ON CONFLICT (one_row_id) DO UPDATE 
    SET params = excluded.params,
        height = excluded.height
WHERE assetnft_params.height <= excluded.height`

	_, err = db.SQL.Exec(stmt, string(paramsBz), params.Height)
	if err != nil {
		return fmt.Errorf("error while storing assetnft params: %s", err)
	}

	return nil
}

// GetAssetNFTParams returns the types.AssetNFTParams instance containing the current params
func (db *Db) GetAssetNFTParams() (*types.AssetNFTParams, error) {
	var rows []dbtypes.AssetNFTParamsRow
	stmt := `SELECT * FROM assetnft_params LIMIT 1`
	err := db.Sqlx.Select(&rows, stmt)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no assetnft params found")
	}

	var assetnftParams assetnfttypes.Params
	err = json.Unmarshal([]byte(rows[0].Params), &assetnftParams)
	if err != nil {
		return nil, err
	}

	return &types.AssetNFTParams{
		Params: assetnftParams,
		Height: rows[0].Height,
	}, nil
}

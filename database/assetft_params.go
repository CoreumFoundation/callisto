package database

import (
	"encoding/json"
	"fmt"

	dbtypes "github.com/forbole/callisto/v4/database/types"
	"github.com/forbole/callisto/v4/types"

	assetfttypes "github.com/tokenize-x/tx-chain/v6/x/asset/ft/types"
)

// SaveAssetFTParams allows to store the given params into the database.
func (db *Db) SaveAssetFTParams(params types.AssetFTParams) error {
	paramsBz, err := json.Marshal(&params.Params)
	if err != nil {
		return fmt.Errorf("error while marshaling assetft params: %s", err)
	}

	stmt := `
INSERT INTO assetft_params (params, height) 
VALUES ($1, $2)
ON CONFLICT (one_row_id) DO UPDATE 
    SET params = excluded.params,
        height = excluded.height
WHERE assetft_params.height <= excluded.height`

	_, err = db.SQL.Exec(stmt, string(paramsBz), params.Height)
	if err != nil {
		return fmt.Errorf("error while storing assetft params: %s", err)
	}

	return nil
}

// GetAssetFTParams returns the types.AssetFTParams instance containing the current params
func (db *Db) GetAssetFTParams() (*types.AssetFTParams, error) {
	var rows []dbtypes.AssetFTParamsRow
	stmt := `SELECT * FROM assetft_params LIMIT 1`
	err := db.Sqlx.Select(&rows, stmt)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no assetft params found")
	}

	var assetftParams assetfttypes.Params
	err = json.Unmarshal([]byte(rows[0].Params), &assetftParams)
	if err != nil {
		return nil, err
	}

	return &types.AssetFTParams{
		Params: assetftParams,
		Height: rows[0].Height,
	}, nil
}

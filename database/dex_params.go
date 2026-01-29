package database

import (
	"encoding/json"
	"fmt"

	dbtypes "github.com/forbole/callisto/v4/database/types"
	"github.com/forbole/callisto/v4/types"

	dextypes "github.com/tokenize-x/tx-chain/v6/x/dex/types"
)

// SaveDEXParams allows to store the given params into the database.
func (db *Db) SaveDEXParams(params types.DEXParams) error {
	paramsBz, err := json.Marshal(&params.Params)
	if err != nil {
		return fmt.Errorf("error while marshaling dex params: %s", err)
	}

	stmt := `
INSERT INTO dex_params (params, height) 
VALUES ($1, $2)
ON CONFLICT (one_row_id) DO UPDATE 
    SET params = excluded.params,
        height = excluded.height
WHERE dex_params.height <= excluded.height`

	_, err = db.SQL.Exec(stmt, string(paramsBz), params.Height)
	if err != nil {
		return fmt.Errorf("error while storing dex params: %s", err)
	}

	return nil
}

// GetDEXParams returns the types.DEXParams instance containing the current params
func (db *Db) GetDEXParams() (*types.DEXParams, error) {
	var rows []dbtypes.DEXParamsRow
	stmt := `SELECT * FROM dex_params LIMIT 1`
	err := db.Sqlx.Select(&rows, stmt)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no dex params found")
	}

	var dexParams dextypes.Params
	err = json.Unmarshal([]byte(rows[0].Params), &dexParams)
	if err != nil {
		return nil, err
	}

	return &types.DEXParams{
		Params: dexParams,
		Height: rows[0].Height,
	}, nil
}

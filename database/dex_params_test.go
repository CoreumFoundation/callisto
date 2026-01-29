package database_test

import (
	"encoding/json"

	dbtypes "github.com/forbole/callisto/v4/database/types"
	"github.com/forbole/callisto/v4/types"

	dextypes "github.com/tokenize-x/tx-chain/v6/x/dex/types"
)

func (suite *DbTestSuite) TestSaveDEXParams() {
	dexParams := dextypes.DefaultParams()
	err := suite.database.SaveDEXParams(types.NewDEXParams(dexParams, 10))
	suite.Require().NoError(err)

	var rows []dbtypes.DEXParamsRow
	err = suite.database.Sqlx.Select(&rows, `SELECT * FROM dex_params`)
	suite.Require().NoError(err)

	suite.Require().Len(rows, 1)

	var stored dextypes.Params
	err = json.Unmarshal([]byte(rows[0].Params), &stored)
	suite.Require().NoError(err)
	suite.Require().Equal(dexParams, stored)
}

func (suite *DbTestSuite) TestGetDEXParams() {
	dexParams := dextypes.DefaultParams()

	paramsBz, err := json.Marshal(&dexParams)
	suite.Require().NoError(err)

	_, err = suite.database.SQL.Exec(
		`INSERT INTO dex_params (params, height) VALUES ($1, $2)`,
		string(paramsBz), 10,
	)
	suite.Require().NoError(err)

	params, err := suite.database.GetDEXParams()
	suite.Require().NoError(err)

	suite.Require().Equal(&types.DEXParams{
		Params: dexParams,
		Height: 10,
	}, params)
}

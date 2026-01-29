package database_test

import (
	"encoding/json"

	dbtypes "github.com/forbole/callisto/v4/database/types"
	"github.com/forbole/callisto/v4/types"

	feemodeltypes "github.com/tokenize-x/tx-chain/v6/x/feemodel/types"
)

func (suite *DbTestSuite) TestSaveFeeModelParams() {
	feemodelParams := feemodeltypes.DefaultParams()
	err := suite.database.SaveFeeModelParams(types.NewFeeModelParams(feemodelParams, 10))
	suite.Require().NoError(err)

	var rows []dbtypes.FeeModelParamsRow
	err = suite.database.Sqlx.Select(&rows, `SELECT * FROM feemodel_params`)
	suite.Require().NoError(err)

	suite.Require().Len(rows, 1)

	var stored feemodeltypes.Params
	err = json.Unmarshal([]byte(rows[0].Params), &stored)
	suite.Require().NoError(err)
	suite.Require().Equal(feemodelParams, stored)
}

func (suite *DbTestSuite) TestGetFeeModelParams() {
	feemodelParams := feemodeltypes.DefaultParams()

	paramsBz, err := json.Marshal(&feemodelParams)
	suite.Require().NoError(err)

	_, err = suite.database.SQL.Exec(
		`INSERT INTO feemodel_params (params, height) VALUES ($1, $2)`,
		string(paramsBz), 10,
	)
	suite.Require().NoError(err)

	params, err := suite.database.GetFeeModelParams()
	suite.Require().NoError(err)

	suite.Require().Equal(&types.FeeModelParams{
		Params: feemodelParams,
		Height: 10,
	}, params)
}

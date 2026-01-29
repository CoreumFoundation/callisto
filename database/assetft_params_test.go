package database_test

import (
	"encoding/json"

	dbtypes "github.com/forbole/callisto/v4/database/types"
	"github.com/forbole/callisto/v4/types"

	assetfttypes "github.com/tokenize-x/tx-chain/v6/x/asset/ft/types"
)

func (suite *DbTestSuite) TestSaveAssetFTParams() {
	assetFTParams := assetfttypes.DefaultParams()
	err := suite.database.SaveAssetFTParams(types.NewAssetFTParams(assetFTParams, 10))
	suite.Require().NoError(err)

	var rows []dbtypes.AssetFTParamsRow
	err = suite.database.Sqlx.Select(&rows, `SELECT * FROM assetft_params`)
	suite.Require().NoError(err)

	suite.Require().Len(rows, 1)

	var stored assetfttypes.Params
	err = json.Unmarshal([]byte(rows[0].Params), &stored)
	suite.Require().NoError(err)
	suite.Require().Equal(assetFTParams, stored)
}

func (suite *DbTestSuite) TestGetAssetFTParams() {
	assetFTParams := assetfttypes.DefaultParams()

	paramsBz, err := json.Marshal(&assetFTParams)
	suite.Require().NoError(err)

	_, err = suite.database.SQL.Exec(
		`INSERT INTO assetft_params (params, height) VALUES ($1, $2)`,
		string(paramsBz), 10,
	)
	suite.Require().NoError(err)

	params, err := suite.database.GetAssetFTParams()
	suite.Require().NoError(err)

	suite.Require().Equal(&types.AssetFTParams{
		Params: assetFTParams,
		Height: 10,
	}, params)
}

package database_test

import (
	"encoding/json"

	dbtypes "github.com/forbole/callisto/v4/database/types"
	"github.com/forbole/callisto/v4/types"

	assetnfttypes "github.com/tokenize-x/tx-chain/v6/x/asset/nft/types"
)

func (suite *DbTestSuite) TestSaveAssetNFTParams() {
	assetNFTParams := assetnfttypes.DefaultParams()
	err := suite.database.SaveAssetNFTParams(types.NewAssetNFTParams(assetNFTParams, 10))
	suite.Require().NoError(err)

	var rows []dbtypes.AssetNFTParamsRow
	err = suite.database.Sqlx.Select(&rows, `SELECT * FROM assetnft_params`)
	suite.Require().NoError(err)

	suite.Require().Len(rows, 1)

	var stored assetnfttypes.Params
	err = json.Unmarshal([]byte(rows[0].Params), &stored)
	suite.Require().NoError(err)
	suite.Require().Equal(assetNFTParams, stored)
}

func (suite *DbTestSuite) TestGetAssetNFTParams() {
	assetNFTParams := assetnfttypes.DefaultParams()

	paramsBz, err := json.Marshal(&assetNFTParams)
	suite.Require().NoError(err)

	_, err = suite.database.SQL.Exec(
		`INSERT INTO assetnft_params (params, height) VALUES ($1, $2)`,
		string(paramsBz), 10,
	)
	suite.Require().NoError(err)

	params, err := suite.database.GetAssetNFTParams()
	suite.Require().NoError(err)

	suite.Require().Equal(&types.AssetNFTParams{
		Params: assetNFTParams,
		Height: 10,
	}, params)
}

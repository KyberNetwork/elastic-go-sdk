package utils

import (
	"testing"

	"github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/KyberNetwork/promm-sdk-go/constants"
)

func TestComputePoolAddress(t *testing.T) {
	factoryAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	tokenA := entities.NewToken(1, common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), 18, "USDC", "USD Coin")
	tokenB := entities.NewToken(1, common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"), 18, "DAI", "Dai Stablecoin")
	result, err := ComputePoolAddress(factoryAddress, tokenA, tokenB, constants.FeeLow, "")
	if err != nil {
		panic(err)
	}
	assert.Equal(t, result, common.HexToAddress("0xE6843dD76b942866E5d14104BE14dc069b0B4D36"))

	USDC := entities.NewToken(1, common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), 18, "USDC", "USD Coin")
	DAI := entities.NewToken(1, common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"), 18, "DAI", "Dai Stablecoin")
	resultA, err := ComputePoolAddress(factoryAddress, USDC, DAI, constants.FeeLow, "")
	if err != nil {
		panic(err)
	}
	resultB, err := ComputePoolAddress(factoryAddress, DAI, USDC, constants.FeeLow, "")
	if err != nil {
		panic(err)
	}
	assert.Equal(t, resultA, resultB, "should correctly compute the pool address")
}

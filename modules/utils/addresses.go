package utils

import (
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// FilterNonAccountAddresses filters all the non-account addresses from the given slice of addresses, returning a new
// slice containing only account addresses.
func FilterNonAccountAddresses(addresses []string) []string {
	// Filter using only the account addresses as the MessageAddressesParser might return also validator addresses
	var accountAddresses []string
	for _, address := range addresses {
		_, err := sdk.AccAddressFromBech32(address)
		if err == nil {
			accountAddresses = append(accountAddresses, address)
		}
	}
	return accountAddresses
}

// GetModuleAccountAddressFromPrefix derives a module account address
// from its name using the provided bech32 address prefix and codec.
// This provides the same deterministic address derivation as Cosmos SDK
// but allows explicit control of the prefix without relying on global SDK config.
func GetModuleAccountAddressFromPrefix(moduleName, bech32Prefix string) (string, error) {
	// Derive the module account address bytes deterministically
	// moduleAddr = sha256(moduleAddrType + moduleName) where moduleAddrType = 0x0
	// This matches how authtypes.NewModuleAddress computes it internally
	moduleAddrBytes := authtypes.NewModuleAddress(moduleName).Bytes()

	// Create a bech32 codec with the provided prefix
	bech32Codec := address.NewBech32Codec(bech32Prefix)

	// Encode the address bytes using the custom prefix
	addrStr, err := bech32Codec.BytesToString(moduleAddrBytes)
	if err != nil {
		return "", err
	}

	return addrStr, nil
}

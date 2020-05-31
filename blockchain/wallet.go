package blockchain

import (
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tyler-smith/go-bip39"
)

// bitcoin blockchain network type
const (
	Mainnet = "mainnet"
	Testnet = "testnet"
	Regtest = "regtest"
	Simnet  = "simnet"
)

// HDWallet is BIP44 HD Wallet format. Apostrophe is a hardend key to enhance security
// m / purpose' / coin_type' / account' / change / address_index
type HDWallet struct {
	Mnemonic      string
	ExtendedKey   *hdkeychain.ExtendedKey
	ChainParams *chaincfg.Params
}

// Account is HDWallet account
type Account struct {
	ExtendedKey   *hdkeychain.ExtendedKey
	ChainParams *chaincfg.Params
}

// ChildWallet is created from HDWallet by index
type ChildWallet struct {
	Address   btcutil.Address
	WIF       *btcutil.WIF
	PublicKey btcec.PublicKey
}

// Purpose is BIP44 purpose
const Purpose = 44

// BIP44 CoinType
const (
	CoinTypeBitcoin        uint32 = 0
	CoinTypeBitcoinTestnet uint32 = 1
)

// HardenedKey BIP44 hardened key
const HardenedKey = 0x80000000

// BIP44 change type
const (
	ChangeTypeExternal uint32 = 0
	ChangeTypeInternal uint32 = 1
)

//
const (
	AddressP2KH   string = "p2kh"
	AddressP2SH   string = "p2sh"
	AddressBech32 string = "bech32"
)

// GetCoinType return BIP44 cointype by network
func GetCoinType(network string) uint32 {
	if network == Mainnet {
		return CoinTypeBitcoin
	}
	return CoinTypeBitcoinTestnet
}

// GetParamsFromNetwork return chain params from BlockChain Network type
func GetParamsFromNetwork(network string) (*chaincfg.Params, error) {
	switch network {
	case Mainnet:
		return &chaincfg.MainNetParams, nil
	case Testnet:
		return &chaincfg.TestNet3Params, nil
	case Regtest:
		return &chaincfg.RegressionNetParams, nil
	case Simnet:
		return &chaincfg.SimNetParams, nil
	}
	return nil, errors.New("invalid BlockChain Network")
}

// NewHDWallet return mnemonic and HDWallet ExtendedKey and network Params
func NewHDWallet(bitSize int, mnemonic string, network string, password string) (*HDWallet, error) {
	if mnemonic == "" {
		m, err := GenerateMnemonic(bitSize)
		if err != nil {
			return nil, err
		}
		mnemonic = m
	}
	seed, err := GenerateSeed(mnemonic, password)
	if err != nil {
		return nil, err
	}

	networkParams, err := GetParamsFromNetwork(network)
	if err != nil {
		return nil, err
	}

	extendedKey, err := hdkeychain.NewMaster(seed, networkParams)
	if err != nil {
		return nil, err
	}

	return &HDWallet{
		Mnemonic:      mnemonic,
		ExtendedKey:   extendedKey,
		ChainParams: networkParams,
	}, nil
}

// GenerateMnemonic return mnemonic (bitSize must be [128, 256] and a multiple of 32)
func GenerateMnemonic(bitSize int) (string, error) {
	entropy, err := bip39.NewEntropy(bitSize)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

// GenerateSeed return seed from mnemonic and password
func GenerateSeed(mnemonic, password string) ([]byte, error) {
	return bip39.NewSeedWithErrorChecking(mnemonic, password)
}

// NewAccount create Account by BIP44 settings
func (hd *HDWallet) NewAccount(purpose, coinType, account uint32) (*Account, error) {

	// add hardened
	purpose = purpose + HardenedKey
	coinType = coinType + HardenedKey
	account = account + HardenedKey

	purposeChild, err := hd.ExtendedKey.Child(purpose)
	if err != nil {
		return nil, err
	}
	coinTypeChild, err := purposeChild.Child(coinType)
	if err != nil {
		return nil, err
	}
	accountChild, err := coinTypeChild.Child(account)
	if err != nil {
		return nil, err
	}
	return &Account{
		ExtendedKey:   accountChild,
		ChainParams: hd.ChainParams,
	}, nil
}

// DeriveAddress return ChildWallet by change type and addressIndex
func (a *Account) DeriveAddress(change, addressIndex uint32, addressType string) (*ChildWallet, error) {
	changeChild, err := a.ExtendedKey.Child(change)
	if err != nil {
		return nil, err
	}

	childWallet, err := changeChild.Child(addressIndex)
	if err != nil {
		return nil, err
	}

	ecPrivKey, err := childWallet.ECPrivKey()
	if err != nil {
		return nil, err
	}

	// compressed set true
	wif, err := btcutil.NewWIF(ecPrivKey, a.ChainParams, true)
	if err != nil {
		return nil, err
	}
	ecPubkey, err := childWallet.ECPubKey()
	if err != nil {
		return nil, err
	}
	address, err := encodeAddress(*ecPubkey, addressType, *a.ChainParams)

	decodedAddress, err := btcutil.DecodeAddress(address, a.ChainParams)
	if err != nil {
		return nil, err
	}
	return &ChildWallet{
		Address:   decodedAddress,
		WIF:       wif,
		PublicKey: *ecPubkey,
	}, nil
}

// encodeAddress returns address by addressType
func encodeAddress(ecPubkey btcec.PublicKey, addressType string, networkParams chaincfg.Params) (address string, err error) {
	switch addressType {
	case AddressBech32:
		witnessHash := btcutil.Hash160(ecPubkey.SerializeCompressed())
		witnessPubKeyHash, err := btcutil.NewAddressWitnessPubKeyHash(witnessHash, &networkParams)
		if err != nil {
			return address, err
		}
		address = witnessPubKeyHash.EncodeAddress()
		return address, nil
	case AddressP2KH:
		addressPubkey, err := btcutil.NewAddressPubKey(ecPubkey.SerializeCompressed(), &networkParams)
		if err != nil {
			return address, err
		}
		address = addressPubkey.EncodeAddress()
		return address, err
	case AddressP2SH:
		keyHash := btcutil.Hash160(ecPubkey.SerializeCompressed())
		scriptSig, err := txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddData(keyHash).Script()
		if err != nil {
			return address, err
		}
		addressScript, err := btcutil.NewAddressScriptHash(scriptSig, &networkParams)
		if err != nil {
			return address, err
		}
		address = addressScript.EncodeAddress()
		return address, err
	}
	return address, errors.New("invalid address type")
}

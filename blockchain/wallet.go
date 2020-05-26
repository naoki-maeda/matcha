package blockchain

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tyler-smith/go-bip39"
)

// bitcoin blockchain network type
const (
	Mainnet = iota
	Testnet
	Regtest
	Simnet
)

// HDWallet is BIP44 HD Wallet format. Apostrophe is a hardend key to enhance security
// m / purpose' / coin_type' / account' / change / address_index
type HDWallet struct {
	mnemonic    string
	extendedKey *hdkeychain.ExtendedKey
	network     *chaincfg.Params
}

// Account is HDWallet account
type Account struct {
	extendedKey *hdkeychain.ExtendedKey
	network     *chaincfg.Params
}

// ChildWallet is created from HDWallet by index
type ChildWallet struct {
	PrivKey string
	PubKey  string
	Address string
}

// Purpose is BIP44 purpose
const Purpose = 44

// BIP44 CoinType
const (
	CoinTypeBitcoin        = 0
	CoinTypeBitcoinTestnet = 1
)

// HardenedKey BIP44 hardened key
const HardenedKey = 0x80000000

// BIP44 change type
const (
	ChangeTypeExternal uint32 = 0
	ChangeTypeInternal uint32 = 1
)

// GetParamsFromNetwork return chain params from BlockChain Network type
func GetParamsFromNetwork(networkType int) (*chaincfg.Params, error) {
	switch networkType {
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
func NewHDWallet(bitSize, networkType int, password string) (*HDWallet, error) {
	mnemonic, err := GenerateMnemonic(bitSize)
	if err != nil {
		return nil, err
	}

	seed, err := GenerateSeed(mnemonic, password)
	if err != nil {
		return nil, err
	}

	network, err := GetParamsFromNetwork(networkType)
	if err != nil {
		return nil, err
	}

	extendedKey, err := hdkeychain.NewMaster(seed, network)
	if err != nil {
		return nil, err
	}

	return &HDWallet{
		mnemonic:    mnemonic,
		extendedKey: extendedKey,
		network:     network,
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

	purposeChild, err := hd.extendedKey.Child(purpose)
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
		extendedKey: accountChild,
		network:     hd.network,
	}, nil
}

// DeriveAddress return ChildWallet by change type and addressIndex
func (a *Account) DeriveAddress(change, addressIndex uint32) (*ChildWallet, error) {
	changeChild, err := a.extendedKey.Child(change)
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

	privkey := fmt.Sprintf("%x", ecPrivKey.Serialize())

	ecPubkey, err := childWallet.ECPubKey()
	if err != nil {
		return nil, err
	}
	addressPubkey, err := btcutil.NewAddressPubKey(ecPubkey.SerializeCompressed(), a.network)
	if err != nil {
		return nil, err
	}
	pubkey := addressPubkey.String()

	keyHash := btcutil.Hash160(ecPubkey.SerializeCompressed())
	scriptSig, err := txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddData(keyHash).Script()
	if err != nil {
		return nil, err
	}
	shAddress, err := btcutil.NewAddressScriptHash(scriptSig, a.network)
	if err != nil {
		return nil, err
	}
	fmt.Println(shAddress.String())

	witnessHash := btcutil.Hash160(ecPubkey.SerializeCompressed())
	witnessPubKeyHash, err := btcutil.NewAddressWitnessPubKeyHash(witnessHash, a.network)
	if err != nil {
		return nil, err
	}
	fmt.Println(witnessPubKeyHash.String())

	address, err := childWallet.Address(a.network)
	if err != nil {
		return nil, err
	}

	return &ChildWallet{
		PrivKey: privkey,
		PubKey:  pubkey,
		Address: address.EncodeAddress(),
	}, nil
}

package fake

import (
	"context"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/forta-network/forta-json-rpc-proxy/interfaces"
	"github.com/sirupsen/logrus"
)

var ContractMeta = &bind.MetaData{
	ABI: `[
	{
		"inputs": [],
		"name": "enable",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`,
}

type EthClient struct {
	*ethclient.Client
}

func (ec *EthClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return nil
}

type fakeAttester struct {
	ethClient *EthClient
	opts      *bind.TransactOpts
	contract  *bind.BoundContract
}

func NewAttester(targetUrl, validatorAddress, attesterPrivateKey string) interfaces.Attester {
	parsed, err := ContractMeta.GetAbi()
	if err != nil {
		panic(err)
	}

	ethClient, err := ethclient.Dial(targetUrl)
	if err != nil {
		panic(err)
	}

	wrappedEthClient := &EthClient{Client: ethClient}

	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	contract := bind.NewBoundContract(common.HexToAddress(validatorAddress), *parsed, wrappedEthClient, wrappedEthClient, wrappedEthClient)

	keyBytes, err := hex.DecodeString(attesterPrivateKey)
	if err != nil {
		panic(err)
	}
	privateKey, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		panic(err)
	}

	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		panic(err)
	}

	return &fakeAttester{
		ethClient: wrappedEthClient,
		opts:      opts,
		contract:  contract,
	}
}

func (fa *fakeAttester) AttestWithTx(ctx context.Context, req *interfaces.AttestRequest) (hexutil.Bytes, error) {
	tx, err := fa.contract.Transact(fa.opts, "enable")
	if err != nil {
		panic(err)
	}
	b, err := tx.MarshalBinary()
	if err != nil {
		panic(err)
	}
	hexB := (hexutil.Bytes)(b)
	logrus.
		WithField("txHash", tx.Hash().String()).
		WithField("rawTx", hexB.String()).
		Info("successfully created the fake attestation")
	return b, nil
}

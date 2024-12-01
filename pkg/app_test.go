package pkg

import (
	"io"
	"main/assets"
	"main/pkg/fs"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // disabled
func TestAppLoadConfigError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	filesystem := &fs.TestFS{}

	app := NewApp("not-found-config.toml", filesystem, "1.2.3")
	app.Start()
}

//nolint:paralleltest // disabled
func TestAppLoadConfigInvalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	filesystem := &fs.TestFS{}

	app := NewApp("config-invalid.toml", filesystem, "1.2.3")
	app.Start()
}

//nolint:paralleltest // disabled
func TestAppFailToStart(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	filesystem := &fs.TestFS{}

	app := NewApp("config-invalid-listen-address.toml", filesystem, "1.2.3")
	app.Start()
}

//nolint:paralleltest // disabled
func TestAppStopOperation(t *testing.T) {
	filesystem := &fs.TestFS{}

	app := NewApp("config-valid.toml", filesystem, "1.2.3")
	app.Stop()
	assert.True(t, true)
}

//nolint:paralleltest // disabled
func TestAppLoadConfigOk(t *testing.T) {
	filesystem := &fs.TestFS{}

	app := NewApp("config-valid.toml", filesystem, "1.2.3")
	go app.Start()

	for {
		request, err := http.Get("http://localhost:9560/healthcheck")
		if request.Body != nil {
			_ = request.Body.Close()
		}
		if err == nil {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/staking/v1beta1/validators?pagination.count_total=true&pagination.limit=10000",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("validators.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/slashing/v1beta1/signing_infos/cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("signing-info.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/slashing/v1beta1/signing_infos/neutronvalcons1w426hkttrwrve9mj77ld67lzgx5u9m8plhmwc6",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("signing-info.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/staking/v1beta1/params",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("staking-params.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/slashing/v1beta1/params",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("slashing-params.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/slashing/v1beta1/params",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("slashing-params.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/staking/v1beta1/validators/cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e/unbonding_delegations?pagination.count_total=true&pagination.limit=1",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("unbonds.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/staking/v1beta1/validators/cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e/delegations/cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("self-delegation.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/distribution/v1beta1/validators/cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e/commission",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("commission.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/consumer_validators/neutron-1",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("consumer-validators.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/distribution/v1beta1/delegators/cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2/rewards/cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rewards.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/staking/v1beta1/validators/cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e/delegations?pagination.count_total=true&pagination.limit=1",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("delegations.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/bank/v1beta1/balances/cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("balances.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/cosmos/base/tendermint/v1beta1/node_info",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("node-info.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/base/tendermint/v1beta1/node_info",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("node-info.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/params/v1beta1/params?subspace=ccvconsumer&key=SoftOptOutThreshold",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("soft-opt-out-threshold.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.neutron.quokkastake.io/cosmos/bank/v1beta1/balances/neutron1xqz9pemz5e5zycaa89kys5aw6m8rhgsv07va3d",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("balances.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/validator_consumer_addr?chain_id=neutron-1&provider_address=cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("assigned-key.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/consumer_chains_per_validator/cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("validator-consumers.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/consumer_commission_rate/neutron-1/cosmosvalcons1rt4g447zhv6jcqwdl447y88guwm0eevnrelgzc",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("consumer-commission.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.cosmos.quokkastake.io/interchain_security/ccv/provider/consumer_chains",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("consumer-info.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.coingecko.com/api/v3/simple/price?ids=cosmos,neutron,stride,osmosis,terra-luna-2,stargaze,juno-network,sommelier,injective-protocol,cosmos,evmos,umee,comdex,neutron&vs_currencies=usd",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("coingecko.json")),
	)

	httpmock.RegisterResponder("GET", "http://localhost:9560/healthcheck", httpmock.InitialTransport.RoundTrip)
	httpmock.RegisterResponder("GET", "http://localhost:9560/metrics", httpmock.InitialTransport.RoundTrip)

	response, err := http.Get("http://localhost:9560/metrics")
	require.NoError(t, err)
	require.NotEmpty(t, response)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	err = response.Body.Close()
	require.NoError(t, err)
}

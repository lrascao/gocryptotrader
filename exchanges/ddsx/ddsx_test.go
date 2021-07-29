package ddsx

import (
	"log"
	"os"
	"testing"

	"github.com/thrasher-corp/gocryptotrader/config"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
)

// Please supply your own keys here to do authenticated endpoint testing
const (
	apiKey                  = ""
	apiSecret               = ""
	canManipulateRealOrders = false
)

var dd Ddsx

func TestMain(m *testing.M) {
	dd.SetDefaults()
	cfg := config.GetConfig()
	err := cfg.LoadConfig("../../testdata/configtest.json", true)
	if err != nil {
		log.Fatal(err)
	}

	exchCfg, err := cfg.GetExchangeConfig("Ddsx")
	if err != nil {
		log.Fatal(err)
	}

	exchCfg.API.AuthenticatedSupport = true
	exchCfg.API.AuthenticatedWebsocketSupport = true
	exchCfg.API.Credentials.Key = apiKey
	exchCfg.API.Credentials.Secret = apiSecret

	err = dd.Setup(exchCfg)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

// Ensures that this exchange package is compatible with IBotExchange
func TestInterface(t *testing.T) {
	var e exchange.IBotExchange
	if e = new(Ddsx); e == nil {
		t.Fatal("unable to allocate exchange")
	}
}

func areTestAPIKeysSet() bool {
	return dd.ValidateAPICredentials()
}

// Implement tests for API endpoints below

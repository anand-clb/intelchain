package rosetta

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"

	nodeconfig "github.com/zennittians/intelchain/internal/configs/node"
	"github.com/zennittians/intelchain/internal/utils"
	"github.com/zennittians/intelchain/itc"
	"github.com/zennittians/intelchain/rosetta/common"
	"github.com/zennittians/intelchain/rosetta/services"
)

var listener net.Listener

// StartServers starts the rosetta http server
// TODO (dm): optimize rosetta to use single flight & use extra caching type DB to avoid re-processing data
func StartServers(itc *itc.Intelchain, config nodeconfig.RosettaServerConfig, limiterEnable bool, rateLimit int) error {
	if !config.HTTPEnabled {
		utils.Logger().Info().Msg("Rosetta http server disabled...")
		return nil
	}

	network, err := common.GetNetwork(itc.ShardID)
	if err != nil {
		return err
	}
	serverAsserter, err := asserter.NewServer(
		append(common.PlainOperationTypes, common.StakingOperationTypes...),
		nodeconfig.GetShardConfig(itc.ShardID).Role() == nodeconfig.ExplorerNode,
		[]*types.NetworkIdentifier{network}, services.CallMethod, false, "",
	)
	if err != nil {
		return err
	}

	router := recoverMiddleware(server.CorsMiddleware(loggerMiddleware(getRouter(serverAsserter, itc, limiterEnable, rateLimit))))
	utils.Logger().Info().
		Int("port", config.HTTPPort).
		Str("ip", config.HTTPIp).
		Msg("Starting Rosetta server")
	endpoint := fmt.Sprintf("%s:%d", config.HTTPIp, config.HTTPPort)
	if listener, err = net.Listen("tcp", endpoint); err != nil {
		return err
	}
	go newHTTPServer(router).Serve(listener)
	fmt.Printf("Started Rosetta server at: %v\n", endpoint)
	return nil
}

// StopServers stops the rosetta http server
func StopServers() error {
	if listener == nil {
		return nil
	}
	if err := listener.Close(); err != nil {
		return err
	}
	return nil
}

func newHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:      handler,
		ReadTimeout:  common.ReadTimeout,
		WriteTimeout: common.WriteTimeout,
		IdleTimeout:  common.IdleTimeout,
	}
}

func getRouter(asserter *asserter.Asserter, itc *itc.Intelchain, limiterEnable bool, rateLimit int) http.Handler {
	return server.NewRouter(
		server.NewAccountAPIController(services.NewAccountAPI(itc), asserter),
		server.NewBlockAPIController(services.NewBlockAPI(itc), asserter),
		server.NewMempoolAPIController(services.NewMempoolAPI(itc), asserter),
		server.NewNetworkAPIController(services.NewNetworkAPI(itc), asserter),
		server.NewConstructionAPIController(services.NewConstructionAPI(itc), asserter),
		server.NewCallAPIController(
			services.NewCallAPIService(itc, limiterEnable, rateLimit,
				itc.NodeAPI.GetConfig().NodeConfig.RPCServer.EvmCallTimeout),
			asserter,
		),
		server.NewEventsAPIController(services.NewEventAPI(itc), asserter),
		server.NewSearchAPIController(services.NewSearchAPI(itc), asserter),
	)
}

func recoverMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("unknown error")
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				utils.Logger().Error().Err(err).Msg("Rosetta Error")
				// Print to stderr for quick check of rosetta activity
				debug.PrintStack()
				_, _ = fmt.Fprintf(
					os.Stderr, "%s PANIC: %s\n", time.Now().Format("2006-01-02 15:04:05"), err.Error(),
				)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func loggerMiddleware(router http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		router.ServeHTTP(w, r)
		msg := fmt.Sprintf(
			"Rosetta: %s %s %s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
		utils.Logger().Info().Msg(msg)
		// Print to stdout for quick check of rosetta activity
		fmt.Printf("%s %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
	})
}

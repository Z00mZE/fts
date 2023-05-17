package di

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/google/uuid"
	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/Z00mZE/fts/domain/entity"
	"github.com/Z00mZE/fts/gen"
	"github.com/Z00mZE/fts/internal/dictionary/russia"
	"github.com/Z00mZE/fts/internal/dispatcher"
	"github.com/Z00mZE/fts/pkg/index"
	"github.com/Z00mZE/fts/pkg/logger"
)

type closeFunc func() error

type Application struct {
	logger                  *logger.ZapLogger
	ctx                     context.Context
	listener                net.Listener
	server                  *grpc.Server
	apiGatewayListener      net.Listener
	apiGatewayRuntimeServer *runtime.ServeMux
	apiServer               *http.Server
	closeFunc               []closeFunc
	index                   *index.Index
	dictionary              *russia.Dictionary
}

func Run() {
	app := new(Application)
	if initError := app.init(); initError != nil {
		app.Log().Error("init error", zap.Error(initError))
		return
	}
	defer app.close()

	app.start()
	app.Log().Info("application shutdown success")
}

func (a *Application) init() error {
	initSteps := []func() error{
		a.initContext,
		a.initLogger,
		a.initListener,
		a.initServer,
		a.initApiGatewayListener,
		a.initApiGatewayRuntimeServer,
		a.initApiGatewayServer,
		a.initGateway,
		a.initDictionary,
		a.initIndex,
		a.initDispatcher,
	}
	for _, initFunc := range initSteps {
		if hasError := initFunc(); hasError != nil {
			return hasError
		}
	}
	a.Log().Info("end initialize")
	return nil
}

func (a *Application) initLogger() error {
	a.logger = logger.NewZapLogger()
	return nil
}
func (a *Application) Log() *logger.ZapLogger {
	return a.logger
}

func (a *Application) initContext() error {
	a.ctx = context.Background()
	return nil
}
func (a *Application) Context() context.Context {
	return a.ctx
}

func (a *Application) initListener() error {
	a.Log().Info("start initialize gRPC Listener")
	var listenConfig net.ListenConfig

	listener, listenerError := listenConfig.Listen(a.Context(), "tcp", net.JoinHostPort("", "8087"))
	if listenerError != nil {
		return errors.Wrap(listenerError, "an error occurred while creating the listener")
	}
	a.listener = listener
	a.Log().Info("end initialize gRPC Listener")
	return nil
}
func (a *Application) Listener() net.Listener {
	return a.listener
}

func (a *Application) initServer() error {
	a.Log().Info("start initialize gRPC-server")

	recoverOptionHandler := recovery.WithRecoveryHandler(func(p any) (err error) {
		a.Log().Zap().Panic("panic", zap.Any("error", p))
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	})

	a.server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(recoverOptionHandler),
			grpcZap.UnaryServerInterceptor(a.Log().Zap()),
		),
		grpc.ChainStreamInterceptor(
			recovery.StreamServerInterceptor(recoverOptionHandler),
			grpcZap.StreamServerInterceptor(a.Log().Zap()),
		),
	)

	a.Log().Info("end initialize gRPC-server")
	return nil
}
func (a *Application) Server() *grpc.Server {
	return a.server
}

func (a *Application) initApiGatewayListener() error {
	a.Log().Info("start initialize ApiGateway Listener")
	var listenConfig net.ListenConfig
	listener, listenerError := listenConfig.Listen(a.ctx, "tcp", net.JoinHostPort("", "9087"))
	if listenerError != nil {
		return errors.Wrap(listenerError, "an error occurred while creating the ApiGateway Listener")
	}
	a.apiGatewayListener = listener
	a.Log().Info("end initialize ApiGateway Listener")
	return nil
}
func (a *Application) ApiGatewayListener() net.Listener {
	return a.apiGatewayListener
}

func (a *Application) initApiGatewayRuntimeServer() error {
	a.apiGatewayRuntimeServer = runtime.NewServeMux()
	return nil
}
func (a *Application) ApiGatewayRuntimeServer() *runtime.ServeMux {
	return a.apiGatewayRuntimeServer
}

func (a *Application) initApiGatewayServer() error {
	a.apiServer = &http.Server{
		Handler: a.ApiGatewayRuntimeServer(),
		BaseContext: func(listener net.Listener) context.Context {
			return a.Context()
		},
	}
	return nil
}
func (a *Application) ApiGatewayServer() *http.Server {
	return a.apiServer
}

func (a *Application) initGateway() error {
	//return gen.RegisterAdEngineHandlerFromEndpoint(
	return gen.RegisterSnowballHandlerFromEndpoint(
		a.ctx,
		a.ApiGatewayRuntimeServer(),
		a.Listener().Addr().String(),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
}

func (a *Application) initDispatcher() error {
	a.Log().Info("start initialize Dispatcher")
	gen.RegisterSnowballServer(
		a.Server(),
		dispatcher.NewDispatcher(a.Index(), a.Log()),
	)
	a.Log().Info("end initialize Dispatcher")
	return nil
}

func (a *Application) close() {
	for _, closeFuncDestination := range a.closeFunc {
		if hasError := closeFuncDestination(); hasError != nil {
			a.Log().Error("application close", zap.Error(hasError))
		}
	}
}

func (a *Application) start() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	serverErrorCh := make(chan error)

	go func() {
		a.Log().Info("gRPC-server start", zap.String("addr", a.Listener().Addr().String()))
		reflection.Register(a.Server())
		if hasError := a.Server().Serve(a.Listener()); hasError != nil {
			serverErrorCh <- errors.Wrap(hasError, "error occurred on gRPC serve")
		}
	}()
	go func() {
		a.Log().Info("ApiGateway start on ", zap.String("addr", a.ApiGatewayListener().Addr().String()))
		if hasError := http.Serve(a.ApiGatewayListener(), a.ApiGatewayRuntimeServer()); hasError != nil {
			serverErrorCh <- errors.Wrap(hasError, "error occurred on ApiGateway serve")
		}
	}()

loop:
	for {
		select {
		case <-a.ctx.Done():
			a.Log().Info("parent context was closed")
			break loop
		case serverError := <-serverErrorCh:
			a.Log().Error("shutting down the server", zap.Error(serverError))
			break loop
		case <-quit:
			a.Log().Info("received a signal to stop server")
			break loop
		}
	}

	a.Log().Info("start graceful shutdown")
	{
		a.Log().Info("shutdown ApiGateway listener")
		if err := a.ApiGatewayListener().Close(); err != nil {
			a.Log().Error("an error occurred while shutting down ApiGateway listener", zap.Error(err))
		}
		a.Log().Info("ApiGateway Listener closed success")
		a.Log().Info("shutdown ApiGateway Server")
		if apiGatewayShutdownError := a.ApiGatewayServer().Shutdown(a.ctx); apiGatewayShutdownError != nil {
			a.Log().Error("occurred error on shutting down ApiGateway server", zap.Error(apiGatewayShutdownError))
		}
	}
	a.Server().GracefulStop()
	a.Log().Info("shutdown gRPC Server")
}

func (a *Application) initDictionary() error {
	a.dictionary = russia.NewDictionary()
	return nil
}

func (a *Application) initIndex() error {
	a.index = index.NewIndex()
	total := 150_000
	for i := 0; i < total; i++ {
		if i%5_000 == 0 {
			a.Log().Info(fmt.Sprintf("saturation document index %d/%d", i, total))
		}
		a.index.Add(
			entity.Document{
				ID:   uuid.NewString(),
				Text: strings.Join(a.dictionary.Rand(rand.Intn(32)), " "),
			},
		)
	}
	return nil
}
func (a *Application) Index() *index.Index {
	return a.index
}

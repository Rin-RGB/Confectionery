package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	_ "SPOproject/docs"
	coreauth "SPOproject/internal/core/auth"
	"SPOproject/internal/core/config"
	corehash "SPOproject/internal/core/hash"
	coremiddleware "SPOproject/internal/core/http/middleware"
	coreserver "SPOproject/internal/core/http/server"
	corelogger "SPOproject/internal/core/logger"
	"SPOproject/internal/handlers"
	handlerauth "SPOproject/internal/handlers/auth"
	handlerfillings "SPOproject/internal/handlers/fillings"
	handlerorders "SPOproject/internal/handlers/orders"
	postgres "SPOproject/internal/repository/postgres"
	repositoryfillings "SPOproject/internal/repository/postgres/fillings"
	repositoryorders "SPOproject/internal/repository/postgres/orders"
	repositoryusers "SPOproject/internal/repository/postgres/users"
	serviceauth "SPOproject/internal/service/auth"
	servicefillings "SPOproject/internal/service/fillings"
	serviceorders "SPOproject/internal/service/orders"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// @title Confectionery API
// @version 1.0
// @description API для авторизации, управления начинками и заказами кондитерской.
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Access-токен в формате: Bearer {token}
// @tag.name auth
// @tag.description Регистрация, вход и управление JWT-токенами.
// @tag.name fillings
// @tag.description Получение и управление начинками.
// @tag.name orders
// @tag.description Расчёт стоимости и управление заказами.
// @tag.name static
// @tag.description Получение статических изображений.
func main() {
	cfg := config.NewConfigMust()

	logger, err := corelogger.NewLogger(*cfg.LoggerCfg)
	if err != nil {
		log.Fatal("failed to create logger: ", err)
	}
	defer func() {
		if closeErr := logger.Close(); closeErr != nil {
			log.Printf("failed to close logger: %v", closeErr)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Debug("creating postgres pool")
	pool, err := postgres.NewPool(ctx, cfg.DbConfig)
	if err != nil {
		logger.Error("failed to create postgres pool", zap.Error(err))
		return
	}
	defer pool.Close()

	passwordHasher := corehash.NewHasher(cfg.AuthCfg.BcryptCost)
	tokenProvider := coreauth.NewJWTProvider(
		cfg.AuthCfg.SigningKey,
		cfg.AuthCfg.AccessTokenTTL,
		cfg.AuthCfg.RefreshTokenTTL,
	)

	txManager := postgres.NewTxManager(pool, cfg.DbConfig.Timeout)
	usersRepository := repositoryusers.NewRepository(txManager)
	fillingsRepository := repositoryfillings.NewRepository(txManager)
	ordersRepository := repositoryorders.NewRepository(txManager)

	authService := serviceauth.NewService(usersRepository, passwordHasher, tokenProvider, txManager)
	fillingsService := servicefillings.NewService(fillingsRepository)
	ordersService := serviceorders.NewService(ordersRepository, fillingsRepository, txManager)
	requestValidator := validator.New()

	authHandler := handlerauth.NewHandler(authService, requestValidator)
	fillingsHandler := handlerfillings.NewHandler(fillingsService, tokenProvider, requestValidator)
	ordersHandler := handlerorders.NewHandler(ordersService, tokenProvider, requestValidator)

	routerV1 := coreserver.NewAPIVersionRouter(coreserver.Version1)
	routerV1.ChainRoutes(handlers.GetAllRoutes(
		authHandler,
		fillingsHandler,
		ordersHandler,
	)...)

	httpServer := coreserver.New(
		*cfg.ServerCfg,
		[]*coreserver.Router{routerV1},
		coremiddleware.RequestID(),
		coremiddleware.Logger(logger),
		coremiddleware.Trace(),
		coremiddleware.PanicRecoverer(),
	)

	serverError := make(chan error, 1)
	go func() {
		logger.Info("http server started", zap.String("address", httpServer.Addr))
		serverError <- httpServer.ListenAndServe()
	}()

	select {
	case serveErr := <-serverError:
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Error("http server stopped unexpectedly", zap.Error(serveErr))
		}
		return
	case <-ctx.Done():
	}

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		cfg.ServerCfg.ShutdownTimeout,
	)
	defer cancel()

	logger.Info("shutting down http server")
	if err = httpServer.Shutdown(shutdownContext); err != nil {
		logger.Error("failed to shut down http server", zap.Error(err))
	}
}

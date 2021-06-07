package infra

import (
	"context"

	"github.com/bool64/brick"
	"github.com/bool64/brick-template/internal/domain/greeting"
	"github.com/bool64/brick-template/internal/infra/schema"
	"github.com/bool64/brick-template/internal/infra/service"
	"github.com/bool64/brick-template/internal/infra/storage"
	"github.com/bool64/brick/database"
	"github.com/bool64/brick/jaeger"
	"github.com/go-sql-driver/mysql"
	"github.com/swaggest/rest/response/gzip"
)

// NewServiceLocator creates application service locator.
func NewServiceLocator(cfg service.Config) (loc *service.Locator, err error) {
	l := &service.Locator{}

	defer func() {
		if err != nil && l != nil && l.LoggerProvider != nil {
			l.CtxdLogger().Error(context.Background(), err.Error())
		}
	}()

	l.BaseLocator, err = brick.NewBaseLocator(cfg.BaseConfig)
	if err != nil {
		return nil, err
	}

	if err = jaeger.Setup(cfg.Jaeger, cfg.ServiceName, l.BaseLocator); err != nil {
		return nil, err
	}

	schema.SetupOpenapiCollector(l.OpenAPI)

	l.HTTPServerMiddlewares = append(l.HTTPServerMiddlewares, gzip.Middleware)

	if err = setupStorage(l, cfg.Database); err != nil {
		return nil, err
	}

	l.GreetingMakerProvider = &storage.GreetingSaver{
		Upstream: &greeting.SimpleMaker{},
		Storage:  l.Storage,
	}

	return l, nil
}

func setupStorage(l *service.Locator, cfg database.Config) error {
	c, err := mysql.ParseDSN(cfg.DSN)
	if err != nil {
		return err
	}

	conn, err := mysql.NewConnector(c)
	if err != nil {
		return err
	}

	l.Storage, err = database.SetupStorage(cfg, l.CtxdLogger(), "mysql", conn, storage.Migrations)
	if err != nil {
		return err
	}

	return nil
}

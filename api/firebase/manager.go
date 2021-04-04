package firebase

import (
	"context"
	"errors"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/4538cgy/backend-second/config"
	"google.golang.org/api/option"
)

type Firebase interface {
	App() *firebase.App
}

type manager struct {
	conf *config.Config
	app  *firebase.App
}

func NewManager(conf *config.Config) (Firebase, error) {
	m := &manager{
		conf: conf,
	}
	opt := option.WithCredentialsFile(conf.Firebase.ServiceAccountKeyPath)
	var err error
	m.app, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("firebase NewApp failed. err: %s", err))
	}

	return m, nil
}

func (m *manager) App() *firebase.App {
	return m.app
}

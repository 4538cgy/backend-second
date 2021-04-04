package firebase

import (
	"context"
	"errors"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"fmt"
	"github.com/4538cgy/backend-second/config"
	"google.golang.org/api/option"
)

type Firebase interface {
	CreateCustomToken(uniqueId string) (string, error)
	VerifyIDToken(idToken string) (string, error)
	GetUserEmail(idToken string) (string, string, error)
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

func (m *manager) CreateCustomToken(uniqueId string) (string, error) {
	client, err := m.app.Auth(context.Background())
	if err != nil {
		return "", err
	}

	ctx2 := context.Background()
	token, err := client.CustomToken(ctx2, uniqueId)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (m *manager) getAuthToken(idToken string) (*auth.Client, *auth.Token, error) {
	client, err := m.app.Auth(context.Background())
	if err != nil {
		return nil, nil, err
	}

	token, err := client.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return nil, nil, err
	}
	return client, token, nil
}

func (m *manager) VerifyIDToken(idToken string) (string, error) {
	_, token, err := m.getAuthToken(idToken)
	if err != nil {
		return "", err
	}
	return token.UID, nil
}

func (m *manager) GetUserEmail(idToken string) (string, string, error) {
	client, token, err := m.getAuthToken(idToken)
	if err != nil {
		return "", "", err
	}

	user, err := client.GetUser(context.Background(), token.UID)
	if err != nil {
		return "", "", err
	}

	return token.UID, user.Email, nil
}

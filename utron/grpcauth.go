package utron

import (
	"context"
	"encoding/base64"
)

type basicAuth struct {
	username string
	password string
}

func (b basicAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	auth := b.username + ":" + b.password
	enc := base64.StdEncoding.EncodeToString([]byte(auth))
	return map[string]string{"authorization": "Basic " + enc}, nil
}

func (basicAuth) RequireTransportSecurity() bool {
	return false
}

type auth struct {
	token string
}

func (a auth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"x-token": a.token,
	}, nil
}

func (auth) RequireTransportSecurity() bool {
	return false
}

package avanza

import (
	"github.com/vmorsell/avanza/internal/auth"
	"github.com/vmorsell/avanza/internal/client"
)

type Avanza struct {
	client *client.Client
	Auth   *auth.AuthService
}

func New() *Avanza {
	client := client.NewClient()

	return &Avanza{
		client: client,
		Auth:   auth.NewAuthService(client),
	}
}

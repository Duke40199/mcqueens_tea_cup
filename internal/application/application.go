package application

import "McQueens_Tea_Cup/internal/infra/config"

type Application struct {
}

func NewApplication() (*Application, error) {
	_, err := config.LoadConfig("feed_state.json")
	if err != nil {
		return nil, err
	}
	return &Application{}, nil
}

package reminder

import (
	"context"

	"github.com/skwair/harmony"
	"github.com/stretchr/testify/mock"
	"github.com/youkoulayley/pet-reminder-bot/pkg/store"
)

type storerMock struct {
	mock.Mock
}

func (s *storerMock) GetPet(_ context.Context, name string) (store.Pet, error) {
	ret := s.Called(name)

	return ret.Get(0).(store.Pet), ret.Error(1)
}

func (s *storerMock) ListReminds(_ context.Context) ([]store.Remind, error) {
	ret := s.Called()

	return ret.Get(0).([]store.Remind), ret.Error(1)
}

func (s *storerMock) UpdateRemind(_ context.Context, remind store.Remind) error {
	return s.Called(remind).Error(0)
}

type discordMock struct {
	mock.Mock
}

func (d *discordMock) SendMessage(_ context.Context, text string) (*harmony.Message, error) {
	ret := d.Called(text)

	return ret.Get(0).(*harmony.Message), ret.Error(1)
}

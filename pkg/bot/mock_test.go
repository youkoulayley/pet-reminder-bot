package bot

import (
	"context"

	"github.com/skwair/harmony/discord"
	"github.com/stretchr/testify/mock"
	"github.com/youkoulayley/pet-reminder-bot/pkg/store"
)

type discordMock struct {
	mock.Mock
}

func (d *discordMock) Message(_ context.Context, id string) (*discord.Message, error) {
	ret := d.Called(id)

	return ret.Get(0).(*discord.Message), ret.Error(1)
}

func (d *discordMock) SendMessage(_ context.Context, text string) (*discord.Message, error) {
	ret := d.Called(text)

	return ret.Get(0).(*discord.Message), ret.Error(1)
}

type storeMock struct {
	mock.Mock
}

func (s *storeMock) ListRemindsByID(_ context.Context, id string) ([]store.Remind, error) {
	ret := s.Called(id)

	return ret.Get(0).([]store.Remind), ret.Error(1)
}

func (s *storeMock) ListPets(_ context.Context) (store.Pets, error) {
	ret := s.Called()

	return ret.Get(0).(store.Pets), ret.Error(1)
}

func (s *storeMock) GetPet(_ context.Context, name string) (store.Pet, error) {
	ret := s.Called(name)

	return ret.Get(0).(store.Pet), ret.Error(1)
}

func (s *storeMock) CreateRemind(_ context.Context, remind store.Remind) error {
	return s.Called(remind).Error(0)
}

func (s *storeMock) UpdateRemind(_ context.Context, remind store.Remind) error {
	return s.Called(remind).Error(0)
}

func (s *storeMock) GetRemind(_ context.Context, id string) (store.Remind, error) {
	ret := s.Called(id)

	return ret.Get(0).(store.Remind), ret.Error(1)
}

func (s *storeMock) RemoveRemind(_ context.Context, id string) error {
	return s.Called(id).Error(0)
}

type reminderMock struct {
	mock.Mock
}

func (r *reminderMock) SetUpdate() {
	r.Called()
}

package handlers

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/youkoulayley/pet-reminder-bot/pkg/bot"
)

type botMock struct {
	mock.Mock
}

func (b *botMock) ListReminds(_ context.Context, id string) {
	b.Called(id)
}

func (b *botMock) ListPets(_ context.Context) {
	b.Called()
}

func (b *botMock) Remind(_ context.Context, cfg bot.RemindConfig) {
	b.Called(cfg)
}

func (b *botMock) RemoveRemind(_ context.Context, cfg bot.RemoveRemindConfig) {
	b.Called(cfg)
}

func (b *botMock) Help(_ context.Context) {
	b.Called()
}

func (b *botMock) NewCycle(_ context.Context, cfg bot.NewCycleConfig) {
	b.Called(cfg)
}

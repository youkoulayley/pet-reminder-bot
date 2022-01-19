package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testLocation = "Europe/Paris"

func setupHandler(t *testing.T, h Handler) Handler {
	t.Helper()

	tz, err := time.LoadLocation(testLocation)
	require.NoError(t, err)

	h.timezone = tz

	return h
}

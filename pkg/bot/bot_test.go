package bot

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testLocation = "Europe/Paris"

func setupBot(t *testing.T, b Bot) Bot {
	t.Helper()

	tz, err := time.LoadLocation(testLocation)
	require.NoError(t, err)

	b.timezone = tz

	return b
}

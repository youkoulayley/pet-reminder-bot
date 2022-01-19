package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_ListPets(t *testing.T) {
	s := createStore(t, nil)

	got, err := s.ListPets(context.Background())
	require.NoError(t, err)

	// Fill IDs
	p := pets()
	for i := range p {
		p[i].ID = got[i].ID
	}

	assert.Equal(t, p, got)
}

func TestStore_GetPet(t *testing.T) {
	s := createStore(t, nil)

	got, err := s.GetPet(context.Background(), "Chacha")
	require.NoError(t, err)

	want := Pet{
		ID:              got.ID,
		Name:            "Chacha",
		FoodMinDuration: 5 * time.Hour,
		FoodMaxDuration: 18 * time.Hour,
		StatsMax:        map[string]int{"intelligence": 80, "pourcentage_resistance_neutre": 20, "agilite": 80, "vitalite": 80, "force": 80},
	}
	assert.Equal(t, want, got)
}

func TestStore_GetPet_notFoundError(t *testing.T) {
	s := createStore(t, nil)

	_, err := s.GetPet(context.Background(), "Unknown")
	require.ErrorAs(t, err, &NotFoundError{})
}

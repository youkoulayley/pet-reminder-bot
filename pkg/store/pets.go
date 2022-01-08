package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Pet represents a pet.
type Pet struct {
	ID              primitive.ObjectID `bson:"_id"`
	Name            string             `bson:"name"`
	Image           string             `bson:"image"`
	FoodMinDuration time.Duration      `bson:"foodMinDuration"`
	FoodMaxDuration time.Duration      `bson:"foodMaxDuration"`
	StatsMax        map[string]int     `bson:"statsMax"`
}

// Pets represents a list of pet.
type Pets []Pet

// String returns all the pets name.
func (p Pets) String() string {
	var str string

	for _, pet := range p {
		str += pet.Name + "\n"
	}

	return str
}

// ListPets lists all pets.
func (s *Store) ListPets(ctx context.Context) (Pets, error) {
	req, err := s.pets.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}

	var pets Pets
	if err = req.All(ctx, &pets); err != nil {
		return nil, err
	}

	return pets, nil
}

// GetPet returns a pet by the given name.
func (s *Store) GetPet(ctx context.Context, name string) (Pet, error) {
	filter := bson.D{{Key: "name", Value: name}}

	var pet Pet
	if err := s.pets.FindOne(ctx, filter).Decode(&pet); err != nil {
		return Pet{}, fmt.Errorf("find: %w", err)
	}

	return pet, nil
}

package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	databaseName     = "reminderbot"
	petCollection    = "pets"
	remindCollection = "reminds"
)

// Store represents the store.
type Store struct {
	client  *mongo.Client
	pets    *mongo.Collection
	reminds *mongo.Collection
}

// New creates a new Store.
func New(client *mongo.Client) (Store, error) {
	return Store{
		client:  client,
		pets:    client.Database(databaseName).Collection(petCollection),
		reminds: client.Database(databaseName).Collection(remindCollection),
	}, nil
}

// Bootstrap boostraps the database.
func (s *Store) Bootstrap(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "name", Value: 1},
			},
			Options: options.Index().
				SetName("_uniq_name").
				SetUnique(true),
		},
	}

	if _, err := s.pets.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("create workspace indexes: %w", err)
	}

	if err := s.initData(ctx); err != nil {
		return fmt.Errorf("init data: %w", err)
	}

	return nil
}

func (s *Store) initData(ctx context.Context) error {
	pets := []Pet{
		{
			Name:            "Chacha",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 18 * time.Hour,
			StatsMax:        map[string]int{"intelligence": 80, "pourcentage_resistance_neutre": 20, "agilite": 80, "vitalite": 80, "force": 80},
		},
		{
			Name:            "Bwak_Air",
			FoodMinDuration: 11 * time.Hour,
			FoodMaxDuration: 36 * time.Hour,
			StatsMax:        map[string]int{"agilite": 80, "pourcentage_resistance_neutre": 20, "vitalite": 80},
		},
		{
			Name:            "Bwak_Terre",
			FoodMinDuration: 11 * time.Hour,
			FoodMaxDuration: 36 * time.Hour,
			StatsMax:        map[string]int{"force": 80, "pourcentage_resistance_neutre": 20, "vitalite": 80},
		},
		{
			Name:            "Bwak_Feu",
			FoodMinDuration: 11 * time.Hour,
			FoodMaxDuration: 36 * time.Hour,
			StatsMax:        map[string]int{"intelligence": 80, "pourcentage_resistance_neutre": 20, "vitalite": 80},
		},
		{
			Name:            "Bwak_Eau",
			FoodMinDuration: 11 * time.Hour,
			FoodMaxDuration: 36 * time.Hour,
			StatsMax:        map[string]int{"chance": 80, "pourcentage_resistance_neutre": 20, "vitalite": 80},
		},
		{
			Name:            "Bworky",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"pods": 1000},
		},
		{
			Name:            "Chienchien_Noir",
			FoodMinDuration: 11 * time.Hour,
			FoodMaxDuration: 36 * time.Hour,
			StatsMax:        map[string]int{"pourcentage_dommage": 40},
		},
		{
			Name:            "Koalak_Sanguin",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"sagesse": 50},
		},
		{
			Name:            "Nomoon",
			FoodMinDuration: 24 * time.Hour,
			FoodMaxDuration: 48 * time.Hour,
			StatsMax:        map[string]int{"prospection": 80},
		},
		{
			Name:            "Petit_Chacha_Blanc",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 36 * time.Hour,
			StatsMax:        map[string]int{"initiative": 500},
		},
		{
			Name:            "Peki",
			FoodMinDuration: 3 * time.Hour,
			FoodMaxDuration: 36 * time.Hour,
			StatsMax:        map[string]int{"vitalite": 300},
		},
		{
			Name:            "Vilain_Petit_Corbac",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 48 * time.Hour,
			StatsMax:        map[string]int{"prospection": 40},
		},
		{
			Name:            "Atouin",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"dommage": 10, "soin": 10},
		},
		{
			Name:            "Wabbit",
			FoodMinDuration: 24 * time.Hour,
			FoodMaxDuration: 48 * time.Hour,
			StatsMax:        map[string]int{"force": 80, "agilite": 80, "chance": 80, "sagesse": 27},
		},
		{
			Name:            "Fotome",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"vitalite": 150},
		},
		{
			Name:            "Croum",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"pourcentage_resistance_neutre": 20, "pourcentage_resistance_terre": 20, "pourcentage_resistance_eau": 20, "pourcentage_resistance_air": 20, "pourcentage_resistance_feu": 20},
		},
		{
			Name:            "Dragoune_Rose",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"sagesse": 50},
		},
		{
			Name:            "Willy_le_Relou",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"pourcentage_dommage": 50},
		},
		{
			Name:            "Bebe_Pandawa",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"pourcentage_dommage": 50},
		},
		{
			Name:            "Feanor",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"pourcentage_dommage": 50},
		},
		{
			Name:            "Walk",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"pourcentage_dommage": 50},
		},
		{
			Name:            "Mini_Wa",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"force": 80, "agilite": 80, "chance": 80, "intelligence": 80},
		},
		{
			Name:            "Ross",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"soin": 6},
		},
		{
			Name:            "Bilby",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"dommage": 10},
		},
		{
			Name:            "Ecureuil_Chenapan",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"pourcentage_dommage": 50},
		},
		{
			Name:            "Leopardo",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"pourcentage_dommage": 50},
		},
		{
			Name:            "Chacha_Tigre",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"pourcentage_dommage": 50},
		},
		{
			Name:            "Chacha_Angora",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"pourcentage_dommage": 50},
		},
		{
			Name:            "Pioute_Bleu",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"chance": 80},
		},
		{
			Name:            "Pioute_Jaune",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"agilite": 80},
		},
		{
			Name:            "Pioute_Rouge",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"intelligence": 80},
		},
		{
			Name:            "Pioute_Verte",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"force": 80},
		},
		{
			Name:            "Pioute_Rose",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"soin": 10},
		},
		{
			Name:            "Pioute_Violet",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"dommage": 10},
		},
		{
			Name:            "Crocodaille",
			FoodMinDuration: 5 * time.Hour,
			FoodMaxDuration: 72 * time.Hour,
			StatsMax:        map[string]int{"pourcentage_dommage": 50},
		},
	}

	for _, pet := range pets {
		pet.ID = primitive.NewObjectID()
		_, err := s.pets.InsertOne(ctx, pet)
		if err != nil {
			if isMongoDBDuplicateError(err) {
				continue
			}

			return fmt.Errorf("insert document: %w", err)
		}
	}

	return nil
}

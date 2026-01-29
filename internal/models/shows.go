package models

import (
	"context"
	"log"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Show struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Title        string             `bson:"title"`
	TitleEnglish string             `bson:"titleEnglish"`
	ImdbID       string             `bson:"imdb"`
	TmdbID       string             `bson:"tmdb"`
	Year         string             `bson:"year"`
	Image        string             `bson:"image"`
	VideoURL     string             `bson:"videoUrl"`
}

type Episode struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	ShowImdb string             `bson:"showImdb"`
	Season   int                `bson:"season"`
	Episode  int                `bson:"episode"`
	Title    string             `bson:"title"`
	Image    string             `bson:"image"`
}

type ShowImage struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	Title        string             `bson:"title"`
	TitleEnglish string             `bson:"titleEnglish"`
	Image        string             `bson:"image" json:"image"`
}

var (
	showCollection    *mongo.Collection
	episodeCollection *mongo.Collection
)

func InitShowMongo(uri, dbName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	db := client.Database(dbName)
	showCollection = db.Collection("shows")
	episodeCollection = db.Collection("episodes")

	log.Println("MongoDB connected (shows, episodes)")

	return ensureShowIndexes(ctx)
}

func ensureShowIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "imdb", Value: 1}},
			Options: options.Index().
				SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "title", Value: "text"},
				{Key: "titleEnglish", Value: "text"},
			},
		},
	}

	_, err := showCollection.Indexes().CreateMany(ctx, indexes)
	return err
}

/* -------------------- SHOWS -------------------- */

func InsertShows(shows []Show) error {
	if showCollection == nil {
		return mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	docs := make([]any, 0, len(shows))
	for _, s := range shows {
		docs = append(docs, s)
	}

	_, err := showCollection.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
	return err
}

func GetAllShows() ([]Show, error) {
	return findShows(bson.M{})
}

func GetShowByID(id string) (*Show, error) {
	if showCollection == nil {
		return nil, mongo.ErrClientDisconnected
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var show Show
	err = showCollection.FindOne(ctx, bson.M{"_id": oid}).Decode(&show)
	return &show, err
}

func GetAllShowImages() ([]ShowImage, error) {
	if showCollection == nil {
		return nil, mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetProjection(bson.M{
		"title":        1,
		"titleEnglish": 1,
		"image":        1,
	})

	cursor, err := showCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var images []ShowImage
	for cursor.Next(ctx) {
		var img ShowImage
		if err := cursor.Decode(&img); err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, nil
}

func SearchShowsByTitle(query string) ([]Show, error) {
	safe := regexp.QuoteMeta(query)

	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": safe, "$options": "i"}},
			{"titleEnglish": bson.M{"$regex": safe, "$options": "i"}},
		},
	}

	return findShows(filter)
}

func HasShows() (bool, error) {
	if showCollection == nil {
		return false, mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := showCollection.CountDocuments(ctx, bson.D{})
	return count > 0, err
}

func ClearShowsCollection() error {
	if showCollection == nil {
		return mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := showCollection.DeleteMany(ctx, bson.M{})
	return err
}

func findShows(filter bson.M) ([]Show, error) {
	if showCollection == nil {
		return nil, mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := showCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var shows []Show
	for cursor.Next(ctx) {
		var s Show
		if err := cursor.Decode(&s); err != nil {
			return nil, err
		}
		shows = append(shows, s)
	}

	return shows, nil
}

func GetShowByTmdbID(tmdbID string) (*Show, error) {
	if showCollection == nil {
		return nil, mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var show Show
	err := showCollection.FindOne(ctx, bson.M{"tmdb": tmdbID}).Decode(&show)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &show, nil
}

func InsertShow(show Show) error {
	if showCollection == nil {
		return mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := showCollection.InsertOne(ctx, show)
	return err
}

func GetAllShowImdbIds() ([]string, error) {
	if showCollection == nil {
		return nil, mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := showCollection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"imdb": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ids []string
	for cursor.Next(ctx) {
		var s struct {
			ImdbID string `bson:"imdb"`
		}
		if err := cursor.Decode(&s); err != nil {
			return nil, err
		}
		ids = append(ids, s.ImdbID)
	}

	return ids, nil
}

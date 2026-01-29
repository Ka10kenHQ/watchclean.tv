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


type Movie struct {
    Title        string `bson:"title"`
    TitleEnglish string `bson:"titleEnglish"`
    Year         string `bson:"year"`
    ImdbID       string `bson:"imdbId"`
    TmdbID       string `bson:"tmdbId"`
    Image        string `bson:"image"`
    VideoURL     string `bson:"videoUrl"`
    Quality      string `bson:"quality"`
}


type MovieImage struct {
    ID    primitive.ObjectID `bson:"_id" json:"id"`
    Title string             `bson:"title"`
    TitleEnglish string      `bsin:"titleEnglish"`
    Image string             `bson:"image" json:"image"`
}


var movieCollection *mongo.Collection

func InitMongo(uri, dbName, collectionName string) error {
    timeOut := 10 * time.Second
    ctx, cancel := context.WithTimeout(context.Background(), timeOut)
    defer cancel()

    clientOpts := options.Client().ApplyURI(uri)

    client, err := mongo.Connect(ctx, clientOpts)

    if err != nil {
	return err
    }

    log.Println("MongoDB Connected")

    movieCollection = client.Database(dbName).Collection(collectionName)

    return nil
}

func InsertMovies(movies []Movie) error {
    if movieCollection == nil {
	return mongo.ErrClientDisconnected
    }

    timeOut := 15 * time.Second
    ctx, cancel := context.WithTimeout(context.Background(), timeOut)
    defer cancel()

    var docs []any
    for _, m := range movies {
	log.Printf("About to insert: %+v\n", m)
	docs = append(docs, m)
    }

    _, err := movieCollection.InsertMany(ctx, docs)
    return err
}


func GetAllMovies() ([]Movie, error) {
    if movieCollection == nil {
	return nil, mongo.ErrClientDisconnected
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer cancel()

    cursor, err := movieCollection.Find(ctx, bson.M{})
    if err != nil {
	return nil, err
    }
    defer cursor.Close(ctx)

    var movies []Movie
    for cursor.Next(ctx) {
	var m Movie
	if err := cursor.Decode(&m); err != nil {
	    return nil, err
	}
	movies = append(movies, m)
    }

    if err := cursor.Err(); err != nil {
	return nil, err
    }

    return movies, nil
}

func GetAllMovieImages() ([]MovieImage, error) {
    if movieCollection == nil {
	return nil, mongo.ErrClientDisconnected
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    opts := options.Find().SetProjection(bson.M{
	"image": 1,
	"title": 1,
	"titleEnglish" : 1,
    })

    cursor, err := movieCollection.Find(ctx, bson.M{}, opts)
    if err != nil {
	return nil, err
    }
    defer cursor.Close(ctx)

    var images []MovieImage
    for cursor.Next(ctx) {
	var mi MovieImage
	if err := cursor.Decode(&mi); err != nil {
	    return nil, err
	}
	images = append(images, mi)
    }

    return images, nil
}


func GetMovieByID(idStr string) (*Movie, error) {
    if movieCollection == nil {
	return nil, mongo.ErrClientDisconnected
    }

    id, err := primitive.ObjectIDFromHex(idStr)
    if err != nil {
	return nil, err // invalid hex id
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var movie Movie
    err = movieCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&movie)
    if err != nil {
	return nil, err
    }

    return &movie, nil
}

func GetAllMovieImdbIds() ([]string, error) {
    if movieCollection == nil {
	return nil, mongo.ErrClientDisconnected
    }

    timeOut := 10 * time.Second
    ctx, cancel := context.WithTimeout(context.Background(), timeOut)
    defer cancel()

    cursor, err := movieCollection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"imdbId": 1}))
    if err != nil {
	return nil, err
    }
    defer cursor.Close(ctx)

    var ids []string
    for cursor.Next(ctx) {
	var m struct {
	    ImdbID string `bson:"imdbId"`
	}
	if err := cursor.Decode(&m); err != nil {
	    return nil, err
	}
	ids = append(ids, m.ImdbID)
    }

    return ids, nil
}

func ClearMoviesCollection() error {
    if movieCollection == nil {
	return mongo.ErrClientDisconnected
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer cancel()

    _, err := movieCollection.DeleteMany(ctx, bson.M{})
    return err
}


func HasMovies() (bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
    defer cancel()

    count, err := movieCollection.CountDocuments(ctx, bson.D{})
    if err != nil {
	return false, err
    }
    return count > 0, nil
}

func RebuildTextIndex() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if _, err := movieCollection.Indexes().DropOne(ctx, "title_text"); err != nil {
	log.Printf("Failed to drop old index (title_text): %v", err)
    }

    index := mongo.IndexModel{
	Keys: bson.D{
	    {Key: "title", Value: "text"},
	    {Key: "titleEnglish", Value: "text"},
	},
    }
    _, err := movieCollection.Indexes().CreateOne(ctx, index)
    return err
}

func GetMovieByTmdbID(tmdbID string) (*Movie, error) {
    if movieCollection == nil {
	return nil, mongo.ErrClientDisconnected
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var movie Movie
    err := movieCollection.FindOne(ctx, bson.M{"tmdbId": tmdbID}).Decode(&movie)
    if err != nil {
	if err == mongo.ErrNoDocuments {
	    return nil, nil
	}
	return nil, err
    }

    return &movie, nil
}

func InsertMovie(movie Movie) error {
    if movieCollection == nil {
	return mongo.ErrClientDisconnected
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err := movieCollection.InsertOne(ctx, movie)
    return err
}

func SearchMoviesByTitle(query string) ([]Movie, error) {
    if movieCollection == nil {
	return nil, mongo.ErrClientDisconnected
    }

    safe := regexp.QuoteMeta(query)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    filter := bson.M{
	"$or": []bson.M{
	    {"title": bson.M{"$regex": safe, "$options": "i"}},
	    {"titleEnglish": bson.M{"$regex": safe, "$options": "i"}},
	},
    }

    cursor, err := movieCollection.Find(ctx, filter)
    if err != nil {
	return nil, err
    }
    defer cursor.Close(ctx)

    var movies []Movie
    for cursor.Next(ctx) {
	var m Movie
	if err := cursor.Decode(&m); err != nil {
	    return nil, err
	}
	movies = append(movies, m)
    }

    return movies, nil
}

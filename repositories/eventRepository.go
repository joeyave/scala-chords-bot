package repositories

import (
	"context"
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

type EventRepository struct {
	mongoClient *mongo.Client
}

func NewEventRepository(mongoClient *mongo.Client) *EventRepository {
	return &EventRepository{
		mongoClient: mongoClient,
	}
}

func (r *EventRepository) FindAll() ([]*entities.Event, error) {
	events, err := r.find(bson.M{"_id": bson.M{"$ne": ""}})
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepository) FindAllFromToday() ([]*entities.Event, error) {
	now := time.Now()

	return r.find(bson.M{
		"time": bson.M{
			"$gte": time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		},
	})
}

func (r *EventRepository) FindOneOldestByBandID(bandID primitive.ObjectID) (*entities.Event, error) {
	event, err := r.find(
		bson.M{
			"bandId": bandID,
		},
		bson.M{
			"$sort": bson.M{
				"time": 1,
			},
		},
		bson.M{
			"$limit": 1,
		},
	)
	if err != nil {
		return nil, err
	}

	return event[0], err
}

func (r *EventRepository) FindManyFromTodayByBandID(bandID primitive.ObjectID) ([]*entities.Event, error) {
	now := time.Now()

	return r.find(bson.M{
		"bandId": bandID,
		"time": bson.M{
			"$gte": time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		},
	})
}

func (r *EventRepository) FindManyFromTodayByBandIDAndUserID(bandID primitive.ObjectID, userID int64) ([]*entities.Event, error) {
	now := time.Now()

	return r.find(bson.M{
		"bandId":             bandID,
		"memberships.userId": userID,
		"time": bson.M{
			"$gte": time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		},
	})
}

func (r *EventRepository) FindMultipleByIDs(IDs []primitive.ObjectID) ([]*entities.Event, error) {
	return r.find(bson.M{
		"_id": bson.M{
			"$in": IDs,
		},
	})
}

func (r *EventRepository) FindOneByID(ID primitive.ObjectID) (*entities.Event, error) {
	events, err := r.find(bson.M{"_id": ID})
	if err != nil {
		return nil, err
	}

	return events[0], nil
}

func (r *EventRepository) FindOneByNameAndTime(name string, time time.Time) (*entities.Event, error) {
	events, err := r.find(
		bson.M{
			"name": name,
			"time": bson.M{
				"$gte": time,
				"$lt":  time.AddDate(0, 0, 1),
			},
		})

	if err != nil {
		return nil, err
	}

	return events[0], nil
}

func (r *EventRepository) find(m bson.M, opts ...bson.M) ([]*entities.Event, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("events")

	pipeline := bson.A{
		bson.M{
			"$lookup": bson.M{
				"from": "memberships",
				"let":  bson.M{"eventId": "$_id"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$eventId", "$$eventId"}}},
					},
					bson.M{
						"$lookup": bson.M{
							"from": "roles",
							"let":  bson.M{"roleId": "$roleId"},
							"pipeline": bson.A{
								bson.M{
									"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$roleId"}}},
								},
							},
							"as": "role",
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path":                       "$role",
							"preserveNullAndEmptyArrays": true,
						},
					},
					bson.M{
						"$sort": bson.D{
							{"role._id", 1},
							{"role.priority", 1},
						},
					},
					bson.M{
						"$lookup": bson.M{
							"from": "users",
							"let":  bson.M{"userId": "$userId"},
							"pipeline": bson.A{
								bson.M{
									"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$userId"}}},
								},
								bson.M{
									"$lookup": bson.M{
										"from": "bands",
										"let":  bson.M{"bandId": "$bandId"},
										"pipeline": bson.A{
											bson.M{
												"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$bandId"}}},
											},
											bson.M{
												"$lookup": bson.M{
													"from": "roles",
													"let":  bson.M{"bandId": "$_id"},
													"pipeline": bson.A{
														bson.M{
															"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$bandId", "$$bandId"}}},
														},
														bson.M{
															"$sort": bson.M{
																"priority": 1,
															},
														},
													},
													"as": "roles",
												},
											},
										},
										"as": "band",
									},
								},
								bson.M{
									"$unwind": bson.M{
										"path":                       "$band",
										"preserveNullAndEmptyArrays": true,
									},
								},
							},
							"as": "user",
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path":                       "$user",
							"preserveNullAndEmptyArrays": true,
						},
					},
				},
				"as": "memberships",
			},
		},
		bson.M{
			"$match": m,
		},
		bson.M{
			"$lookup": bson.M{
				"from": "bands",
				"let":  bson.M{"bandId": "$bandId"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$bandId"}}},
					},
					bson.M{
						"$lookup": bson.M{
							"from": "roles",
							"let":  bson.M{"bandId": "$_id"},
							"pipeline": bson.A{
								bson.M{
									"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$bandId", "$$bandId"}}},
								},
								bson.M{
									"$sort": bson.M{
										"priority": 1,
									},
								},
							},
							"as": "roles",
						},
					},
				},
				"as": "band",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$band",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"songIds": bson.M{
					"$cond": bson.M{
						"if": bson.M{
							"$ne": bson.A{bson.M{"$type": "$songIds"}, "array"},
						},
						"then": bson.A{},
						"else": "$songIds",
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "songs",
				"let":  bson.M{"songIds": "$songIds"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$_id", "$$songIds"}}},
					},
					bson.M{
						"$lookup": bson.M{
							"from": "bands",
							"let":  bson.M{"bandId": "$bandId"},
							"pipeline": bson.A{
								bson.M{
									"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$_id", "$$bandId"}}},
								},
							},
							"as": "band",
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path":                       "$band",
							"preserveNullAndEmptyArrays": true,
						},
					},
					bson.M{
						"$lookup": bson.M{
							"from": "voices",
							"let":  bson.M{"songId": "$_id"},
							"pipeline": bson.A{
								bson.M{
									"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$songId", "$$songId"}}},
								},
							},
							"as": "voices",
						},
					},
					bson.M{
						"$addFields": bson.M{
							"sort": bson.M{
								"$indexOfArray": bson.A{"$$songIds", "$_id"},
							},
						},
					},
					bson.M{
						"$sort": bson.M{"sort": 1},
					},
					bson.M{
						"$addFields": bson.M{
							"sort": "$$REMOVE",
						},
					},
				},
				"as": "songs",
			},
		},
		bson.M{
			"$sort": bson.M{
				"time": 1,
			},
		},
	}

	for _, o := range opts {
		pipeline = append(pipeline, o)
	}

	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	var events []*entities.Event
	err = cur.All(context.TODO(), &events)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return events, nil
}

func (r *EventRepository) UpdateOne(event entities.Event) (*entities.Event, error) {
	if event.ID.IsZero() {
		event.ID = r.generateUniqueID()
	}

	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("events")

	filter := bson.M{"_id": event.ID}

	event.Memberships = nil
	event.Band = nil
	event.Songs = nil
	update := bson.M{
		"$set": event,
	}

	after := options.After
	upsert := true
	opts := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	result := collection.FindOneAndUpdate(context.TODO(), filter, update, &opts)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var newEvent *entities.Event
	err := result.Decode(&newEvent)
	if err != nil {
		return nil, err
	}

	return r.FindOneByID(newEvent.ID)
}

func (r *EventRepository) DeleteOneByID(ID primitive.ObjectID) error {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("events")

	_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": ID})
	return err
}

func (r *EventRepository) GetSongs(eventID primitive.ObjectID) ([]*entities.Song, error) {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("events")

	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": eventID,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"songIds": bson.M{
					"$cond": bson.M{
						"if": bson.M{
							"$ne": bson.A{bson.M{"$type": "$songIds"}, "array"},
						},
						"then": bson.A{},
						"else": "$songIds",
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from": "songs",
				"let":  bson.M{"songIds": "$songIds"},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$_id", "$$songIds"}}},
					},
					bson.M{
						"$addFields": bson.M{
							"sort": bson.M{
								"$indexOfArray": bson.A{"$$songIds", "$_id"},
							},
						},
					},
					bson.M{
						"$sort": bson.M{"sort": 1},
					},
					bson.M{
						"$addFields": bson.M{
							"sort": "$$REMOVE",
						},
					},
				},
				"as": "songs",
			},
		},
	}

	cur, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	var events []*entities.Event
	err = cur.All(context.TODO(), &events)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return []*entities.Song{}, nil
	}

	return events[0].Songs, nil
}

func (r *EventRepository) PushSongID(eventID primitive.ObjectID, songID primitive.ObjectID) error {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("events")

	filter := bson.M{
		"_id":     eventID,
		"songIds": bson.M{"$nin": bson.A{songID}},
	}

	update := bson.M{
		"$push": bson.M{
			"songIds": songID,
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *EventRepository) ChangeSongIDPosition(eventID primitive.ObjectID, songID primitive.ObjectID, newPosition int) error {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("events")

	filter := bson.M{
		"_id": eventID,
	}

	update := bson.M{
		"$pull": bson.M{
			"songIds": songID,
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)

	filter = bson.M{
		"_id":     eventID,
		"songIds": bson.M{"$nin": bson.A{songID}},
	}

	update = bson.M{
		"$push": bson.M{
			"songIds": bson.M{
				"$each":     bson.A{songID},
				"$position": newPosition,
			},
		},
	}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *EventRepository) PullSongID(eventID primitive.ObjectID, songID primitive.ObjectID) error {
	collection := r.mongoClient.Database(os.Getenv("MONGODB_DATABASE_NAME")).Collection("events")

	filter := bson.M{"_id": eventID}

	update := bson.M{
		"$pull": bson.M{
			"songIds": songID,
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *EventRepository) generateUniqueID() primitive.ObjectID {
	ID := primitive.NilObjectID

	for ID.IsZero() {
		ID = primitive.NewObjectID()
		_, err := r.FindOneByID(ID)
		if err == nil {
			ID = primitive.NilObjectID
		}
	}

	return ID
}

package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/drive/v3"
	"net/url"
	"time"
)

type User struct {
	ID    int64  `bson:"_id,omitempty"`
	Name  string `bson:"name,omitempty"`
	Role  string `bson:"role,omitempty"`
	State *State `bson:"state,omitempty"`

	BandID primitive.ObjectID `bson:"bandId,omitempty"`
	Band   *Band              `bson:"band,omitempty"`
}

type UserExtra struct {
	User *User `bson:",inline"`

	Events []*Event `bson:"events,omitempty"`
}

type State struct {
	Index        int      `bson:"index,omitempty"`
	Name         int      `bson:"name,omitempty"`
	Context      Context  `bson:"context,omitempty"`
	CallbackData *url.URL `bson:"-"`

	Prev *State `bson:"prev,omitempty"`
	Next *State `bson:"next,omitempty"`
}

type Context struct {
	SongNames        []string `bson:"songNames,omitempty"`
	MessagesToDelete []int    `bson:"messagesToDelete,omitempty"`
	Query            string   `bson:"query,omitempty"`
	QueryType        string   `bson:"queryType,omitempty"`

	DriveFileID       string        `bson:"currentSongId,omitempty"`
	FoundDriveFileIDs []string      `bson:"foundDriveFileIds,omitempty"`
	DriveFiles        []*drive.File `bson:"driveFiles,omitempty"`

	Voice *Voice `bson:"currentVoice,omitempty"`

	Band  *Band   `bson:"currentBand,omitempty"`
	Bands []*Band `bson:"bands,omitempty"`

	Role *Role `bson:"role,omitempty"`

	EventID primitive.ObjectID `bson:"eventId,omitempty"`

	CreateSongPayload struct {
		Name   string `bson:"name,omitempty"`
		Lyrics string `bson:"lyrics,omitempty"`
		Key    string `bson:"key,omitempty"`
		BPM    string `bson:"bpm,omitempty"`
		Time   string `bson:"time,omitempty"`
	} `bson:"createSongPayload,omitempty"`

	Map  map[string]string `bson:"map,omitempty"`
	Time time.Time         `bson:"time,omitempty"`

	PageIndex int `bson:"index, omitempty"`

	NextPageToken *NextPageToken `bson:"nextPageToken,omitempty"`
}

type NextPageToken struct {
	Token         string         `bson:"token"`
	PrevPageToken *NextPageToken `bson:"prevPageToken,omitempty"`
}

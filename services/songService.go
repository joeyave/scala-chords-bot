package services

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/drive/v3"
	"scalaChordsBot/entities"
	"scalaChordsBot/repositories"
	"time"
)

type SongService struct {
	songRepository *repositories.SongRepository
	driveClient    *drive.Service
}

func NewSongService(songRepository *repositories.SongRepository, driveClient *drive.Service) *SongService {
	return &SongService{
		songRepository: songRepository,
		driveClient:    driveClient,
	}
}

/*
Searches for Song on Google Drive then returns uncached versions of Songs for performance reasons.
*/
func (s *SongService) FindByName(name string) ([]entities.Song, error) {

	var songs []entities.Song

	var pageToken string

	for {
		res, err := s.driveClient.Files.List().
			Q(fmt.Sprintf("fullText contains '\"%s\"'", name)).
			Fields("nextPageToken, files(id, name, modifiedTime, webViewLink)").
			PageToken(pageToken).
			Do()

		if err != nil {
			return nil, err
		}

		for _, file := range res.Files {
			actualSong := entities.Song{
				ID:           &file.Id,
				Name:         file.Name,
				ModifiedTime: file.ModifiedTime,
				WebViewLink:  file.WebViewLink,
			}

			songs = append(songs, actualSong)
		}

		pageToken = res.NextPageToken

		if pageToken == "" {
			break
		}
	}

	if len(songs) == 0 {
		return nil, mongo.ErrEmptySlice
	}

	return songs, nil
}

func (s *SongService) FindCached(song entities.Song) (entities.Song, error) {
	if song.ID == nil {
		return song, fmt.Errorf("ID is missing for Song: %v", song)
	}

	cachedSong, err := s.songRepository.FindOneByID(*song.ID)
	if err != nil {
		return song, err
	}

	song.Voices = cachedSong.Voices

	cachedModifiedTime, err := time.Parse(time.RFC3339, cachedSong.ModifiedTime)
	actualModifiedTime, err := time.Parse(time.RFC3339, song.ModifiedTime)

	if err != nil || !actualModifiedTime.After(cachedModifiedTime) {
		song.TgFileID = cachedSong.TgFileID
	}

	return song, err
}

func (s *SongService) DownloadPDF(song entities.Song) (*tgbotapi.FileReader, error) {
	if song.ID == nil {
		return nil, fmt.Errorf("ID is missing for Song: %v", song)
	}

	res, err := s.driveClient.Files.Export(*song.ID, "application/pdf").Download()
	if err != nil {
		return nil, err
	}

	fileReader := &tgbotapi.FileReader{
		Name:   song.Name + ".pdf",
		Reader: res.Body,
		Size:   res.ContentLength,
	}

	return fileReader, err
}

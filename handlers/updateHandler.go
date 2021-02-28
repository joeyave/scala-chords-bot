package handlers

import (
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/helpers"
	"github.com/joeyave/scala-chords-bot/services"
	tgbotapi "github.com/joeyave/telegram-bot-api/v5"
	"os"
	"strconv"
)

type UpdateHandler struct {
	bot         *tgbotapi.BotAPI
	userService *services.UserService
	songService *services.SongService
	bandService *services.BandService
}

func NewHandler(bot *tgbotapi.BotAPI, userService *services.UserService, songService *services.SongService, bandService *services.BandService) *UpdateHandler {
	return &UpdateHandler{
		bot:         bot,
		userService: userService,
		songService: songService,
		bandService: bandService,
	}
}

func (u *UpdateHandler) HandleUpdate(update *tgbotapi.Update) error {
	user, err := u.userService.FindOrCreate(update.Message.Chat.ID)

	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Что-то пошло не так.")
		_, _ = u.bot.Send(msg)
		return fmt.Errorf("couldn't get User's state %v", err)
	}

	// Handle buttons.
	switch update.Message.Text {
	case helpers.Cancel:
		if user.State.Prev != nil {
			user.State = user.State.Prev
			user.State.Index = 0
		} else {
			user.State = &entities.State{
				Index: 0,
				Name:  helpers.MainMenuState,
			}
		}
	}

	// Catch voice anywhere.
	if update.Message.Voice != nil {
		user.State = &entities.State{
			Index: 0,
			Name:  helpers.UploadVoiceState,
			Context: entities.Context{
				CurrentVoice: &entities.Voice{
					TgFileID: update.Message.Voice.FileID,
				},
			},
			Prev: user.State,
		}
	}

	user, err = u.enterStateHandler(update, *user)

	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка. Поправим.")
		_, _ = u.bot.Send(msg)

		channelId, convErr := strconv.ParseInt(os.Getenv("LOG_CHANNEL"), 10, 0)
		if convErr == nil {
			msg = tgbotapi.NewMessage(channelId, fmt.Sprintf("<code>%v</code>", err))
			msg.ParseMode = tgbotapi.ModeHTML
			_, _ = u.bot.Send(msg)
		}
	} else {
		user, err = u.userService.UpdateOne(*user)
	}

	return err
}

func (u *UpdateHandler) enterStateHandler(update *tgbotapi.Update, user entities.User) (*entities.User, error) {
	handleFuncs, ok := stateHandlers[user.State.Name]

	if ok == false || user.State.Index >= len(handleFuncs) || user.State.Index < 0 {
		user.State.Index = 0
		user.State.Name = helpers.MainMenuState
		handleFuncs = stateHandlers[user.State.Name]
	}

	return handleFuncs[user.State.Index](u, update, user)
}

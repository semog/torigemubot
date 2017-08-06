package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

type botUpdateEvent func(bot *tgbotapi.BotAPI, msg *tgbotapi.Update) bool
type botMessageEvent func(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) bool
type botInlineQueryEvent func(bot *tgbotapi.BotAPI, query *tgbotapi.InlineQuery) bool
type botChosenInlineResultEvent func(bot *tgbotapi.BotAPI, result *tgbotapi.ChosenInlineResult) bool
type botCallbackQueryEvent func(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) bool
type botShippingQueryEvent func(bot *tgbotapi.BotAPI, query *tgbotapi.ShippingQuery) bool
type botPreCheckoutQueryEvent func(bot *tgbotapi.BotAPI, query *tgbotapi.PreCheckoutQuery) bool

type botEventHandlers struct {
	onUpdate             botUpdateEvent
	onMessage            botMessageEvent
	onEditedMessage      botMessageEvent
	onChannelPost        botMessageEvent
	onEditedChannelPost  botMessageEvent
	onInlineQuery        botInlineQueryEvent
	onChosenInlineResult botChosenInlineResultEvent
	onCallbackQuery      botCallbackQueryEvent
	onShippingQuery      botShippingQueryEvent
	onPreCheckoutQuery   botPreCheckoutQueryEvent
}

func runBot(bot *tgbotapi.BotAPI, botUpdate botEventHandlers) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	var keepgoing = true
	for update := range updates {
		// Support generic handling of the raw update message.
		if botUpdate.onUpdate != nil && !botUpdate.onUpdate(bot, &update) {
			break
		}

		// Call specific event handlers
		switch {
		case update.Message != nil && botUpdate.onMessage != nil:
			keepgoing = botUpdate.onMessage(bot, update.Message)
		case update.EditedMessage != nil && botUpdate.onEditedMessage != nil:
			keepgoing = botUpdate.onEditedMessage(bot, update.EditedMessage)
		case update.ChannelPost != nil && botUpdate.onChannelPost != nil:
			keepgoing = botUpdate.onChannelPost(bot, update.ChannelPost)
		case update.EditedChannelPost != nil && botUpdate.onEditedChannelPost != nil:
			keepgoing = botUpdate.onEditedChannelPost(bot, update.EditedChannelPost)
		case update.InlineQuery != nil && botUpdate.onInlineQuery != nil:
			keepgoing = botUpdate.onInlineQuery(bot, update.InlineQuery)
		case update.ChosenInlineResult != nil && botUpdate.onChosenInlineResult != nil:
			keepgoing = botUpdate.onChosenInlineResult(bot, update.ChosenInlineResult)
		case update.CallbackQuery != nil && botUpdate.onCallbackQuery != nil:
			keepgoing = botUpdate.onCallbackQuery(bot, update.CallbackQuery)
		case update.ShippingQuery != nil && botUpdate.onShippingQuery != nil:
			keepgoing = botUpdate.onShippingQuery(bot, update.ShippingQuery)
		case update.PreCheckoutQuery != nil && botUpdate.onPreCheckoutQuery != nil:
			keepgoing = botUpdate.onPreCheckoutQuery(bot, update.PreCheckoutQuery)
		default:
			if bot.Debug {
				log.Print("Unhandled Bot Event...")
			}
			keepgoing = true
		}

		if !keepgoing {
			break
		}
	}
	log.Printf("Shutting down %s", bot.Self.UserName)
}

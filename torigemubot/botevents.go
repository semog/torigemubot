package main

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

type botUpdateEvent func(bot *tg.BotAPI, msg *tg.Update) bool
type botMessageEvent func(bot *tg.BotAPI, msg *tg.Message) bool
type botInlineQueryEvent func(bot *tg.BotAPI, query *tg.InlineQuery) bool
type botChosenInlineResultEvent func(bot *tg.BotAPI, result *tg.ChosenInlineResult) bool
type botCallbackQueryEvent func(bot *tg.BotAPI, query *tg.CallbackQuery) bool
type botShippingQueryEvent func(bot *tg.BotAPI, query *tg.ShippingQuery) bool
type botPreCheckoutQueryEvent func(bot *tg.BotAPI, query *tg.PreCheckoutQuery) bool

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

func runBot(bot *tg.BotAPI, handler botEventHandlers) {
	u := tg.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	var keepgoing = true
	for update := range updates {
		// Support generic handling of the raw update message.
		if handler.onUpdate != nil && !handler.onUpdate(bot, &update) {
			break
		}

		// Call specific event handlers
		switch {
		case update.Message != nil && handler.onMessage != nil:
			keepgoing = handler.onMessage(bot, update.Message)
		case update.EditedMessage != nil && handler.onEditedMessage != nil:
			keepgoing = handler.onEditedMessage(bot, update.EditedMessage)
		case update.ChannelPost != nil && handler.onChannelPost != nil:
			keepgoing = handler.onChannelPost(bot, update.ChannelPost)
		case update.EditedChannelPost != nil && handler.onEditedChannelPost != nil:
			keepgoing = handler.onEditedChannelPost(bot, update.EditedChannelPost)
		case update.InlineQuery != nil && handler.onInlineQuery != nil:
			keepgoing = handler.onInlineQuery(bot, update.InlineQuery)
		case update.ChosenInlineResult != nil && handler.onChosenInlineResult != nil:
			keepgoing = handler.onChosenInlineResult(bot, update.ChosenInlineResult)
		case update.CallbackQuery != nil && handler.onCallbackQuery != nil:
			keepgoing = handler.onCallbackQuery(bot, update.CallbackQuery)
		case update.ShippingQuery != nil && handler.onShippingQuery != nil:
			keepgoing = handler.onShippingQuery(bot, update.ShippingQuery)
		case update.PreCheckoutQuery != nil && handler.onPreCheckoutQuery != nil:
			keepgoing = handler.onPreCheckoutQuery(bot, update.PreCheckoutQuery)
		default:
			if bot.Debug {
				log.Printf("Unhandled Bot Event: %v", update)
			}
			keepgoing = true
		}

		if !keepgoing {
			break
		}
	}
	log.Printf("Shutting down %s", bot.Self.UserName)
}

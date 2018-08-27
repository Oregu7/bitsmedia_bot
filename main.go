package main

import (
	"fmt"
	"log"
	"os"
	"parsers/bitsmedia/helpers"

	"github.com/jasonlvhit/gocron"
	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
	"github.com/syndtr/goleveldb/leveldb"
	"gopkg.in/telegram-bot-api.v4"
)

func main() {
	godotenv.Load()
	fp := gofeed.NewParser()
	db, err := leveldb.OpenFile("storage.db", nil)
	if err != nil {
		log.Panic(err)
	}
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	gocron.Every(20).Minutes().Do(listener, fp, db, bot)
	<-gocron.Start()
}

func listener(fp *gofeed.Parser, db *leveldb.DB, bot *tgbotapi.BotAPI) {
	items := helpers.ParseUpdates("https://bits.media/news/")
	updates := helpers.GetUpdates(db, items)
	fmt.Printf("updates => %v\n", len(updates))
	for indx := range updates {
		helpers.PostToChannel(bot, updates[len(updates)-1-indx])
	}
}

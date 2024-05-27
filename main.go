package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gabriel-vasile/mimetype"
	"github.com/joho/godotenv"
	"io"
	"log"
	"macOS-auto-backup-to-Discord-be/configs"
	"macOS-auto-backup-to-Discord-be/handlers/error_handlers"
	"macOS-auto-backup-to-Discord-be/prisma/db"
	"macOS-auto-backup-to-Discord-be/utils/files"
	"macOS-auto-backup-to-Discord-be/utils/security"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var ctx = context.Background()

var client *db.PrismaClient

var discordBot *discordgo.Session

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	_discordBot, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		fmt.Println("Error initializing go project: ", err)
		return err
	}
	discordBot = _discordBot

	err = discordBot.Open()
	if err != nil {
		fmt.Println("Error opening connection to Discord: ", err)
	}
	defer func(discordBot *discordgo.Session) {
		err := discordBot.Close()
		if err != nil {
			panic(err)
		}
	}(discordBot)

	client = db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		fmt.Println("Error connecting to database: ", err)
		return err
	}
	defer func() {
		if err := client.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/message", &messageHandler{})

	log.Println("Server started on localhost:8080")
	log.Fatalln(http.ListenAndServe(":8080", mux))

	return nil
}

type messageHandler struct{}

func (h *messageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// This endpoint only allows POST method
	if r.Method != http.MethodPost {
		error_handlers.MethodNotAllowedHandler(w, r)
		return
	}
	err := r.ParseMultipartForm(configs.MaxUploadSize)
	mimetype.SetLimit(uint32(configs.MaxUploadSize))
	if err != nil {
		error_handlers.InternalServerErrorHandler(w, r)
		log.Fatalln("Error parsing form: ", err)
	}

	// Receive uploaded file and write it locally
	file, fileHeader, err := r.FormFile("File")
	if err != nil {
		error_handlers.InternalServerErrorHandler(w, r)
		log.Fatalln("Error parsing file: ", err)
	}

	fileName := fileHeader.Filename
	fileBytes, err := io.ReadAll(file)
	fileType := mimetype.Detect(fileBytes).String()
	fileName = fileName[:len(fileName)-len(fileName[strings.LastIndex(fileName, "."):])]

	newFile, err := client.File.CreateOne(
		db.File.FileName.Set(fileName),
		db.File.FileType.Set(fileType),
	).Exec(ctx)
	if err != nil {
		error_handlers.InternalServerErrorHandler(w, r)
		log.Fatalln("Error creating new file: ", err)
	}

	encryptedFileBytes, err := security.Encrypt([]byte(os.Getenv("ENCRYPTION_KEY")), fileBytes)
	var newChunks = files.Chunkify(encryptedFileBytes)
	fmt.Println("Number of chunks: ", len(newChunks))
	discordChannelID := os.Getenv("DISCORD_CHANNEL_ID")
	for idx, newByteChunk := range newChunks {
		newFileName := "Chunk #" + strconv.FormatInt(int64(idx), 10) + " of " + fileName + ".txt"
		discordSent, err := discordBot.ChannelFileSend(discordChannelID, newFileName, bytes.NewReader(newByteChunk))
		if err != nil {
			fmt.Println("Error sending file to Discord: ", err)
			return
		}

		newChunk, err := client.ChunkFile.CreateOne(
			db.ChunkFile.File.Link(db.File.ID.Equals(newFile.ID)),
			db.ChunkFile.DiscordMessageID.Set(discordSent.ID),
			db.ChunkFile.Order.Set(idx),
		).Exec(ctx)
		if err != nil {
			error_handlers.InternalServerErrorHandler(w, r)
			log.Fatalln("Error creating new chunk: ", err)
		}
		fmt.Println(newChunk)
	}
}

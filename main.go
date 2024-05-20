package main

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gabriel-vasile/mimetype"
	"io"
	"log"
	"macOS-auto-backup-to-Discord-be/handlers/error_handlers"
	"macOS-auto-backup-to-Discord-be/prisma/db"
	"net/http"
)

const MaxUploadSize int64 = 2 * 1024 * 1024 // 2 MB
const uploadPath = "./tmp"

var ctx = context.Background()

var client *db.PrismaClient

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	_, err := discordgo.New("Bot " + "authentication token")
	if err != nil {
		fmt.Println("Error initializing go project: ", err)
		return err
	}

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
	err := r.ParseMultipartForm(MaxUploadSize)
	mimetype.SetLimit(uint32(MaxUploadSize))
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

	newFile, err := client.File.CreateOne(
		db.File.FileName.Set(fileName),
		db.File.FileType.Set(fileType),
	).Exec(ctx)
	if err != nil {
		error_handlers.InternalServerErrorHandler(w, r)
		log.Fatalln("Error creating new file: ", err)
	}

	newChunk, err := client.ChunkFile.CreateOne(
		db.ChunkFile.File.Link(db.File.ID.Equals(newFile.ID)),
		db.ChunkFile.DiscordMessageID.Set(""),
		db.ChunkFile.Order.Set(0),
	).Exec(ctx)
	if err != nil {
		error_handlers.InternalServerErrorHandler(w, r)
		log.Fatalln("Error creating new chunk: ", err)
	}

	fmt.Println(newChunk)
}

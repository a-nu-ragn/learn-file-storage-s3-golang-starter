package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")

	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil || (mediatype != "image/jpeg" && mediatype != "image/png"){
		respondWithError(w, http.StatusBadRequest, "Bad file format", err)
		return
	}

	videoMeta, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	thumbnailName := fmt.Sprintf("%v.%v", videoIDString, strings.Split(contentType, "/")[1])

	thumbnailPath := filepath.Join(cfg.assetsRoot, thumbnailName)

	f, err := os.Create(thumbnailPath)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
		return
	}

	_, err = io.Copy(f, file)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
		return
	}

	fileURL := fmt.Sprintf("http://localhost:%v/assets/%v", cfg.port, thumbnailName)
	videoMeta.ThumbnailURL = &fileURL

	cfg.db.UpdateVideo(videoMeta)

	respondWithJSON(w, http.StatusOK, videoMeta)
}

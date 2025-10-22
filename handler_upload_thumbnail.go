package main

import (
	"fmt"
	"net/http"
	"io"
	"os"
	"mime"
	"encoding/base64"
	"crypto/rand"
	"path/filepath"

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
	

	vid, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Video does not exist", err)
		return
	}
	
	if vid.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
		
	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png"{
		respondWithError(w, http.StatusUnauthorized, "Invalid mime type", nil)
		return
	}

	ext, err := mime.ExtensionsByType(contentType)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Error reading file extension", err)
		return
	}

	key := make([]byte, 32)
	rand.Read(key)

	fileName := fmt.Sprintf("%s%s", base64.RawURLEncoding.EncodeToString(key), ext[0])
	path := filepath.Join(cfg.assetsRoot, fileName)
	
	osFile, err := os.Create(path)
	if err != nil{
		respondWithError(w, http.StatusBadRequest, "Error creating file", err)
		return
	}

	_, err = io.Copy(osFile, file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error populating file", err)
		return
	}

	thumbnailUrl := fmt.Sprintf("http://localhost:8091/assets/%s", fileName) 
	vid.ThumbnailURL = &thumbnailUrl 
	cfg.db.UpdateVideo(vid)

	respondWithJSON(w, http.StatusOK, vid)
}

package main

import (
	"net/http"
	"fmt"
	"mime"
	"context"
	"crypto/rand"
	"encoding/base64"
	"os"
	"io"
	
	"github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("uploading video", videoID, "by user", userID)

	vid, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Video does not exist", err)
		return
	}

	if vid.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	const maxMemory = 1 << 30
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()
	
	contentType := header.Header.Get("Content-Type")
	if contentType != "video/mp4"{
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
	
	tmpFile, err := os.CreateTemp("", fileName)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error creating tmp file", err)
		return
	}
	
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error populating file", err)
		return
	}
	
	fastStart, err := processVideoForFastStart(tmpFile.Name())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error preprocessing file", err)
		return
	}
	defer os.Remove(fastStart)
	
	ratio, err := getVideoAspectRatio(fastStart)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error while calculating aspect ratio of file", err)
		return
	}

	var prefix string
	switch ratio {
	case "16:9":
		prefix = "landscape"
	case "9:16":
		prefix = "portrait"
	default:
		prefix = "other"
	}

	f, err := os.Open(fastStart)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error while calculating aspect ratio of file", err)
		return
	}
	defer f.Close()

	fileUrl := fmt.Sprintf("%s/%s", prefix, fileName)
	input := &s3.PutObjectInput{
		Bucket: aws.String(cfg.s3Bucket),
		Key: aws.String(fileUrl),
		Body: f,
		ContentType: aws.String(contentType),
	}
	
	_, err = cfg.s3Client.PutObject(context.TODO(), input)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error failed to upload file", err)
		return
	}
	
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, fileUrl)
	vid.VideoURL = &url
	cfg.db.UpdateVideo(vid)

	respondWithJSON(w, http.StatusOK, vid)
}

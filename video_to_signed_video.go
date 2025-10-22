package main

import(
	"strings"
	"time"
	"fmt"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error){
	if video.VideoURL == nil {
		return video, nil
	}

	values := strings.Split(*video.VideoURL, ",")
	if len(values) < 2 {
		return database.Video{}, fmt.Errorf("video.VideoURL has invalid format: %s", *video.VideoURL)
	}

	signedUrl, err := generatePresignedURL(cfg.s3Client, values[0], values[1], 10 * time.Minute)
	if err != nil {
		return database.Video{}, fmt.Errorf("failed to generate presigned url: %s", err)
	}

	video.VideoURL = &signedUrl
	return video, nil
}

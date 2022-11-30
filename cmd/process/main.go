package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/gif"
	_ "net/http/pprof"

	"gif-doggo/internal/logger"

	"github.com/andybons/gogif"
	"github.com/go-redis/redis/v9"
	"golang.org/x/image/draw"
)

var redis_client *redis.Client

func init() {
	redis_client = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
}

type doggo_request struct {
	Images     []string `json:"images"`
	Output     string   `json:"output"`
	Delays     []int    `json:"delays"`
	Expiration int      `json:"expiration"`
}

func main() {
	ctx := context.Background()
	subscriber := redis_client.Subscribe(ctx, "doggos")
	for {
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			logger.Errorw("Failed to receive message", "error", err)
		}

		req := doggo_request{}
		if err := json.Unmarshal([]byte(msg.Payload), &req); err != nil {
			logger.Errorw("Invalid request", "error", err)
			return
		}

		uid_string := msg.Payload
		err = redis_client.Set(context.Background(), uid_string, "processing", 3600).Err()
		if err != nil {
			logger.Errorw("Failed to update request status", "error", err)
			continue
		}

		process(&req)

		err = redis_client.Set(context.Background(), uid_string, "done", 3600).Err()
		if err != nil {
			logger.Errorw("Failed to update request status", "error", err)
			continue
		}
	}
}

func process(req *doggo_request) (*bytes.Buffer, error) {

	var images []*image.Paletted

	var rect image.Rectangle
	skip := false

	for _, image_key := range req.Images {
		image_body, err := redis_client.Get(context.Background(), image_key).Bytes()
		if err != nil {
			logger.Errorw("Failed to get image", "name", image_key, "error", err)
			return nil, err
		}

		img, _, err := image.Decode(bytes.NewReader(image_body))
		if err != nil {
			logger.Errorw("Failed decode", "name", image_key, "error", err)
			return nil, err
		}

		if !skip {
			rect = img.Bounds()
			skip = true
		}

		palettedImage := image.NewPaletted(rect, nil)
		quantizer := gogif.MedianCutQuantizer{NumColor: 64}
		quantizer.Quantize(palettedImage, rect, img, image.Point{0, 0})

		draw.CatmullRom.Scale(palettedImage, palettedImage.Rect, img, img.Bounds(), draw.Over, nil)

		images = append(images, palettedImage)
	}

	var b bytes.Buffer
	buf_writer := bufio.NewWriter(&b)

	return &b, gif.EncodeAll(buf_writer, &gif.GIF{
		Image: images,
		Delay: req.Delays,
	})
}

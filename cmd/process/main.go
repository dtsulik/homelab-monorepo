package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/gif"
	"net/http"
	_ "net/http/pprof"
	"time"

	"gif-doggo/internal/jaegerexport"
	"gif-doggo/internal/logger"

	"github.com/andybons/gogif"
	"github.com/go-redis/redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/image/draw"
)

var redis_client *redis.Client
var tracer_name = "doggo-processing"

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
	tp, err := jaegerexport.JaegerTraceProvider("http://jaeger:14268/api/traces")
	if err != nil {
		logger.Fatalw("Failed to create tracer provider", "error", err)
	}

	otel.SetTracerProvider(tp)

	ctx := context.Background()

	// TODO add redis otel
	ctx, span := otel.Tracer(tracer_name).Start(ctx, "receive-requests")
	defer span.End()

	go func() {
		logger.Infow("Starting server for health/readiness checks", "port", 80)
		http.HandleFunc("/readyz", func(w http.ResponseWriter, request *http.Request) {})
		http.HandleFunc("/livez", func(w http.ResponseWriter, request *http.Request) {})
		err := http.ListenAndServe(":80", nil)
		if err != nil {
			logger.Fatalw("Failed to start server", "error", err)
		}
	}()

	subscriber := redis_client.Subscribe(ctx, "doggos")
	for {
		// FIXME change this to channel implementation
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			logger.Errorw("Failed to receive message", "error", err)
		}

		if msg == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		req := doggo_request{}
		if err := json.Unmarshal([]byte(msg.Payload), &req); err != nil {
			logger.Errorw("Invalid request", "error", err)
			return
		}

		uid_string := msg.Payload
		err = redis_client.Set(ctx, uid_string, "processing", 3600).Err()
		if err != nil {
			logger.Errorw("Failed to update request status", "error", err)
			continue
		}

		process(ctx, &req)

		err = redis_client.Set(ctx, uid_string, "done", 3600).Err()
		if err != nil {
			logger.Errorw("Failed to update request status", "error", err)
			continue
		}
	}
}

func process(ctx context.Context, req *doggo_request) (*bytes.Buffer, error) {
	_, span := otel.Tracer(tracer_name).Start(ctx, "image-processing")
	defer span.End()

	var images []*image.Paletted

	var rect image.Rectangle
	skip := false

	for idx, image_key := range req.Images {
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
		span.AddEvent("Processed image",
			trace.WithAttributes(attribute.Int("no", idx),
				attribute.String("name", image_key),
				attribute.Int("size", len(image_body))))
	}

	var b bytes.Buffer
	buf_writer := bufio.NewWriter(&b)

	return &b, gif.EncodeAll(buf_writer, &gif.GIF{
		Image: images,
		Delay: req.Delays,
	})
}

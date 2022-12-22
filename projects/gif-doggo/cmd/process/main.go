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

	"gif-doggo/internal/helpers"
	"gif-doggo/internal/jaegerexport"
	"gif-doggo/internal/logger"

	"github.com/andybons/gogif"

	"github.com/go-redis/redis/extra/redisotel/v9"
	"github.com/go-redis/redis/v9"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"golang.org/x/image/draw"
)

var (
	redis_client     *redis.Client
	tracer_name      = "doggo-processing"
	doggos_processed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_doggos_processed",
		Help: "The total number of doggo images ingested for creating GIFs",
	})
	doggos_created = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_doggos_created",
		Help: "The total number of doggo GIFs created",
	})
	doggos_failed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_doggos_failed",
		Help: "The total number of doggo GIFs failed to create",
	})
)

func init() {
	redis_client = redis.NewClient(&redis.Options{
		Addr: helpers.GetEnv("REDIS_ENDPOINT", "redis:6379"),
	})
	if err := redisotel.InstrumentTracing(redis_client); err != nil {
		logger.Fatalw("Unable to start redis otel")
	}
}

type doggo_request struct {
	Images     []string `json:"images"`
	Output     string   `json:"output"`
	Delays     []int    `json:"delays"`
	Expiration int      `json:"expiration"`
}

func main() {
	tc_ep := helpers.GetEnv("TRACECOLLECTOR_ENDPOINT", "http://jaeger:14268/api/traces")
	logger.Infow("Sending traces to " + tc_ep)

	tp, err := jaegerexport.JaegerTraceProvider(tracer_name, tc_ep)
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
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":80", nil)
		if err != nil {
			logger.Fatalw("Failed to start server", "error", err)
		}
	}()

	subscriber := redis_client.Subscribe(ctx, "doggos")
	for {
		// FIXME change this to channel implementation
		status := "processing"
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
		err = redis_client.Set(ctx, uid_string, status, 3600).Err()
		if err != nil {
			logger.Errorw("Failed to update request status", "error", err)
			continue
		}

		// TODO store output doggos
		_, err = process(ctx, &req)
		if err != nil {
			logger.Errorw("Failed to process doggo", "error", err)
			status = "failed"
			doggos_failed.Inc()
		} else {
			status = "done"
			doggos_created.Inc()
		}

		err = redis_client.Set(ctx, uid_string, status, 3600).Err()
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
		image_body, err := redis_client.Get(ctx, image_key).Bytes()
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
		doggos_processed.Inc()
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

package main

import (
	"bytes"
	"context"
	_ "embed"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

var port string

func main() {
	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true), // headless=false に変更
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	chromedp.Run(ctx,
		chromedp.Navigate(os.Getenv("GRAFANA_DASHBOARD_URL")),
		chromedp.Sleep(2*time.Second),
		chromedp.SendKeys(`input[name='user']`, os.Getenv("GRAFANA_USER")),
		chromedp.SendKeys(`input[name='password']`, os.Getenv("GRAFANA_PASSWORD")),
		chromedp.Click(`button[type='submit']`),
	)
	if err := srv.ListenAndServe(); err != nil {
		log.Print(err)
	}

}

var ctx context.Context

func handler(w http.ResponseWriter, r *http.Request) {
	var res []byte
	chromedp.Run(ctx,
		chromedp.Navigate(os.Getenv("GRAFANA_DASHBOARD_URL")),
		chromedp.WaitVisible(`div.u-over`),
		chromedp.EmulateViewport(540, 950),
		chromedp.Screenshot(`div.grafana-app`, &res),
	)
	imageP, _ := png.Decode(bytes.NewReader(res))
	tmp := filepath.Join(os.TempDir(), "grafana.jpg")
	f, _ := os.Create(tmp)
	jpeg.Encode(f, imageP, nil)
	http.ServeFile(w, r, tmp)
}

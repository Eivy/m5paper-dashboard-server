package main

import (
	"bytes"
	"context"
	_ "embed"
	"image"
	"image/color"
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
	if os.Getenv("GRAFANA_NEED_LOGIN") != "" {
		chromedp.Run(ctx,
			chromedp.Navigate(os.Getenv("GRAFANA_DASHBOARD_URL")),
			chromedp.Sleep(2*time.Second),
			chromedp.SendKeys(`input[name='user']`, os.Getenv("GRAFANA_USER")),
			chromedp.SendKeys(`input[name='password']`, os.Getenv("GRAFANA_PASSWORD")),
			chromedp.Click(`button[type='submit']`),
		)
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Print(err)
	}

}

var ctx context.Context

func handler(w http.ResponseWriter, r *http.Request) {
	var res []byte
	chromedp.Run(ctx,
		chromedp.Navigate(os.Getenv("GRAFANA_DASHBOARD_URL")),
		chromedp.EmulateViewport(950, 540),
		chromedp.WaitVisible(`div.u-over`),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.CaptureScreenshot(&res),
	)
	imageP, _ := png.Decode(bytes.NewReader(res))
	dst := image.NewGray(imageP.Bounds())
	for y := 0; y < imageP.Bounds().Dy(); y++ {
		for x := 0; x < imageP.Bounds().Dx(); x++ {
			c := color.GrayModel.Convert(imageP.At(x, y))
			gray, _ := c.(color.Gray)
			if gray.Y > 200 {
				dst.Set(x, y, color.Gray{Y: 255})
			} else if gray.Y < 25 {
				dst.Set(x, y, color.Gray{Y: 0})
			} else {
				dst.Set(x, y, gray)
			}
		}
	}
	tmp := filepath.Join(os.TempDir(), "grafana.jpg")
	f, _ := os.Create(tmp)
	jpeg.Encode(f, dst, nil)
	http.ServeFile(w, r, tmp)
}

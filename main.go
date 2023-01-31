package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

	if err := srv.ListenAndServe(); err != nil {
		log.Print(err)
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	var res []byte
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true), // headless=false に変更
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	chromedp.Run(ctx,
		chromedp.Navigate(os.Getenv("GRAFANA_DASHBOARD_URL")),
		chromedp.Sleep(2*time.Second),
		chromedp.SendKeys(`input[name='user']`, os.Getenv("GRAFANA_USER")),
		chromedp.SendKeys(`input[name='password']`, os.Getenv("GRAFANA_PASSWORD")),
		chromedp.Click(`button[type='submit']`),
		chromedp.Sleep(5*time.Second),
		chromedp.EmulateViewport(540, 950),
		chromedp.Screenshot(`div.grafana-app`, &res),
	)
	b, err := io.Copy(w, bytes.NewReader(res))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Conetnt-Lengh", fmt.Sprint(b))
	w.Header().Add("Content-Type", "image/png")
	w.Header().Write(w)
}

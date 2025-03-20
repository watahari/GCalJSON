package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/patrickmn/go-cache"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// GCalJSON API - Swagger documentation
// @title GCalJSON API
// @version 1.0
// @description Google Calendar の情報をJSON形式で応答するAPI。Grafana のBusiness Calendar Plugin 用。
// @contact.name Your Name
// @contact.email your.email@example.com
// @host localhost:8080
// @BasePath /

type Event struct {
	Title       string `json:"title"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
}

// eventCache は Google Calendar API の結果をキャッシュします。
// キャッシュ期間は環境変数 GCALJSON_CACHE_DURATION (例:"5m") から設定します。
var eventCache *cache.Cache

// transformEvent は Google Calendar のイベントを GCalJSON 用の形式に変換します。
func transformEvent(item *calendar.Event) Event {
	start := item.Start.DateTime
	if start == "" {
		start = item.Start.Date
	}
	end := item.End.DateTime
	if end == "" {
		end = item.End.Date
	}
	return Event{
		Title:       item.Summary,
		Start:       start,
		End:         end,
		Description: item.Description,
		Location:    item.Location,
	}
}

// fetchEvents は Google Calendar API からイベントを取得し、キャッシュします。
func fetchEvents(srv *calendar.Service, calendarID string) ([]Event, error) {
	const cacheKey = "events"
	if cached, found := eventCache.Get(cacheKey); found {
		if events, ok := cached.([]Event); ok {
			return events, nil
		}
	}

	nowTime := time.Now()

	// 前月の初日を計算
	var prevMonth time.Month
	var prevYear int
	if nowTime.Month() == time.January {
		prevMonth = time.December
		prevYear = nowTime.Year() - 1
	} else {
		prevMonth = nowTime.Month() - 1
		prevYear = nowTime.Year()
	}
	tMinTime := time.Date(prevYear, prevMonth, 1, 0, 0, 0, 0, nowTime.Location())

	// 来月の最終日を計算
	// 今月 + 2ヶ月目の初日から1秒引くと、来月の最終日になる
	tMaxBase := time.Date(nowTime.Year(), nowTime.Month()+2, 1, 0, 0, 0, 0, nowTime.Location())
	tMaxTime := tMaxBase.Add(-time.Second)

	timeMin := tMinTime.Format(time.RFC3339)
	timeMax := tMaxTime.Format(time.RFC3339)

	eventsResult, err := srv.Events.List(calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMin).
		TimeMax(timeMax).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, err
	}

	var res []Event
	for _, item := range eventsResult.Items {
		res = append(res, transformEvent(item))
	}
	eventCache.Set(cacheKey, res, cache.DefaultExpiration)
	return res, nil
}

// errorResponse は詳細なエラーメッセージをログ出力しつつ、JSON レスポンスを返します。
func errorResponse(w http.ResponseWriter, code int, message string, err error) {
	log.Printf("Error: %s: %v", message, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// getEventsHandler は /events エンドポイントのハンドラです。
// @Summary Get calendar events
// @Description Google Calendar からイベントを取得し、Grafana のBusiness Calendar Plugin 用の形式で返します。
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {array} Event
// @Failure 500 {object} map[string]string "error message"
// @Router /events [get]
func getEventsHandler(srv *calendar.Service, calendarID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		events, err := fetchEvents(srv, calendarID)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to fetch events", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(events); err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to encode response", err)
		}
	}
}

func main() {
	// 環境変数は接頭辞 GCALJSON_ を利用
	// Base64エンコードされた認証情報をデコードして使用する
	encodedCred := os.Getenv("GCALJSON_GOOGLE_CREDENTIAL")
	calendarID := os.Getenv("GCALJSON_GOOGLE_CALENDAR_ID")
	cacheDurationStr := os.Getenv("GCALJSON_CACHE_DURATION")
	if encodedCred == "" || calendarID == "" {
		log.Fatal("GCALJSON_GOOGLE_CREDENTIAL と GCALJSON_GOOGLE_CALENDAR_ID を設定してください")
	}
	if cacheDurationStr == "" {
		cacheDurationStr = "5m"
	}
	cacheDuration, err := time.ParseDuration(cacheDurationStr)
	if err != nil {
		log.Fatalf("Invalid GCALJSON_CACHE_DURATION: %v", err)
	}
	// キャッシュの有効期間は環境変数から設定（クリーニング間隔は2倍の期間）
	eventCache = cache.New(cacheDuration, 2*cacheDuration)

	credJSON, err := base64.StdEncoding.DecodeString(encodedCred)
	if err != nil {
		log.Fatalf("Failed to decode credentials: %v", err)
	}

	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithCredentialsJSON(credJSON))
	if err != nil {
		log.Fatalf("Google Calendar サービスの作成に失敗: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/events", getEventsHandler(srv, calendarID))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// サーバーをゴルーチンで起動
	go func() {
		log.Println("GCalJSON API server started on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// SIGINT/SIGTERM を捕捉してグレースフルシャットダウン
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}

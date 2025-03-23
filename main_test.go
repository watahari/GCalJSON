package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"google.golang.org/api/calendar/v3"
)

func TestTransformEvent(t *testing.T) {
	// DateTime 指定の場合
	ev1 := &calendar.Event{
		Summary:     "Test Event",
		Description: "Test Description",
		Location:    "Test Location",
		Start:       &calendar.EventDateTime{DateTime: "2023-02-23T10:00:00Z"},
		End:         &calendar.EventDateTime{DateTime: "2023-02-23T11:00:00Z"},
	}
	res1 := transformEvent(ev1)
	if res1.Title != "Test Event" {
		t.Errorf("Expected title 'Test Event', got '%s'", res1.Title)
	}
	if res1.Start != "2023-02-23T10:00:00Z" {
		t.Errorf("Expected start '2023-02-23T10:00:00Z', got '%s'", res1.Start)
	}
	if res1.End != "2023-02-23T11:00:00Z" {
		t.Errorf("Expected end '2023-02-23T11:00:00Z', got '%s'", res1.End)
	}
	if res1.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", res1.Description)
	}
	if res1.Location != "Test Location" {
		t.Errorf("Expected location 'Test Location', got '%s'", res1.Location)
	}

	// All-day イベントの場合
	ev2 := &calendar.Event{
		Summary:     "All Day Event",
		Description: "",
		Location:    "",
		Start:       &calendar.EventDateTime{Date: "2023-02-24"},
		End:         &calendar.EventDateTime{Date: "2023-02-25"},
	}
	res2 := transformEvent(ev2)
	if res2.Start != "2023-02-24" {
		t.Errorf("Expected start '2023-02-24', got '%s'", res2.Start)
	}
	if res2.End != "2023-02-25" {
		t.Errorf("Expected end '2023-02-25', got '%s'", res2.End)
	}
	if res2.Description != "" {
		t.Errorf("Expected empty description, got '%s'", res2.Description)
	}
	if res2.Location != "" {
		t.Errorf("Expected empty location, got '%s'", res2.Location)
	}

	// すべての項目が空でも正しく動作することを確認
	ev3 := &calendar.Event{
		Summary:  "Empty Fields Event",
		Start:    &calendar.EventDateTime{DateTime: "2023-03-01T09:00:00Z"},
		End:      &calendar.EventDateTime{DateTime: "2023-03-01T10:00:00Z"},
	}
	res3 := transformEvent(ev3)
	if res3.Description != "" {
		t.Errorf("Expected empty description, got '%s'", res3.Description)
	}
	if res3.Location != "" {
		t.Errorf("Expected empty location, got '%s'", res3.Location)
	}
}

func TestErrorResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	errorResponse(rr, 500, "Test error", nil)
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}
	if resp["error"] != "Test error" {
		t.Errorf("Expected error message 'Test error', got '%s'", resp["error"])
	}
}

func TestGetEventsHandler(t *testing.T) {
	// テスト用にキャッシュを再初期化（デフォルト5分）
	eventCache = cache.New(5*time.Minute, 10*time.Minute)
	testEvents := []Event{
		{Title: "Dummy Event", Start: "2023-03-01T09:00:00Z", End: "2023-03-01T10:00:00Z"},
	}
	eventCache.Set("events", testEvents, cache.DefaultExpiration)

	req := httptest.NewRequest("GET", "/events", nil)
	rr := httptest.NewRecorder()

	// カレンダーサービスはキャッシュヒットを前提として nil でもOK
	handler := getEventsHandler(nil, "dummyCalendarID")
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	var events []Event
	if err := json.NewDecoder(rr.Body).Decode(&events); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].Title != "Dummy Event" {
		t.Errorf("Expected event title 'Dummy Event', got '%s'", events[0].Title)
	}
}

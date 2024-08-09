package main

import (
	"testing"
	"time"
)

func TestCalcWrkHrs(t *testing.T) {
	testCases := []struct {
		refDate string
		expectDays int
		expectErr bool
	}{
        {"2024-July-30", 4, false}, // Adjust according to today's date
        {"2024-July-28", 5, false},
        {"2024-July-01", 25, false},
        {"invalid-date", 0, true},
    }

	for _, tc := range testCases {
        t.Run(tc.refDate, func(t *testing.T) {
            done := make(chan bool)
            var days int
            var err error

            go func() {
                days, err = CalcWkDays(tc.refDate)
                done <- true
            }()

            select {
            case <-done:
                if (err != nil) != tc.expectErr {
                    t.Errorf("expected error: %v, got: %v", tc.expectErr, err)
                }
                if days != tc.expectDays {
                    t.Errorf("expected %d days, got %d days", tc.expectDays, days)
                }
            case <-time.After(1 * time.Second):
                t.Errorf("test for input date %s timed out", tc.refDate)
            }
        })
    }
}
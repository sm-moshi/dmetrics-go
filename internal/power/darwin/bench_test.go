//go:build darwin
// +build darwin

package darwin

import (
	"context"
	"testing"
	"time"
)

func BenchmarkGetStats(b *testing.B) {
	provider := NewProvider()
	ctx := b.Context()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats, err := provider.GetStats(ctx)
		if err != nil {
			b.Fatal(err)
		}
		if stats == nil {
			b.Fatal("stats should not be nil")
		}
	}
}

func BenchmarkGetPowerSource(b *testing.B) {
	provider := NewProvider()
	ctx := b.Context()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.GetPowerSource(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBatteryPercentage(b *testing.B) {
	provider := NewProvider()
	ctx := b.Context()

	// Skip if no battery
	if _, err := provider.GetBatteryPercentage(ctx); err != nil {
		b.Skip("No battery available")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.GetBatteryPercentage(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConcurrentAccess(b *testing.B) {
	provider := NewProvider()
	ctx := b.Context()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := provider.GetStats(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkWatch(b *testing.B) {
	provider := NewProvider()
	ctx := b.Context()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		ch, err := provider.Watch(ctx, 10*time.Millisecond)
		if err != nil {
			b.Fatal(err)
		}

		var count int
		for range ch {
			count++
		}
		cancel()

		if count == 0 {
			b.Fatal("expected at least one reading")
		}
	}
}

// Memory allocation benchmarks.
func BenchmarkGetStatsAlloc(b *testing.B) {
	provider := NewProvider()
	ctx := b.Context()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := provider.GetStats(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

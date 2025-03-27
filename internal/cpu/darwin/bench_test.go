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
	defer provider.Shutdown()
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

func BenchmarkGetCoreUsage(b *testing.B) {
	provider := NewProvider()
	defer provider.Shutdown()
	ctx := b.Context()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats, err := provider.GetStats(ctx)
		if err != nil {
			b.Fatal(err)
		}
		if len(stats.CoreUsage) == 0 {
			b.Fatal("expected core usage stats")
		}
	}
}

func BenchmarkGetFrequency(b *testing.B) {
	provider := NewProvider()
	defer provider.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		freq, err := provider.GetFrequency()
		if err != nil {
			b.Fatal(err)
		}
		if freq <= 0 {
			b.Fatal("expected positive frequency")
		}
	}
}

func BenchmarkConcurrentAccess(b *testing.B) {
	provider := NewProvider()
	defer provider.Shutdown()
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
	defer provider.Shutdown()
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

// Memory allocation benchmarks
func BenchmarkGetStatsAlloc(b *testing.B) {
	provider := NewProvider()
	defer provider.Shutdown()
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

func BenchmarkWatchAlloc(b *testing.B) {
	provider := NewProvider()
	defer provider.Shutdown()
	ctx := b.Context()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		ch, err := provider.Watch(ctx, 10*time.Millisecond)
		if err != nil {
			b.Fatal(err)
		}

		for range ch {
			// Consume updates
		}
		cancel()
	}
}

// Baseline measurement for CPU stats collection
func BenchmarkBaselineLatency(b *testing.B) {
	provider := NewProvider()
	defer provider.Shutdown()
	ctx := b.Context()

	// Warm up
	_, err := provider.GetStats(ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		_, err := provider.GetStats(ctx)
		if err != nil {
			b.Fatal(err)
		}
		b.SetBytes(1) // To calculate MB/s throughput
		b.ReportMetric(float64(time.Since(start).Nanoseconds()), "ns/op")
	}
}

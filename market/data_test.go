package market

import (
	"math"
	"testing"
)

// generateTestKlines generates test K-line data
func generateTestKlines(count int) []Kline {
	klines := make([]Kline, count)
	for i := 0; i < count; i++ {
		// Generate simulated price data with some fluctuation
		basePrice := 100.0
		variance := float64(i%10) * 0.5
		open := basePrice + variance
		high := open + 1.0
		low := open - 0.5
		close := open + 0.3
		volume := 1000.0 + float64(i*100)

		klines[i] = Kline{
			OpenTime:  int64(i * 180000), // 3-minute interval
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			CloseTime: int64((i+1)*180000 - 1),
		}
	}
	return klines
}

// TestCalculateIntradaySeries_VolumeCollection tests Volume data collection
func TestCalculateIntradaySeries_VolumeCollection(t *testing.T) {
	tests := []struct {
		name           string
		klineCount     int
		expectedVolLen int
	}{
		{
			name:           "Normal case - 20 K-lines",
			klineCount:     20,
			expectedVolLen: 10, // Should collect latest 10
		},
		{
			name:           "Exactly 10 K-lines",
			klineCount:     10,
			expectedVolLen: 10,
		},
		{
			name:           "Less than 10 K-lines",
			klineCount:     5,
			expectedVolLen: 5, // Should return all 5
		},
		{
			name:           "More than 10 K-lines",
			klineCount:     30,
			expectedVolLen: 10, // Should only return latest 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			klines := generateTestKlines(tt.klineCount)
			data := calculateIntradaySeries(klines)

			if data == nil {
				t.Fatal("calculateIntradaySeries returned nil")
			}

			if len(data.Volume) != tt.expectedVolLen {
				t.Errorf("Volume length = %d, want %d", len(data.Volume), tt.expectedVolLen)
			}

			// Verify Volume data correctness
			if len(data.Volume) > 0 {
				// Calculate expected start index
				start := tt.klineCount - 10
				if start < 0 {
					start = 0
				}

				// Verify first Volume value
				expectedFirstVolume := klines[start].Volume
				if data.Volume[0] != expectedFirstVolume {
					t.Errorf("First volume = %.2f, want %.2f", data.Volume[0], expectedFirstVolume)
				}

				// Verify last Volume value
				expectedLastVolume := klines[tt.klineCount-1].Volume
				lastVolume := data.Volume[len(data.Volume)-1]
				if lastVolume != expectedLastVolume {
					t.Errorf("Last volume = %.2f, want %.2f", lastVolume, expectedLastVolume)
				}
			}
		})
	}
}

// TestCalculateIntradaySeries_VolumeValues tests Volume value correctness
func TestCalculateIntradaySeries_VolumeValues(t *testing.T) {
	klines := []Kline{
		{Close: 100.0, Volume: 1000.0, High: 101.0, Low: 99.0, Open: 100.0},
		{Close: 101.0, Volume: 1100.0, High: 102.0, Low: 100.0, Open: 101.0},
		{Close: 102.0, Volume: 1200.0, High: 103.0, Low: 101.0, Open: 102.0},
		{Close: 103.0, Volume: 1300.0, High: 104.0, Low: 102.0, Open: 103.0},
		{Close: 104.0, Volume: 1400.0, High: 105.0, Low: 103.0, Open: 104.0},
		{Close: 105.0, Volume: 1500.0, High: 106.0, Low: 104.0, Open: 105.0},
		{Close: 106.0, Volume: 1600.0, High: 107.0, Low: 105.0, Open: 106.0},
		{Close: 107.0, Volume: 1700.0, High: 108.0, Low: 106.0, Open: 107.0},
		{Close: 108.0, Volume: 1800.0, High: 109.0, Low: 107.0, Open: 108.0},
		{Close: 109.0, Volume: 1900.0, High: 110.0, Low: 108.0, Open: 109.0},
	}

	data := calculateIntradaySeries(klines)

	expectedVolumes := []float64{1000.0, 1100.0, 1200.0, 1300.0, 1400.0, 1500.0, 1600.0, 1700.0, 1800.0, 1900.0}

	if len(data.Volume) != len(expectedVolumes) {
		t.Fatalf("Volume length = %d, want %d", len(data.Volume), len(expectedVolumes))
	}

	for i, expected := range expectedVolumes {
		if data.Volume[i] != expected {
			t.Errorf("Volume[%d] = %.2f, want %.2f", i, data.Volume[i], expected)
		}
	}
}

// TestCalculateIntradaySeries_ATR14 tests ATR14 calculation
func TestCalculateIntradaySeries_ATR14(t *testing.T) {
	tests := []struct {
		name          string
		klineCount    int
		expectZero    bool
		expectNonZero bool
	}{
		{
			name:          "Sufficient data - 20 K-lines",
			klineCount:    20,
			expectNonZero: true,
		},
		{
			name:          "Exactly 15 K-lines (ATR14 requires at least 15)",
			klineCount:    15,
			expectNonZero: true,
		},
		{
			name:       "Insufficient data - 14 K-lines",
			klineCount: 14,
			expectZero: true,
		},
		{
			name:       "Insufficient data - 10 K-lines",
			klineCount: 10,
			expectZero: true,
		},
		{
			name:       "Insufficient data - 5 K-lines",
			klineCount: 5,
			expectZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			klines := generateTestKlines(tt.klineCount)
			data := calculateIntradaySeries(klines)

			if data == nil {
				t.Fatal("calculateIntradaySeries returned nil")
			}

			if tt.expectZero && data.ATR14 != 0 {
				t.Errorf("ATR14 = %.3f, expected 0 (insufficient data)", data.ATR14)
			}

			if tt.expectNonZero && data.ATR14 <= 0 {
				t.Errorf("ATR14 = %.3f, expected > 0", data.ATR14)
			}
		})
	}
}

// TestCalculateATR tests ATR calculation function
func TestCalculateATR(t *testing.T) {
	tests := []struct {
		name       string
		klines     []Kline
		period     int
		expectZero bool
	}{
		{
			name: "Normal calculation - sufficient data",
			klines: []Kline{
				{High: 102.0, Low: 100.0, Close: 101.0},
				{High: 103.0, Low: 101.0, Close: 102.0},
				{High: 104.0, Low: 102.0, Close: 103.0},
				{High: 105.0, Low: 103.0, Close: 104.0},
				{High: 106.0, Low: 104.0, Close: 105.0},
				{High: 107.0, Low: 105.0, Close: 106.0},
				{High: 108.0, Low: 106.0, Close: 107.0},
				{High: 109.0, Low: 107.0, Close: 108.0},
				{High: 110.0, Low: 108.0, Close: 109.0},
				{High: 111.0, Low: 109.0, Close: 110.0},
				{High: 112.0, Low: 110.0, Close: 111.0},
				{High: 113.0, Low: 111.0, Close: 112.0},
				{High: 114.0, Low: 112.0, Close: 113.0},
				{High: 115.0, Low: 113.0, Close: 114.0},
				{High: 116.0, Low: 114.0, Close: 115.0},
			},
			period:     14,
			expectZero: false,
		},
		{
			name: "Insufficient data - equal to period",
			klines: []Kline{
				{High: 102.0, Low: 100.0, Close: 101.0},
				{High: 103.0, Low: 101.0, Close: 102.0},
			},
			period:     2,
			expectZero: true,
		},
		{
			name: "Insufficient data - less than period",
			klines: []Kline{
				{High: 102.0, Low: 100.0, Close: 101.0},
			},
			period:     14,
			expectZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atr := calculateATR(tt.klines, tt.period)

			if tt.expectZero {
				if atr != 0 {
					t.Errorf("calculateATR() = %.3f, expected 0 (insufficient data)", atr)
				}
			} else {
				if atr <= 0 {
					t.Errorf("calculateATR() = %.3f, expected > 0", atr)
				}
			}
		})
	}
}

// TestCalculateATR_TrueRange tests ATR True Range calculation correctness
func TestCalculateATR_TrueRange(t *testing.T) {
	// Create a simple test case, manually calculate expected ATR
	klines := []Kline{
		{High: 50.0, Low: 48.0, Close: 49.0}, // TR = 2.0
		{High: 51.0, Low: 49.0, Close: 50.0}, // TR = max(2.0, 2.0, 1.0) = 2.0
		{High: 52.0, Low: 50.0, Close: 51.0}, // TR = max(2.0, 2.0, 1.0) = 2.0
		{High: 53.0, Low: 51.0, Close: 52.0}, // TR = 2.0
		{High: 54.0, Low: 52.0, Close: 53.0}, // TR = 2.0
	}

	atr := calculateATR(klines, 3)

	// Expected calculation:
	// TR[1] = max(51-49, |51-49|, |49-49|) = 2.0
	// TR[2] = max(52-50, |52-50|, |50-50|) = 2.0
	// TR[3] = max(53-51, |53-51|, |51-51|) = 2.0
	// Initial ATR = (2.0 + 2.0 + 2.0) / 3 = 2.0
	// TR[4] = max(54-52, |54-52|, |52-52|) = 2.0
	// Smoothed ATR = (2.0*2 + 2.0) / 3 = 2.0

	expectedATR := 2.0
	tolerance := 0.01 // Allow small floating point error

	if math.Abs(atr-expectedATR) > tolerance {
		t.Errorf("calculateATR() = %.3f, want approximately %.3f", atr, expectedATR)
	}
}

// TestCalculateIntradaySeries_ConsistencyWithOtherIndicators tests Volume and other indicators consistency
func TestCalculateIntradaySeries_ConsistencyWithOtherIndicators(t *testing.T) {
	klines := generateTestKlines(30)
	data := calculateIntradaySeries(klines)

	// All arrays should exist
	if data.MidPrices == nil {
		t.Error("MidPrices should not be nil")
	}
	if data.Volume == nil {
		t.Error("Volume should not be nil")
	}

	// MidPrices and Volume should have the same length (both latest 10)
	if len(data.MidPrices) != len(data.Volume) {
		t.Errorf("MidPrices length (%d) should equal Volume length (%d)",
			len(data.MidPrices), len(data.Volume))
	}

	// All Volume values should be > 0
	for i, vol := range data.Volume {
		if vol <= 0 {
			t.Errorf("Volume[%d] = %.2f, should be > 0", i, vol)
		}
	}
}

// TestCalculateIntradaySeries_EmptyKlines tests empty K-line data
func TestCalculateIntradaySeries_EmptyKlines(t *testing.T) {
	klines := []Kline{}
	data := calculateIntradaySeries(klines)

	if data == nil {
		t.Fatal("calculateIntradaySeries should not return nil for empty klines")
	}

	// All slices should be empty
	if len(data.MidPrices) != 0 {
		t.Errorf("MidPrices length = %d, want 0", len(data.MidPrices))
	}
	if len(data.Volume) != 0 {
		t.Errorf("Volume length = %d, want 0", len(data.Volume))
	}

	// ATR14 should be 0 (insufficient data)
	if data.ATR14 != 0 {
		t.Errorf("ATR14 = %.3f, want 0", data.ATR14)
	}
}

// TestCalculateIntradaySeries_VolumePrecision tests Volume precision preservation
func TestCalculateIntradaySeries_VolumePrecision(t *testing.T) {
	klines := []Kline{
		{Close: 100.0, Volume: 1234.5678, High: 101.0, Low: 99.0},
		{Close: 101.0, Volume: 9876.5432, High: 102.0, Low: 100.0},
		{Close: 102.0, Volume: 5555.1111, High: 103.0, Low: 101.0},
	}

	data := calculateIntradaySeries(klines)

	expectedVolumes := []float64{1234.5678, 9876.5432, 5555.1111}

	for i, expected := range expectedVolumes {
		if data.Volume[i] != expected {
			t.Errorf("Volume[%d] = %.4f, want %.4f (precision not preserved)",
				i, data.Volume[i], expected)
		}
	}
}

// TestIsStaleData_NormalData tests that normal fluctuating data returns false
func TestIsStaleData_NormalData(t *testing.T) {
	klines := []Kline{
		{Close: 100.0, Volume: 1000},
		{Close: 100.5, Volume: 1200},
		{Close: 99.8, Volume: 900},
		{Close: 100.2, Volume: 1100},
		{Close: 100.1, Volume: 950},
	}

	result := isStaleData(klines, "BTCUSDT")

	if result {
		t.Error("Expected false for normal fluctuating data, got true")
	}
}

// TestIsStaleData_PriceFreezeWithZeroVolume tests that frozen price + zero volume returns true
func TestIsStaleData_PriceFreezeWithZeroVolume(t *testing.T) {
	klines := []Kline{
		{Close: 100.0, Volume: 0},
		{Close: 100.0, Volume: 0},
		{Close: 100.0, Volume: 0},
		{Close: 100.0, Volume: 0},
		{Close: 100.0, Volume: 0},
	}

	result := isStaleData(klines, "DOGEUSDT")

	if !result {
		t.Error("Expected true for frozen price + zero volume, got false")
	}
}

// TestIsStaleData_PriceFreezeWithVolume tests that frozen price but normal volume returns false
func TestIsStaleData_PriceFreezeWithVolume(t *testing.T) {
	klines := []Kline{
		{Close: 100.0, Volume: 1000},
		{Close: 100.0, Volume: 1200},
		{Close: 100.0, Volume: 900},
		{Close: 100.0, Volume: 1100},
		{Close: 100.0, Volume: 950},
	}

	result := isStaleData(klines, "STABLECOIN")

	if result {
		t.Error("Expected false for frozen price but normal volume (low volatility market), got true")
	}
}

// TestIsStaleData_InsufficientData tests that insufficient data (<5 klines) returns false
func TestIsStaleData_InsufficientData(t *testing.T) {
	klines := []Kline{
		{Close: 100.0, Volume: 0},
		{Close: 100.0, Volume: 0},
		{Close: 100.0, Volume: 0},
	}

	result := isStaleData(klines, "BTCUSDT")

	if result {
		t.Error("Expected false for insufficient data (<5 klines), got true")
	}
}

// TestIsStaleData_ExactlyFiveKlines tests edge case with exactly 5 klines
func TestIsStaleData_ExactlyFiveKlines(t *testing.T) {
	// Stale case: exactly 5 frozen klines with zero volume
	staleKlines := []Kline{
		{Close: 100.0, Volume: 0},
		{Close: 100.0, Volume: 0},
		{Close: 100.0, Volume: 0},
		{Close: 100.0, Volume: 0},
		{Close: 100.0, Volume: 0},
	}

	result := isStaleData(staleKlines, "TESTUSDT")
	if !result {
		t.Error("Expected true for exactly 5 frozen klines with zero volume, got false")
	}

	// Normal case: exactly 5 klines with fluctuation
	normalKlines := []Kline{
		{Close: 100.0, Volume: 1000},
		{Close: 100.1, Volume: 1100},
		{Close: 99.9, Volume: 900},
		{Close: 100.0, Volume: 1000},
		{Close: 100.05, Volume: 950},
	}

	result = isStaleData(normalKlines, "TESTUSDT")
	if result {
		t.Error("Expected false for exactly 5 normal klines, got true")
	}
}

// TestIsStaleData_WithinTolerance tests price changes within tolerance (0.01%)
func TestIsStaleData_WithinTolerance(t *testing.T) {
	// Price changes within 0.01% tolerance should be treated as frozen
	basePrice := 10000.0
	tolerance := 0.0001                        // 0.01%
	smallChange := basePrice * tolerance * 0.5 // Half of tolerance

	klines := []Kline{
		{Close: basePrice, Volume: 1000},
		{Close: basePrice + smallChange, Volume: 1000},
		{Close: basePrice - smallChange, Volume: 1000},
		{Close: basePrice, Volume: 1000},
		{Close: basePrice + smallChange, Volume: 1000},
	}

	result := isStaleData(klines, "BTCUSDT")

	// Should return false because there's normal volume despite tiny price changes
	if result {
		t.Error("Expected false for price within tolerance but with volume, got true")
	}
}

// TestIsStaleData_MixedScenario tests realistic scenario with some history before freeze
func TestIsStaleData_MixedScenario(t *testing.T) {
	// Simulate: normal trading → suddenly freezes
	klines := []Kline{
		{Close: 100.0, Volume: 1000}, // Normal
		{Close: 100.5, Volume: 1200}, // Normal
		{Close: 100.2, Volume: 1100}, // Normal
		{Close: 50.0, Volume: 0},     // Freeze starts
		{Close: 50.0, Volume: 0},     // Frozen
		{Close: 50.0, Volume: 0},     // Frozen
		{Close: 50.0, Volume: 0},     // Frozen
		{Close: 50.0, Volume: 0},     // Frozen (last 5 are all frozen)
	}

	result := isStaleData(klines, "DOGEUSDT")

	// Should detect stale data based on last 5 klines
	if !result {
		t.Error("Expected true for frozen last 5 klines with zero volume, got false")
	}
}

// TestIsStaleData_EmptyKlines tests edge case with empty slice
func TestIsStaleData_EmptyKlines(t *testing.T) {
	klines := []Kline{}

	result := isStaleData(klines, "BTCUSDT")

	if result {
		t.Error("Expected false for empty klines, got true")
	}
}

func TestCalculateDonchian(t *testing.T) {
	// Create test klines with known high/low values
	klines := []Kline{
		{High: 100, Low: 90},
		{High: 105, Low: 88},
		{High: 102, Low: 92},
		{High: 108, Low: 85},
		{High: 103, Low: 91},
	}

	upper, lower := ExportCalculateDonchian(klines, 5)

	if upper != 108 {
		t.Errorf("Expected upper = 108, got %v", upper)
	}
	if lower != 85 {
		t.Errorf("Expected lower = 85, got %v", lower)
	}
}

func TestCalculateDonchian_PartialPeriod(t *testing.T) {
	klines := []Kline{
		{High: 100, Low: 90},
		{High: 105, Low: 88},
	}

	upper, lower := ExportCalculateDonchian(klines, 10)

	// Should use all available klines when period > len(klines)
	if upper != 105 {
		t.Errorf("Expected upper = 105, got %v", upper)
	}
	if lower != 88 {
		t.Errorf("Expected lower = 88, got %v", lower)
	}
}

func TestCalculateDonchian_InvalidPeriod(t *testing.T) {
	klines := []Kline{
		{High: 100, Low: 90},
	}

	// Zero period should return (0, 0)
	upper, lower := ExportCalculateDonchian(klines, 0)
	if upper != 0 || lower != 0 {
		t.Errorf("Expected (0, 0) for zero period, got (%v, %v)", upper, lower)
	}

	// Negative period should return (0, 0)
	upper, lower = ExportCalculateDonchian(klines, -1)
	if upper != 0 || lower != 0 {
		t.Errorf("Expected (0, 0) for negative period, got (%v, %v)", upper, lower)
	}
}

func TestCalculateBoxData(t *testing.T) {
	// Create synthetic kline data
	klines := make([]Kline, 500)
	for i := 0; i < 500; i++ {
		basePrice := 100.0
		klines[i] = Kline{
			High:  basePrice + float64(i%10),
			Low:   basePrice - float64(i%10),
			Close: basePrice,
		}
	}

	box := ExportCalculateBoxData(klines, 100.0)

	if box.ShortUpper == 0 || box.ShortLower == 0 {
		t.Error("Short box should not be zero")
	}
	if box.MidUpper == 0 || box.MidLower == 0 {
		t.Error("Mid box should not be zero")
	}
	if box.LongUpper == 0 || box.LongLower == 0 {
		t.Error("Long box should not be zero")
	}
	if box.CurrentPrice != 100.0 {
		t.Errorf("Expected CurrentPrice = 100.0, got %v", box.CurrentPrice)
	}
}

// TestDetectPivots tests swing high/low detection (depth 1: left=1, right=1).
func TestDetectPivots(t *testing.T) {
	// Series: bar 1 has local max high (100) and local min low (98)
	klines := []Kline{
		{High: 99, Low: 99, Close: 98},
		{High: 100, Low: 98, Close: 99}, // swing high 100, swing low 98 at bar 1
		{High: 99, Low: 99, Close: 98},
	}
	highs, lows := detectPivots(klines, 1, 1)
	if len(highs) != 1 {
		t.Fatalf("expected 1 swing high, got %d", len(highs))
	}
	if highs[0].Price != 100 || highs[0].BarIndex != 1 {
		t.Errorf("swing high: want Price=100 BarIndex=1, got Price=%.2f BarIndex=%d", highs[0].Price, highs[0].BarIndex)
	}
	if len(lows) != 1 {
		t.Fatalf("expected 1 swing low, got %d", len(lows))
	}
	if lows[0].Price != 98 || lows[0].BarIndex != 1 {
		t.Errorf("swing low: want Price=98 BarIndex=1, got Price=%.2f BarIndex=%d", lows[0].Price, lows[0].BarIndex)
	}
}

// TestCalculateStructure tests Structure calculation: sweep and break events.
func TestCalculateStructure(t *testing.T) {
	// Build klines: swing high at bar 2 = 105. Bar 4 sweeps it (high 106, close 104). Bar 5 breaks (close 106).
	klines := []Kline{
		{OpenTime: 0, High: 100, Low: 99, Close: 99.5},
		{OpenTime: 1, High: 102, Low: 99, Close: 101},
		{OpenTime: 2, High: 105, Low: 101, Close: 103}, // swing high 105 at bar 2
		{OpenTime: 3, High: 104, Low: 102, Close: 103},
		{OpenTime: 4, High: 106, Low: 103, Close: 104}, // sweep: high > 105, close < 105
		{OpenTime: 5, High: 107, Low: 104, Close: 106}, // break: close > 105
	}
	opts := StructureOpts{Enabled: true, Depth: 1, Lookback: 100, MaxEvents: 20}
	data := CalculateStructure(klines, opts)
	if data == nil {
		t.Fatal("CalculateStructure returned nil")
	}
	// Should have at least one sweep_bear and one break (bos_bull or choch_bull)
	var hasSweepBear, hasBreak bool
	for _, ev := range data.Events {
		if ev.Type == "sweep_bear" {
			hasSweepBear = true
		}
		if ev.Type == "bos_bull" || ev.Type == "choch_bull" {
			hasBreak = true
		}
	}
	if !hasSweepBear {
		t.Error("expected at least one sweep_bear event")
	}
	if !hasBreak {
		t.Error("expected at least one bos_bull or choch_bull event")
	}
	if data.LastTrend != "bullish" {
		t.Errorf("expected LastTrend bullish after break, got %q", data.LastTrend)
	}
}

// TestCalculateStructure_Disabled verifies that opts.Enabled=false returns nil.
func TestCalculateStructure_Disabled(t *testing.T) {
	klines := generateTestKlines(50)
	opts := StructureOpts{Enabled: false, Lookback: 500}
	data := CalculateStructure(klines, opts)
	if data != nil {
		t.Error("expected nil when Enabled=false")
	}
}

// TestDetectPivots_EmptyAndShort verifies empty or too-short klines return nil.
func TestDetectPivots_EmptyAndShort(t *testing.T) {
	if highs, lows := detectPivots(nil, 1, 1); highs != nil || lows != nil {
		t.Error("nil klines should return nil, nil")
	}
	if highs, lows := detectPivots([]Kline{{High: 1, Low: 0}}, 1, 1); highs != nil || lows != nil {
		t.Error("single bar should return nil (need 3 for depth 1)")
	}
	klines2 := []Kline{{High: 1, Low: 0}, {High: 2, Low: 1}}
	if highs, lows := detectPivots(klines2, 1, 1); highs != nil || lows != nil {
		t.Error("two bars should return nil for depth 1")
	}
}

// TestDetectPivots_Depth2 verifies pivot detection with leftBars=2, rightBars=2.
func TestDetectPivots_Depth2(t *testing.T) {
	// Need at least 5 bars. Bar 2 is local max if high[0,1,2,3,4] has max at 2.
	klines := []Kline{
		{High: 101, Low: 100},
		{High: 102, Low: 101},
		{High: 105, Low: 102}, // swing high at bar 2 (need 2 left, 2 right)
		{High: 103, Low: 101},
		{High: 104, Low: 100},
	}
	highs, _ := detectPivots(klines, 2, 2)
	if len(highs) != 1 {
		t.Fatalf("expected 1 swing high with depth 2, got %d", len(highs))
	}
	if highs[0].BarIndex != 2 || highs[0].Price != 105 {
		t.Errorf("swing high: want BarIndex=2 Price=105, got BarIndex=%d Price=%.2f", highs[0].BarIndex, highs[0].Price)
	}
}

// TestDetectPivots_MultiplePivots verifies multiple swing highs in one series.
func TestDetectPivots_MultiplePivots(t *testing.T) {
	klines := []Kline{
		{High: 98, Low: 97},
		{High: 100, Low: 98}, // swing high 100
		{High: 99, Low: 97},
		{High: 99, Low: 96},
		{High: 101, Low: 99}, // swing high 101
		{High: 100, Low: 98},
	}
	highs, _ := detectPivots(klines, 1, 1)
	if len(highs) != 2 {
		t.Fatalf("expected 2 swing highs, got %d", len(highs))
	}
	if highs[0].Price != 100 || highs[0].BarIndex != 1 {
		t.Errorf("first swing high: want 100 @ 1, got %.2f @ %d", highs[0].Price, highs[0].BarIndex)
	}
	if highs[1].Price != 101 || highs[1].BarIndex != 4 {
		t.Errorf("second swing high: want 101 @ 4, got %.2f @ %d", highs[1].Price, highs[1].BarIndex)
	}
}

// TestCalculateStructure_TooFewKlines verifies that fewer than 3 klines returns nil.
func TestCalculateStructure_TooFewKlines(t *testing.T) {
	opts := StructureOpts{Enabled: true, Depth: 1, Lookback: 500}
	if data := CalculateStructure(nil, opts); data != nil {
		t.Error("nil klines should return nil")
	}
	if data := CalculateStructure([]Kline{}, opts); data != nil {
		t.Error("empty klines should return nil")
	}
	if data := CalculateStructure([]Kline{{High: 1, Low: 0}, {High: 2, Low: 1}}, opts); data != nil {
		t.Error("2 klines should return nil")
	}
}

// TestCalculateStructure_ChochBear verifies bearish ChoCH (break below swing low after uptrend).
func TestCalculateStructure_ChochBear(t *testing.T) {
	// Bar 2 = swing high 105. Bar 5 close 106 > 105 → bos_bull. Only one swing low at bar 6 = 98 (bars 3,4 have higher lows so no extra pivots). Bar 8 close < 98 → choch_bear.
	klines := []Kline{
		{OpenTime: 0, High: 100, Low: 99, Close: 99.5},
		{OpenTime: 1, High: 102, Low: 99, Close: 101},
		{OpenTime: 2, High: 105, Low: 101, Close: 104}, // swing high 105 at bar 2
		{OpenTime: 3, High: 104, Low: 103, Close: 103},
		{OpenTime: 4, High: 104, Low: 103, Close: 103},
		{OpenTime: 5, High: 106, Low: 104, Close: 106}, // close > 105 → bos_bull
		{OpenTime: 6, High: 104, Low: 98, Close: 100},  // swing low 98 at bar 6 (103, 98, 99)
		{OpenTime: 7, High: 100, Low: 99, Close: 99},
		{OpenTime: 8, High: 99, Low: 97, Close: 97},    // close < 98 → choch_bear
	}
	opts := StructureOpts{Enabled: true, Depth: 1, Lookback: 100, MaxEvents: 20}
	data := CalculateStructure(klines, opts)
	if data == nil {
		t.Fatal("CalculateStructure returned nil")
	}
	var hasChochBear bool
	for _, ev := range data.Events {
		if ev.Type == "choch_bear" {
			hasChochBear = true
			break
		}
	}
	if !hasChochBear {
		t.Error("expected choch_bear event after break below swing low in uptrend")
	}
	if data.LastTrend != "bearish" {
		t.Errorf("expected LastTrend bearish, got %q", data.LastTrend)
	}
}

// TestCalculateStructure_SweepBull verifies bullish sweep (wick below swing low, close above).
func TestCalculateStructure_SweepBull(t *testing.T) {
	klines := []Kline{
		{OpenTime: 0, High: 102, Low: 100, Close: 101},
		{OpenTime: 1, High: 103, Low: 99, Close: 102},
		{OpenTime: 2, High: 101, Low: 97, Close: 99},  // swing low 97 at bar 2
		{OpenTime: 3, High: 99, Low: 98, Close: 98.5},
		{OpenTime: 4, High: 99, Low: 96, Close: 98}, // low < 97, close > 97 → sweep_bull
	}
	opts := StructureOpts{Enabled: true, Depth: 1, Lookback: 100, MaxEvents: 20}
	data := CalculateStructure(klines, opts)
	if data == nil {
		t.Fatal("CalculateStructure returned nil")
	}
	var hasSweepBull bool
	for _, ev := range data.Events {
		if ev.Type == "sweep_bull" {
			hasSweepBull = true
			if ev.Level != 97 {
				t.Errorf("sweep_bull level want 97, got %.2f", ev.Level)
			}
			break
		}
	}
	if !hasSweepBull {
		t.Error("expected sweep_bull event")
	}
}

// TestCalculateStructure_MaxEventsCapped verifies that only last MaxEvents events are kept.
func TestCalculateStructure_MaxEventsCapped(t *testing.T) {
	// Create many swing points and breaks so we get more than MaxEvents.
	klines := make([]Kline, 0, 50)
	for i := 0; i < 50; i++ {
		var high, low, close float64
		if i%4 == 1 {
			high = 100 + float64(i)
			low = 99 + float64(i)
			close = 99.5 + float64(i)
		} else if i%4 == 3 {
			high = 101 + float64(i)
			low = 98 + float64(i)
			close = 100 + float64(i) // break above previous high
		} else {
			high = 99 + float64(i)
			low = 98 + float64(i)
			close = 98.5 + float64(i)
		}
		klines = append(klines, Kline{OpenTime: int64(i), High: high, Low: low, Close: close})
	}
	opts := StructureOpts{Enabled: true, Depth: 1, Lookback: 500, MaxEvents: 5}
	data := CalculateStructure(klines, opts)
	if data == nil {
		t.Fatal("CalculateStructure returned nil")
	}
	if len(data.Events) > 5 {
		t.Errorf("expected at most 5 events, got %d", len(data.Events))
	}
}

// TestCalculateStructure_Depth2 uses depth 2 (fewer pivots).
func TestCalculateStructure_Depth2(t *testing.T) {
	// With depth 2 we need 2 bars on each side; so first pivot possible at bar 2, need 5 bars min.
	klines := []Kline{
		{OpenTime: 0, High: 100, Low: 99, Close: 99.5},
		{OpenTime: 1, High: 101, Low: 99, Close: 100},
		{OpenTime: 2, High: 105, Low: 101, Close: 103}, // swing high 105 (depth 2)
		{OpenTime: 3, High: 103, Low: 102, Close: 102.5},
		{OpenTime: 4, High: 104, Low: 102, Close: 103},
		{OpenTime: 5, High: 106, Low: 103, Close: 106}, // close > 105 → break
	}
	opts := StructureOpts{Enabled: true, Depth: 2, Lookback: 100, MaxEvents: 10}
	data := CalculateStructure(klines, opts)
	if data == nil {
		t.Fatal("CalculateStructure returned nil")
	}
	// Should have at least one break (close > swing high 105)
	var hasBreak bool
	for _, ev := range data.Events {
		if ev.Type == "bos_bull" || ev.Type == "choch_bull" {
			hasBreak = true
			break
		}
	}
	if !hasBreak {
		t.Error("expected break event with depth 2")
	}
}

// TestCalculateStructure_DefaultLookbackMaxEvents verifies zero lookback/maxEvents use defaults.
func TestCalculateStructure_DefaultLookbackMaxEvents(t *testing.T) {
	klines := []Kline{
		{OpenTime: 0, High: 99, Low: 98, Close: 98.5},
		{OpenTime: 1, High: 100, Low: 98, Close: 99},
		{OpenTime: 2, High: 101, Low: 99, Close: 100},
		{OpenTime: 3, High: 102, Low: 100, Close: 102}, // break above 101
	}
	opts := StructureOpts{Enabled: true, Depth: 1, Lookback: 0, MaxEvents: 0}
	data := CalculateStructure(klines, opts)
	if data == nil {
		t.Fatal("CalculateStructure returned nil with default opts")
	}
	// Should not panic and should return valid structure (lookback 500, maxEvents 15 applied internally)
	if len(data.Events) > 15 {
		t.Errorf("events should be capped at default 15, got %d", len(data.Events))
	}
}

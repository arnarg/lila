package util

import "testing"

func TestConvertBytes(t *testing.T) {
	tests := []struct {
		name      string
		inBytes   int64
		outAmount float64
		outUnit   string
	}{
		{
			name:      "unchanged below KiB",
			inBytes:   1000,
			outAmount: 1000.0,
			outUnit:   BytesUnitBytes,
		},
		{
			name:      "above KiB correct",
			inBytes:   1024,
			outAmount: 1.0,
			outUnit:   BytesUnitKiB,
		},
		{
			name:      "above MiB correct",
			inBytes:   1024 * 1024,
			outAmount: 1.0,
			outUnit:   BytesUnitMiB,
		},
		{
			name:      "above GiB correct",
			inBytes:   1024 * 1024 * 1024,
			outAmount: 1.0,
			outUnit:   BytesUnitGiB,
		},
		{
			name:      "above TiB correct",
			inBytes:   1024 * 1024 * 1024 * 1024,
			outAmount: 1.0,
			outUnit:   BytesUnitTiB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, unit := ConvertBytes(tt.inBytes)

			if out != tt.outAmount {
				t.Errorf("converted amount is '%f' but '%f' was expected", out, tt.outAmount)
			}

			if unit != tt.outUnit {
				t.Errorf("converted unit is '%s' but '%s' was expected", unit, tt.outUnit)
			}
		})
	}
}

func TestConvertBytesToUnit(t *testing.T) {
	tests := []struct {
		name      string
		inBytes   int64
		inUnit    string
		outAmount float64
	}{
		{
			name:      "unchanged with KiB",
			inBytes:   1000,
			inUnit:    BytesUnitBytes,
			outAmount: 1000.0,
		},
		{
			name:      "with KiB correct",
			inBytes:   512,
			inUnit:    BytesUnitKiB,
			outAmount: 0.5,
		},
		{
			name:      "with MiB correct",
			inBytes:   512 * 1024,
			inUnit:    BytesUnitMiB,
			outAmount: 0.5,
		},
		{
			name:      "with GiB correct",
			inBytes:   512 * 1024 * 1024,
			inUnit:    BytesUnitGiB,
			outAmount: 0.5,
		},
		{
			name:      "with TiB correct",
			inBytes:   512 * 1024 * 1024 * 1024,
			inUnit:    BytesUnitTiB,
			outAmount: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := ConvertBytesToUnit(tt.inBytes, tt.inUnit)

			if out != tt.outAmount {
				t.Errorf("converted amount is '%f' but '%f' was expected", out, tt.outAmount)
			}
		})
	}
}

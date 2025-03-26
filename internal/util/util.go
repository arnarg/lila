package util

const (
	_ = 1 << (10 * iota)
	KiB
	MiB
	GiB
	TiB
)

const (
	BytesUnitBytes = "B"
	BytesUnitKiB   = "KiB"
	BytesUnitMiB   = "MiB"
	BytesUnitGiB   = "GiB"
	BytesUnitTiB   = "TiB"
)

func ConvertBytes(b int64) (float64, string) {
	if b >= TiB {
		return float64(b) / TiB, BytesUnitTiB
	} else if b >= GiB {
		return float64(b) / GiB, BytesUnitGiB
	} else if b >= MiB {
		return float64(b) / MiB, BytesUnitMiB
	} else if b >= KiB {
		return float64(b) / KiB, BytesUnitKiB
	}
	return float64(b), BytesUnitBytes
}

func ConvertBytesToUnit(b int64, unit string) float64 {
	switch unit {
	case BytesUnitTiB:
		return float64(b) / TiB
	case BytesUnitGiB:
		return float64(b) / GiB
	case BytesUnitMiB:
		return float64(b) / MiB
	case BytesUnitKiB:
		return float64(b) / KiB
	default:
		return float64(b)
	}
}

package helper

type PartSize int64

const (
	SizeBytes     PartSize = 1
	SizeKiloBytes          = 1024 * SizeBytes
	SizeMegaBytes          = 1024 * SizeKiloBytes
	SizeGigaBytes          = 1024 * SizeMegaBytes
	SizeTeraBytes          = 1024 * SizeGigaBytes
	SizePetaBytes          = 1024 * SizeTeraBytes
)

func SetSize(val int) PartSize {
	return PartSize(val)
}

func SetSizeInt64(val int64) PartSize {
	return PartSize(val)
}

func (p PartSize) Int() int {
	return int(p)
}

func (p PartSize) Int64() int64 {
	return int64(p)
}

func (p PartSize) String() string {
	switch p {
	case SizePetaBytes:
		return "PB"
	case SizeTeraBytes:
		return "TB"
	case SizeGigaBytes:
		return "GB"
	case SizeMegaBytes:
		return "MB"
	case SizeKiloBytes:
		return "KB"
	case SizeBytes:
		return "B"
	}

	return ""
}

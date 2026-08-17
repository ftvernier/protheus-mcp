package protheus

import "time"

func processNowMS() (int64, error) { return time.Now().UnixMilli(), nil }

func round(value float64, places int) float64 {
	factor := 1.0
	for i := 0; i < places; i++ { factor *= 10 }
	return float64(int(value*factor+0.5)) / factor
}

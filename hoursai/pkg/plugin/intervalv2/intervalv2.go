package intervalv2

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/gtime"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

var (
	DefaultRes         int64 = 1500
	defaultMinInterval       = time.Millisecond * 1
	year                     = time.Hour * 24 * 365
	day                      = time.Hour * 24
)

type Interval struct {
	Text  string
	Value time.Duration
}

type IntervalCalculator struct {
	minInterval time.Duration
}

type Calculator interface {
	Calculate(timerange backend.TimeRange, minInterval time.Duration, maxDataPoints int64) Interval
	CalculateSafeInterval(timerange backend.TimeRange, resolution int64) Interval
}

type CalculatorOptions struct {
	MinInterval time.Duration
}

func NewCalculator(opts ...CalculatorOptions) *IntervalCalculator {
	calc := &IntervalCalculator{}

	for _, o := range opts {
		if o.MinInterval == 0 {
			calc.minInterval = defaultMinInterval
		} else {
			calc.minInterval = o.MinInterval
		}
	}

	return calc
}

// Calculate 根据from、to、maxDataPoints、queryInterval计算合理的interval
func (ic *IntervalCalculator) Calculate(timerange backend.TimeRange, queryInterval time.Duration,
	maxDataPoints int64) Interval {
	// 将from、to转换为Unix时间，单位为纳秒
	to := timerange.To.UnixNano()
	from := timerange.From.UnixNano()

	// 如果最大点数传入有误的话，则赋值为默认值1500
	resolution := maxDataPoints
	if resolution == 0 {
		resolution = DefaultRes
	}

	// 根据数据起止时间和最大点数算出最小interval
	minInterval := time.Duration((to - from) / resolution)

	// 如果minInterval小于queryInterval，则返回最小interval
	// 如果minInterval大于queryInterval，则只能使用最小interval（使用queryInterval得到的点数会超过最大点数）
	if minInterval < queryInterval {
		return Interval{Text: FormatDuration(queryInterval), Value: queryInterval}
	}

	// 给minInterval向上取整
	rounded := roundInterval(minInterval)
	return Interval{Text: FormatDuration(rounded), Value: rounded}
}

func (ic *IntervalCalculator) CalculateSafeInterval(timerange backend.TimeRange, safeRes int64) Interval {
	to := timerange.To.UnixNano()
	from := timerange.From.UnixNano()
	safeInterval := time.Duration((to - from) / safeRes)

	rounded := roundInterval(safeInterval)
	return Interval{Text: FormatDuration(rounded), Value: rounded}
}

// GetIntervalFrom 从request中interval、intervalMs、timeInterval中取出正确的interval
func GetIntervalFrom(timeInterval, queryInterval string, queryIntervalMS int64,
	defaultInterval time.Duration) (time.Duration, error) {
	interval := queryInterval
	if interval == "0s" {
		interval = ""
	}

	if interval == "" {
		if queryIntervalMS != 0 {
			return time.Duration(queryIntervalMS) * time.Millisecond, nil
		}
	}
	if interval == "" && timeInterval != "" {
		interval = timeInterval
	}
	if interval == "" {
		return defaultInterval, nil
	}
	parsedInterval, err := ParseIntervalStringToTimeDuration(interval)
	if err != nil {
		return time.Duration(0), err
	}
	return parsedInterval, nil
}

// ParseIntervalStringToTimeDuration 把timeInterval表达式转换成interval
func ParseIntervalStringToTimeDuration(interval string) (time.Duration, error) {
	formattedInterval := strings.Replace(strings.Replace(interval, "<", "", 1), ">", "", 1)
	isPureNum, err := regexp.MatchString(`^\d+$`, formattedInterval)
	if err != nil {
		return time.Duration(0), err
	}
	if isPureNum {
		formattedInterval += "s"
	}
	parsedInterval, err := gtime.ParseDuration(formattedInterval)
	if err != nil {
		return time.Duration(0), err
	}
	return parsedInterval, nil
}

// roundInterval 给interval向上取整
func roundInterval(interval time.Duration) time.Duration {
	switch {
	// 0.01s
	case interval <= 10*time.Millisecond:
		return time.Millisecond * 1 // 0.001s
	// 0.015s
	case interval <= 15*time.Millisecond:
		return time.Millisecond * 10 // 0.01s
	// 0.035s
	case interval <= 35*time.Millisecond:
		return time.Millisecond * 20 // 0.02s
	// 0.075s
	case interval <= 75*time.Millisecond:
		return time.Millisecond * 50 // 0.05s
	// 0.15s
	case interval <= 150*time.Millisecond:
		return time.Millisecond * 100 // 0.1s
	// 0.35s
	case interval <= 350*time.Millisecond:
		return time.Millisecond * 200 // 0.2s
	// 0.75s
	case interval <= 750*time.Millisecond:
		return time.Millisecond * 500 // 0.5s
	// 1.5s
	case interval <= 1500*time.Millisecond:
		return time.Millisecond * 1000 // 1s
	// 3.5s
	case interval <= 3500*time.Millisecond:
		return time.Millisecond * 2000 // 2s
	// 7.5s
	case interval <= 7500*time.Millisecond:
		return time.Millisecond * 5000 // 5s
	// 12.5s
	case interval <= 12500*time.Millisecond:
		return time.Millisecond * 10000 // 10s
	// 17.5s
	case interval <= 17500*time.Millisecond:
		return time.Millisecond * 15000 // 15s
	// 25s
	case interval <= 25000*time.Millisecond:
		return time.Millisecond * 20000 // 20s
	// 45s
	case interval <= 45000*time.Millisecond:
		return time.Millisecond * 30000 // 30s
	// 1.5m
	case interval <= 90000*time.Millisecond:
		return time.Millisecond * 60000 // 1m
	// 3.5m
	case interval <= 210000*time.Millisecond:
		return time.Millisecond * 120000 // 2m
	// 7.5m
	case interval <= 450000*time.Millisecond:
		return time.Millisecond * 300000 // 5m
	// 12.5m
	case interval <= 750000*time.Millisecond:
		return time.Millisecond * 600000 // 10m
	// 17.5m
	case interval <= 1050000*time.Millisecond:
		return time.Millisecond * 900000 // 15m
	// 25m
	case interval <= 1500000*time.Millisecond:
		return time.Millisecond * 1200000 // 20m
	// 45m
	case interval <= 2700000*time.Millisecond:
		return time.Millisecond * 1800000 // 30m
	// 1.5h
	case interval <= 5400000*time.Millisecond:
		return time.Millisecond * 3600000 // 1h
	// 2.5h
	case interval <= 9000000*time.Millisecond:
		return time.Millisecond * 7200000 // 2h
	// 4.5h
	case interval <= 16200000*time.Millisecond:
		return time.Millisecond * 10800000 // 3h
	// 9h
	case interval <= 32400000*time.Millisecond:
		return time.Millisecond * 21600000 // 6h
	// 24h
	case interval <= 86400000*time.Millisecond:
		return time.Millisecond * 43200000 // 12h
	// 48h
	case interval <= 172800000*time.Millisecond:
		return time.Millisecond * 86400000 // 24h
	// 1w
	case interval <= 604800000*time.Millisecond:
		return time.Millisecond * 86400000 // 24h
	// 3w
	case interval <= 1814400000*time.Millisecond:
		return time.Millisecond * 604800000 // 1w
	// 2y
	case interval < 3628800000*time.Millisecond:
		return time.Millisecond * 2592000000 // 30d
	default:
		return time.Millisecond * 31536000000 // 1y
	}
}

// FormatDuration 将毫秒转换为对应时间
func FormatDuration(inter time.Duration) string {
	if inter >= year {
		return fmt.Sprintf("%dy", inter/year)
	}

	if inter >= day {
		return fmt.Sprintf("%dd", inter/day)
	}

	if inter >= time.Hour {
		return fmt.Sprintf("%dh", inter/time.Hour)
	}

	if inter >= time.Minute {
		return fmt.Sprintf("%dm", inter/time.Minute)
	}

	if inter >= time.Second {
		return fmt.Sprintf("%ds", inter/time.Second)
	}

	if inter >= time.Millisecond {
		return fmt.Sprintf("%dms", inter/time.Millisecond)
	}

	return "1ms"
}

package helper

import (
	//"fmt"
	"time"

	str2duration "github.com/xhit/go-str2duration/v2"
)

func GetDurationTime(duration string, cur time.Time) time.Time {
	durationFromString, err := str2duration.ParseDuration(duration)
	if err != nil {
		panic(err)
	}
	return cur.Add(durationFromString)
}

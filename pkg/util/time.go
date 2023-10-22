package util

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// P[n]Y[n]M[n]DT[n]H[n]M[n]S

type ISODuration struct {
	Hour   int
	Minute int
	Second int
}

func (d *ISODuration) Duration() time.Duration {
	return (time.Duration(d.Second) * time.Second) +
		(time.Duration(d.Minute) * time.Minute) +
		(time.Duration(d.Hour) * time.Hour)
}

func (d *ISODuration) String() string {
	if d == nil {
		return ""
	}

	result := fmt.Sprintf("%02d:%02d", d.Minute, d.Second)
	if d.Hour > 0 {
		result = fmt.Sprintf("%02d:%s", d.Hour, result)
	}

	return result
}

var (
	// ignoring year/month/week/day
	isoDurationPattern = regexp.MustCompile(`P.*?T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?`)
	isoDurationNames   = isoDurationPattern.SubexpNames()
)

func ParseISODuration(input string) (*ISODuration, error) {
	matches := isoDurationPattern.FindStringSubmatch(input)

	d := ISODuration{}
	for i, name := range isoDurationNames {
		if i == 0 || name == "" || matches[i] == "" {
			continue
		}

		val, err := strconv.Atoi(matches[i])
		if err != nil {
			return nil, fmt.Errorf("failed to parse duration segment %s: %s", matches[i], err)
		}

		switch name {
		case "hour":
			d.Hour = val
		case "minute":
			d.Minute = val
		case "second":
			d.Second = val
		}
	}

	return &d, nil
}

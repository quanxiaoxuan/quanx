package timex

import (
	"time"
)

type Time time.Time
type Date time.Time

func (t *Time) UnmarshalJSON(data []byte) error {
	location, err := time.ParseInLocation(`"2006-01-02 15:04:05"`, string(data), time.Local)
	if err != nil {
		return err
	}
	*t = Time(location)
	return nil
}

func (t *Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(TimeFmt)+2)
	b = append(b, '"')
	b = time.Time(*t).AppendFormat(b, TimeFmt)
	b = append(b, '"')
	return b, nil
}

func (t *Time) String() string {
	return time.Time(*t).Format(TimeFmt)
}

func (t *Date) UnmarshalJSON(data []byte) error {
	location, err := time.ParseInLocation(`"2006-01-02"`, string(data), time.Local)
	if err != nil {
		return err
	}
	*t = Date(location)
	return nil
}

func (t *Date) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(DateFmt)+2)
	b = append(b, '"')
	b = time.Time(*t).AppendFormat(b, DateFmt)
	b = append(b, '"')
	return b, nil
}

func (t *Date) String() string {
	return time.Time(*t).Format(DateFmt)
}

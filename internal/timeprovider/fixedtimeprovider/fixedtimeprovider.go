package fixedtimeprovider

import "time"

type FixedTimeProvider struct {
	Time time.Time
}

func (f *FixedTimeProvider) Now() time.Time {
	return f.Time
}

func (f *FixedTimeProvider) SetTime(t time.Time) {
	f.Time = t
}

func New(t time.Time) *FixedTimeProvider {
	return &FixedTimeProvider{Time: t}
}

package realtimeprovider

import "time"

type RealTimeProvider struct{}

func (r *RealTimeProvider) Now() time.Time {
	return time.Now()
}

func New() *RealTimeProvider {
	return &RealTimeProvider{}
}

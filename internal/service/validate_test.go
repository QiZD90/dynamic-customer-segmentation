package service

import (
	"testing"
	"time"

	"github.com/QiZD90/dynamic-customer-segmentation/internal/entity"
)

func TestValidateSegmentLists(t *testing.T) {
	mockService := &Service{}

	testCases := []struct {
		testName       string
		addSegments    []entity.SegmentExpiration
		removeSegments []entity.SegmentExpiration
		want           bool
	}{
		{
			testName: "valid input",
			addSegments: []entity.SegmentExpiration{
				{Slug: "SEGMENT_UNIQUE_1", ExpiresAt: time.Time{}},
				{Slug: "SEGMENT_UNIQUE_2", ExpiresAt: time.Time{}},
			},

			removeSegments: []entity.SegmentExpiration{
				{Slug: "SEGMENT_UNIQUE_3", ExpiresAt: time.Time{}},
				{Slug: "SEGMENT_UNIQUE_4", ExpiresAt: time.Time{}},
			},

			want: true,
		},
		{
			testName: "duplicate in add",
			addSegments: []entity.SegmentExpiration{
				{Slug: "SEGMENT_DUPLICATE", ExpiresAt: time.Time{}},
				{Slug: "SEGMENT_DUPLICATE", ExpiresAt: time.Time{}},
			},

			removeSegments: []entity.SegmentExpiration{
				{Slug: "SEGMENT_UNIQUE_3", ExpiresAt: time.Time{}},
				{Slug: "SEGMENT_UNIQUE_4", ExpiresAt: time.Time{}},
			},

			want: false,
		},
		{
			testName: "duplicate in remove",
			addSegments: []entity.SegmentExpiration{
				{Slug: "SEGMENT_UNIQUE_1", ExpiresAt: time.Time{}},
				{Slug: "SEGMENT_UNIQUE_2", ExpiresAt: time.Time{}},
			},

			removeSegments: []entity.SegmentExpiration{
				{Slug: "SEGMENT_DUPLICATE", ExpiresAt: time.Time{}},
				{Slug: "SEGMENT_DUPLICATE", ExpiresAt: time.Time{}},
			},

			want: false,
		},
		{
			testName: "segment appears in both lists",
			addSegments: []entity.SegmentExpiration{
				{Slug: "SEGMENT_UNIQUE_1", ExpiresAt: time.Time{}},
				{Slug: "SEGMENT_DUPLICATE", ExpiresAt: time.Time{}},
			},

			removeSegments: []entity.SegmentExpiration{
				{Slug: "SEGMENT_DUPLICATE", ExpiresAt: time.Time{}},
				{Slug: "SEGMENT_UNIQUE_2", ExpiresAt: time.Time{}},
			},

			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			if got := mockService.ValidateSegmentLists(tc.addSegments, tc.removeSegments); got != tc.want {
				t.Errorf("ValidateSegmentsLists -- %s -- want: %t, got: %t", tc.testName, tc.want, got)
			}
		})
	}
}

package service

import "github.com/QiZD90/dynamic-customer-segmentation/internal/entity"

// ValidateSegmentLists checks for segments appearing more than once in either list
// and for segments appearing in both lists.
// Returns `false` in either case and `true` otherwise
func (s *Service) ValidateSegmentLists(addSegments []entity.SegmentExpiration, removeSegments []entity.SegmentExpiration) bool {
	addSet := make(map[string]struct{})
	removeSet := make(map[string]struct{})

	for _, s := range addSegments {
		_, ok := addSet[s.Slug]
		if ok { // Uh oh, segment slug appeared twice
			return false
		}

		addSet[s.Slug] = struct{}{}
	}

	for _, s := range removeSegments {
		_, ok := removeSet[s.Slug]
		if ok { // Uh oh, segment slug appeared twice
			return false
		}

		removeSet[s.Slug] = struct{}{}
	}

	for k := range addSet {
		_, ok := removeSet[k]
		if ok { // Uh oh, segment appeared in both lists
			return false
		}
	}

	return true
}

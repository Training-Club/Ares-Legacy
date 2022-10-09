package util

import "ares/model"

// Contains returns true if the provided value is within
// the provided collection
//
// This function utilizes generics, but will not work with some
// types.
func Contains[K comparable](value K, collection []K) bool {
	for _, v := range collection {
		if value == v {
			return true
		}
	}

	return false
}

// ContainsStr returns true if the provided value is within
// the provided collection of strings
func ContainsStr(value string, collection []string) bool {
	for _, v := range collection {
		if value == v {
			return true
		}
	}

	return false
}

// ContainsPerm returns true if the provided permission is within
// the provided collection of permissions
func ContainsPerm(value model.Permission, collection []model.Permission) bool {
	for _, v := range collection {
		if value == v {
			return true
		}
	}

	return false
}

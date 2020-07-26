package utils

import "go.mongodb.org/mongo-driver/bson/primitive"

// IndexOf indexof helper
func IndexOf(A []primitive.ObjectID, E primitive.ObjectID) int {
	for i, e := range A {
		if e == E {
			return i
		}
	}
	return -1
}

package utils

import (
	"bytes"
	"main/pkg/constants"
	"net/http"
	"strconv"

	"github.com/btcsuite/btcutil/bech32"
)

func BoolToFloat64(b bool) float64 {
	if b {
		return 1
	}

	return 0
}

func StrToFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}

	return f
}

func StrToInt64(s string) int64 {
	f, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}

	return f
}

func ChangeBech32Prefix(source, newPrefix string) (string, error) {
	_, bytes, err := bech32.Decode(source)
	if err != nil {
		return "", err
	}

	return bech32.Encode(newPrefix, bytes)
}

func Filter[T any](slice []T, f func(T) bool) []T {
	var n []T
	for _, e := range slice {
		if f(e) {
			n = append(n, e)
		}
	}
	return n
}

func Map[T any, V any](slice []T, f func(T) V) []V {
	n := make([]V, len(slice))

	for index, e := range slice {
		n[index] = f(e)
	}
	return n
}

func Find[T any](slice []T, predicate func(T) bool) (*T, bool) {
	for _, elt := range slice {
		if predicate(elt) {
			return &elt, true
		}
	}
	return nil, false
}

func FindIndex[T any](slice []T, predicate func(T) bool) (int, bool) {
	for index, elt := range slice {
		if predicate(elt) {
			return index, true
		}
	}
	return 0, false
}

func CompareTwoBech32(first, second string) (bool, error) {
	_, firstBytes, err := bech32.Decode(first)
	if err != nil {
		return false, err
	}

	_, secondBytes, err := bech32.Decode(second)
	if err != nil {
		return false, err
	}

	return bytes.Equal(firstBytes, secondBytes), nil
}

func GetBlockHeightFromHeader(header http.Header) (int64, error) {
	valueStr := header.Get(constants.HeaderBlockHeight)
	if valueStr == "" {
		return 0, nil
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}

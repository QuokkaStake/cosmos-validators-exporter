package main

import (
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

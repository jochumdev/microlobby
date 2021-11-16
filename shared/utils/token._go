package utils

import (
	"errors"
	"strings"
)

type TokenFormat int32

const TokenFormatUnknown TokenFormat = 0
const TokenFormatBearer TokenFormat = 1
const TokenFormatToken TokenFormat = 2

func ExtractToken(hdr string) (string, TokenFormat, error) {
	if hdr == "" {
		return "", TokenFormatUnknown, errors.New("no authorization header")
	}

	th := strings.Split(hdr, " ")
	if len(th) != 2 {
		return "", TokenFormatUnknown, errors.New("incomplete authorization header")
	}

	if strings.ToLower(th[0]) == "bearer" {
		return th[1], TokenFormatBearer, nil
	} else if strings.ToLower(th[0]) == "token" {
		return th[1], TokenFormatToken, nil
	}

	return "", TokenFormatUnknown, errors.New("unknow token format")
}

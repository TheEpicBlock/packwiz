package core

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"strings"
	"errors"
	"hash"
)

// GetHashImpl gets an implementation of hash.Hash for the given hash type string
func GetHashImpl(hashType string) (hash.Hash, error) {
	switch strings.ToLower(hashType) {
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	case "md5":
		return md5.New(), nil
	}
	// TODO: implement murmur2
	return nil, errors.New("hash implementation not found")
}

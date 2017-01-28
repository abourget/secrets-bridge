package secrets

import (
	"encoding/base64"
	"strings"
)

type Store struct {
	Secrets map[string][]byte
}

func (s *Store) Add(key string, value []byte) (err error) {
	if strings.HasPrefix(key, "b64:") {
		key = key[4:]
		value, err = base64.StdEncoding.DecodeString(string(value))
	} else if strings.HasPrefix(key, "b64u:") {
		key = key[5:]
		value, err = base64.URLEncoding.DecodeString(string(value))
	} else if strings.HasPrefix(key, "rb64:") {
		key = key[5:]
		value, err = base64.RawStdEncoding.DecodeString(string(value))
	} else if strings.HasPrefix(key, "rb64u:") {
		key = key[6:]
		value, err = base64.RawURLEncoding.DecodeString(string(value))
	}
	if err != nil {
		return
	}

	if s.Secrets == nil {
		s.Secrets = make(map[string][]byte)
	}
	s.Secrets[key] = value

	return nil
}

func (s *Store) Get(key string) []byte {
	var encoder func([]byte) string
	switch {
	case strings.HasPrefix(key, "b64:"):
		key = key[4:]
		encoder = base64.StdEncoding.EncodeToString
	case strings.HasPrefix(key, "b64u:"):
		key = key[5:]
		encoder = base64.URLEncoding.EncodeToString
	case strings.HasPrefix(key, "rb64:"):
		key = key[5:]
		encoder = base64.RawStdEncoding.EncodeToString
	case strings.HasPrefix(key, "rb64u:"):
		key = key[6:]
		encoder = base64.RawURLEncoding.EncodeToString
	}

	secret := s.Secrets[key]
	if secret == nil {
		return nil
	}

	if encoder != nil {
		secret = []byte(encoder(secret))
	}

	return secret
}

package user

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

// issueToken returns a self-contained bearer token for userID: no state is
// persisted anywhere (see Service doc). The token encodes userID and an
// absolute expiry, HMAC-signed with s.signingKey so it cannot be forged or
// tampered with without that key.
func (s *Service) issueToken(userID string) string {
	payload := userID + "." + strconv.FormatInt(time.Now().Add(s.tokenTTL).Unix(), 10)
	return encodeSegment([]byte(payload)) + "." + encodeSegment(s.sign([]byte(payload)))
}

// verifyToken checks token's signature and expiry and returns the userID it
// was issued for. ok is false for any malformed, tampered, or expired token.
func (s *Service) verifyToken(token string) (userID string, ok bool) {
	payloadSeg, macSeg, found := strings.Cut(token, ".")
	if !found {
		return "", false
	}
	payload, err := decodeSegment(payloadSeg)
	if err != nil {
		return "", false
	}
	mac, err := decodeSegment(macSeg)
	if err != nil {
		return "", false
	}
	if !hmac.Equal(mac, s.sign(payload)) {
		return "", false
	}

	id, expStr, found := strings.Cut(string(payload), ".")
	if !found {
		return "", false
	}
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return "", false
	}
	if time.Now().Unix() > exp {
		return "", false
	}
	return id, true
}

func (s *Service) sign(payload []byte) []byte {
	mac := hmac.New(sha256.New, s.signingKey)
	mac.Write(payload)
	return mac.Sum(nil)
}

func encodeSegment(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func decodeSegment(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }

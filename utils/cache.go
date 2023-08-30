package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"path/filepath"
)

// BaseURLToFilename turns the base of a URL (scheme, hostname, port) into a valid filename.
// This method replaces the port part with a special string token (because ":" cannot be used
// in filenames on some platforms).
func BaseURLToFilename(u *url.URL) string {
	out := filepath.Join(u.Scheme)

	switch u.Scheme {
	case "http", "https":
		host := u.Hostname()
		port := u.Port()
		if port != "" {
			host += "_PORT" + port
		}
		out = filepath.Join(out, host)
	default:
		panic(fmt.Sprintf("Don't know how to create a cache name for scheme: %s", u.Scheme))
	}

	return out
}

// UrlToFilename turns the provided URL into a hashed filename.
// URLs can contain characters that cannot be used in filenames (like "?", "#", ":"), so in
// order to cache them properly, they are deterministically hashed into ASCII strings.
func UrlToFilename(u *url.URL) string {
	cacheFilename := BaseURLToFilename(u)

	restStr := u.Path
	if u.RawQuery != "" {
		restStr += "?" + u.RawQuery
	}

	hashedFilename := sha256.Sum256([]byte(restStr))
	cacheFilename = filepath.Join(cacheFilename, hex.EncodeToString(hashedFilename[:]))
	return cacheFilename
}

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func ReadJSON[T any](r *http.Request, target T) error {
	dec := json.NewDecoder(r.Body)
	return dec.Decode(target)
}

func WriteJSON[T any](w http.ResponseWriter, r *http.Request, data T) error {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)

	err := enc.Encode(data)

	if err != nil {
		return err
	}

	return nil
}

func ErrorJSON[T any](w http.ResponseWriter, error T, code int) {
	w.WriteHeader(code)
	_ = WriteJSON(w, nil, error)
}

// CacheControlOpts represents Cache-Control response directives
type CacheControlOpts struct {
	Public          bool
	Private         bool
	NoCache         bool
	NoStore         bool
	NoTransform     bool
	MustRevalidate  bool
	ProxyRevalidate bool
	MaxAge          *int
	SMaxAge         *int
	Immutable       bool
}

// Validate checks if the cache control options are valid
func (cc *CacheControlOpts) Validate() error {
	if cc.Public && cc.Private {
		return fmt.Errorf("public and private directives cannot be used together")
	}

	if cc.NoStore && cc.MaxAge != nil {
		return fmt.Errorf("no-store cannot be used with max-age")
	}

	if cc.NoCache && cc.Immutable {
		return fmt.Errorf("no-cache cannot be used with immutable")
	}

	if cc.MaxAge != nil && *cc.MaxAge < 0 {
		return fmt.Errorf("max-age must be non-negative")
	}

	if cc.SMaxAge != nil && *cc.SMaxAge < 0 {
		return fmt.Errorf("s-maxage must be non-negative")
	}

	return nil
}

// EncodeCacheControl converts the CacheControl struct to a header string
func EncodeCacheControl(cc *CacheControlOpts) (string, error) {
	if cc == nil {
		return "", nil
	}

	if err := cc.Validate(); err != nil {
		return "", err
	}

	directives := []string{}

	// Append boolean directives
	if cc.Public {
		directives = append(directives, "public")
	}
	if cc.Private {
		directives = append(directives, "private")
	}
	if cc.NoCache {
		directives = append(directives, "no-cache")
	}
	if cc.NoStore {
		directives = append(directives, "no-store")
	}
	if cc.NoTransform {
		directives = append(directives, "no-transform")
	}
	if cc.MustRevalidate {
		directives = append(directives, "must-revalidate")
	}
	if cc.ProxyRevalidate {
		directives = append(directives, "proxy-revalidate")
	}
	if cc.Immutable {
		directives = append(directives, "immutable")
	}

	// Append numeric directives
	if cc.MaxAge != nil {
		directives = append(directives, fmt.Sprintf("max-age=%d", *cc.MaxAge))
	}
	if cc.SMaxAge != nil {
		directives = append(directives, fmt.Sprintf("s-maxage=%d", *cc.SMaxAge))
	}

	return strings.Join(directives, ", "), nil
}

func WriteCacheControl(w http.ResponseWriter, cc *CacheControlOpts) error {
	cacheControl, err := EncodeCacheControl(cc)

	if err != nil {
		return err
	}

	w.Header().Set("Cache-Control", cacheControl)
	return nil
}

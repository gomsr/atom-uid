package shorturl

import (
	"github.com/gomsr/atom-uid/generator"
	"github.com/gomsr/atom-uid/generator/generators"
	"github.com/gomsr/atom-uid/utilu"
	"math/rand"
)

type DefaultShortUrl struct {
	generator.UidGenerator
}

func New() (*DefaultShortUrl, error) {
	gen6, err := generators.NewDefaultUidGenerator(32, 11, 20, rand.Int63n(2^11))
	if err != nil {
		return nil, err
	} else {
		return &DefaultShortUrl{gen6}, nil
	}
}

// NewV6 it will be valid util 2029-07-30
func NewV6() (*DefaultShortUrl, error) {
	gen6, err := generators.NewDefaultUidGenerator(64-1-3-3, 3, 3, rand.Int63n(2^3), "2024-07-30")
	if err != nil {
		return nil, err
	} else {
		return &DefaultShortUrl{gen6}, nil
	}
}

// NewV7 it will be valid util 2053-07-30
func NewV7() (*DefaultShortUrl, error) {
	gen6, err := generators.NewDefaultUidGenerator(64-1-6-5, 6, 5, rand.Int63n(2^6), "2023-07-30")
	if err != nil {
		return nil, err
	} else {
		return &DefaultShortUrl{gen6}, nil
	}
}

// NewV8 it will be valid util 2050-07-30
func NewV8() (*DefaultShortUrl, error) {
	gen6, err := generators.NewDefaultUidGenerator(64-1-8-9, 8, 9, rand.Int63n(2^8), "2023-07-30")
	if err != nil {
		return nil, err
	} else {
		return &DefaultShortUrl{gen6}, nil
	}
}

func (c *DefaultShortUrl) ShortUrl() string {
	return utilu.ToBase62R(c.MustUID())
}

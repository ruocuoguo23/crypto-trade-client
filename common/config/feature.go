package config

import "errors"

type FeatureName string

const (
	FeatResetHeight FeatureName = "FeatResetHeight"
)

var (
	ErrBlockHeightRestricted = errors.New("cannot read the height that smaller than the config one")
)

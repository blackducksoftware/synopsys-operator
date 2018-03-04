package model

type AuxiliaryConfig struct {
	Namespace string

	// AUTH CONFIGS
	// These are given to containers through secrets or other mechanisms.
	// Not necessarily a one-to-one text replacement.
	// TODO Lets try to have this injected on serviceaccount
	// at pod startup, eventually Service accounts.
	DockerPassword string
	DockerUsername string
}

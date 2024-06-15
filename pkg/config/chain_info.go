package config

type ChainInfo interface {
	GetQueries() Queries
	GetHost() string
	GetName() string
}

package config

import "errors"

type Validator struct {
	Address          string `toml:"address"`
	ConsensusAddress string `toml:"consensus-address"`
}

func (v *Validator) Validate() error {
	if v.Address == "" {
		return errors.New("validator address is expected!")
	}

	return nil
}

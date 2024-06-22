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

func (v *Validator) DisplayWarnings(chain *Chain) []Warning {
	warnings := []Warning{}
	if v.ConsensusAddress == "" {
		warnings = append(warnings, Warning{
			Message: "Consensus address is not set, cannot display signing info metrics.",
			Labels: map[string]string{
				"chain":     chain.Name,
				"validator": v.Address,
			},
		})
	}

	return warnings
}

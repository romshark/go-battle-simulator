package battle

import "github.com/pkg/errors"

// SoldierAttributes represents a soldier's attributes
type SoldierAttributes struct {
	HealthMin             float64
	HealthMax             float64
	AttackStrengthMin     float64
	AttackStrengthMax     float64
	DodgeChanceMin        float64
	DodgeChanceMax        float64
	HitChanceMin          float64
	HitChanceMax          float64
	MoraleIncrementFactor float64
	MoraleDecrementFactor float64
}

// Verify verifies attribute values
func (attrs *SoldierAttributes) Verify() error {
	verifyPercentage := func(attrName string, min, max float64) error {
		if min < 0 || min > 1 {
			return errors.Errorf("%s: invalid minimum %%: %.1f", attrName, min)
		}
		if max < 0 || max > 1 {
			return errors.Errorf("%s: invalid maximum %%: %.1f", attrName, max)
		}
		if min > max {
			return errors.Errorf(
				"%s: minimum (%.1f) is greater maximum (%.1f)",
				attrName,
				min,
				max,
			)
		}
		return nil
	}

	verifyMinMax := func(attrName string, allowedMin, min, max float64) error {
		if min < 1 {
			return errors.Errorf(
				"%s: invalid min: %.1f (allowed min: %.1f)",
				attrName,
				min,
				allowedMin,
			)
		}
		if min > max {
			return errors.Errorf(
				"%s: min (%.1f) greater max (%.1f)",
				attrName,
				min,
				max,
			)
		}
		return nil
	}

	if err := verifyMinMax(
		"health",
		1,
		attrs.HealthMin,
		attrs.HealthMax,
	); err != nil {
		return err
	}

	if attrs.MoraleIncrementFactor < 0 {
		return errors.Errorf(
			"morale increment factor: invalid %.1f",
			attrs.MoraleIncrementFactor,
		)
	}

	if attrs.MoraleDecrementFactor < 0 {
		return errors.Errorf(
			"morale decrement factor: invalid %.1f",
			attrs.MoraleDecrementFactor,
		)
	}

	if err := verifyMinMax(
		"attack strength",
		1,
		attrs.AttackStrengthMin,
		attrs.AttackStrengthMax,
	); err != nil {
		return err
	}

	if err := verifyPercentage(
		"dodge chance",
		attrs.DodgeChanceMin,
		attrs.DodgeChanceMax,
	); err != nil {
		return err
	}

	return verifyPercentage(
		"hit chance",
		attrs.HitChanceMin,
		attrs.HitChanceMax,
	)
}

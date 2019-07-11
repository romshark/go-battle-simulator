package battle

// SoldierStatistics represents the statistics of a soldier
type SoldierStatistics struct {
	// Misses represents the amount of missed attacks
	Misses uint

	// Hits represents the amount of hits performed
	Hits uint

	// DamageTaken represents the total sum of damage taken
	DamageTaken float64

	// DamageCaused represents the total sum of damage caused
	DamageCaused float64

	// Kills represents the amount of kills performed
	Kills uint

	// Dodges represents the amount of dodged attacks
	Dodges uint
}

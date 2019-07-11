package battle

import "fmt"

// SoldierID represents a soldier's unique identifier
type SoldierID struct {
	Faction string
	Name    string
}

// String stringifies the soldier ID
func (id SoldierID) String() string {
	return fmt.Sprintf("%s (%s)", id.Name, id.Faction)
}

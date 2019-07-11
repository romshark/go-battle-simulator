package battle

import "fmt"

// Event represents an abstract battle event
type Event interface{}

// EventDodge represents an event describing a dodged attack
type EventDodge struct {
	Attacker      Soldier
	Defernder     Soldier
	MoralePenalty float64
}

// String turns the event into a message
func (ev EventDodge) String() string {
	return fmt.Sprintf(
		"%s dodged an attack of %s (morale penalty for the attacker: %.1f%%)",
		ev.Defernder.ID(),
		ev.Attacker.ID(),
		ev.MoralePenalty*100,
	)
}

// EventMiss represents an event describing an unsuccessful attack
type EventMiss struct {
	Attacker      Soldier
	Attacked      Soldier
	MoralePenalty float64
}

// String turns the event into a message
func (ev EventMiss) String() string {
	return fmt.Sprintf(
		"%s missed when trying to attack %s (morale penalty: %.1f%%)",
		ev.Attacker.ID(),
		ev.Attacked.ID(),
		ev.MoralePenalty*100,
	)
}

// EventHit represents an event describing a successful attack
type EventHit struct {
	Attacker    Soldier
	Attacked    Soldier
	DamageDealt float64
	MoraleBonus float64
}

// String turns the event into a message
func (ev EventHit) String() string {
	return fmt.Sprintf(
		"%s hit and dealt %.1f damage to %s (morale bonus: %.1f%%)",
		ev.Attacker.ID(),
		ev.DamageDealt,
		ev.Attacked.ID(),
		ev.MoraleBonus*100,
	)
}

// EventKill represents an event describing a kill
type EventKill struct {
	Attacker    Soldier
	Killed      Soldier
	DamageDealt float64
	MoraleBonus float64
}

// String turns the event into a message
func (ev EventKill) String() string {
	return fmt.Sprintf(
		"%s hit, dealt %.1f damage and killed %s (morale bonus: %.1f%%)",
		ev.Attacker.ID(),
		ev.DamageDealt,
		ev.Killed.ID(),
		ev.MoraleBonus*100,
	)
}

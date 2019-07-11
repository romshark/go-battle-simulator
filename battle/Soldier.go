package battle

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Soldier represents an abstract soldier
type Soldier interface {
	// ID returns the id of the soldier
	ID() SoldierID

	// Status returns the soldier's status
	Status() SoldierStatus

	// Stats returns the soldier's statistics
	Stats() SoldierStatistics

	// IsAlive returns true if the soldier is still alive
	IsAlive() bool

	// JoinBattle makes a soldier join the battle
	JoinBattle(ctx context.Context)

	// TakeDamage makes a soldier take damage and returns an error
	// if the attack was successfully dodged
	TakeDamage(
		from Soldier,
		damage float64,
	) (
		damageDealt float64,
		killed bool,
		err error,
	)

	// Attack makes a soldier attack an opponent and returns an error
	// if the attack was successfully dodged by the opponent or the soldier
	// missed
	Attack(opponent Soldier) (
		damageDealt float64,
		killed bool,
		err error,
	)

	// AddMorale increases or decreases the morale depending on whether
	// a positive or a negative percentage was passed
	AddMorale(percent float64) (
		newMorale float64,
		newActionDelay time.Duration,
	)
}

/*************************************************************\
	Implementation
\*************************************************************/

// soldier represents a soldier implementation
type soldier struct {
	lock         *sync.Mutex
	actionTicker *DynamicTicker
	endOfLife    chan struct{}
	attrs        SoldierAttributes
	id           SoldierID
	maxHealth    float64
	status       SoldierStatus
	stats        SoldierStatistics
	battleConfig Config
	battlefield  Battlefield
	battleLog    LogWriter
}

// newSoldier creates a new randomly parameterized soldier instance
func newSoldier(
	name string,
	factionName string,
	attrs SoldierAttributes,
	battleConfig Config,
	battlefield Battlefield,
	battleLog LogWriter,
) (*soldier, error) {
	if battlefield == nil {
		return nil, errors.New("missing battlefield")
	}

	if battleLog == nil {
		return nil, errors.New("missing battle log writer")
	}

	// Verify the input attributes
	if err := attrs.Verify(); err != nil {
		return nil, errors.Wrap(err, "invalid attributes")
	}

	// Determine random max health
	maxHealth := random(attrs.HealthMin, attrs.HealthMax)

	// Verify faction name
	if len(factionName) < 1 {
		return nil, errors.Errorf("invalid faction name: '%s'", factionName)
	}

	return &soldier{
		lock:         &sync.Mutex{},
		actionTicker: NewDynamicTicker(),
		endOfLife:    make(chan struct{}),
		id: SoldierID{
			Faction: factionName,
			Name:    name,
		},
		status: SoldierStatus{
			Health: maxHealth,
			Morale: 1.0,
		},
		maxHealth:    maxHealth,
		attrs:        attrs,
		battleConfig: battleConfig,
		battlefield:  battlefield,
		battleLog:    battleLog,
	}, nil
}

// IsAlive implements the Soldier interface
func (s *soldier) IsAlive() bool {
	s.lock.Lock()
	isAlive := s.status.Health > 0
	s.lock.Unlock()
	return isAlive
}

// ResetActionTicker recalculates the action delay based on the current morale
// and resets the action ticker
//
// This method is thread-safe
func (s *soldier) ResetActionTicker() time.Duration {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.resetActionTicker()
}

// resetActionTicker recalculates the action delay based on the current morale
// and resets the action ticker
func (s *soldier) resetActionTicker() time.Duration {
	// Affect action ticker
	actionDelay := s.calculateActionDelay(s.battleConfig.BaseActionDelay)
	s.actionTicker.Reset(actionDelay)
	return actionDelay
}

// AddMorale increases the morale of the soldier incase a positive value is
// given or decreases it if the value is negative.
//
// This method is thread-safe.
func (s *soldier) AddMorale(percent float64) (
	newMorale float64,
	newActionDelay time.Duration,
) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.addMorale(percent)
}

// addMorale accept both negative and positive percentages
func (s *soldier) addMorale(percent float64) (
	newMorale float64,
	newActionDelay time.Duration,
) {
	if percent < -1 || percent > 1 {
		panic(errors.Errorf("invalid percentage value: %.1f", percent))
	}

	// Select morale factor based on
	// whether the morale is increased or decreased
	factor := s.attrs.MoraleIncrementFactor
	if percent < 0 {
		factor = s.attrs.MoraleDecrementFactor
	}

	s.status.Morale += percent * factor
	if s.status.Morale > 1 {
		// Full morale
		s.status.Morale = 1
	} else if s.status.Morale < 0 {
		// No morale
		s.status.Morale = 0
	}
	newMorale = s.status.Morale
	newActionDelay = s.resetActionTicker()

	return
}

func (s *soldier) takeAction() {
	// Find an opponent
	opponent, err := s.battlefield.FindOpponent(s.id.Faction)
	switch err {
	case ErrNoMoreOpponents:
		// The battle is won! No more opponents are left on the battlefield
		s.endLife(false)
		return
	case nil:
		// A new opponent is found
	default:
		// Unexpected error
		panic(errors.Wrap(err, "unexpected opponent seach err"))
	}

	if s.ID() == opponent.ID() {
		panic(errors.Errorf("Soldier %s attacks himself", s.ID()))
	}

	// Try to deal some damage to the opponent and log any event
	damageDealt, killed, err := s.Attack(opponent)
	switch err {
	case ErrDodged:
		// Dammit, the opponent dodged!
		// Decrease morale by 5%
		moralePenalty := -0.05
		s.AddMorale(moralePenalty)
		s.battleLog.PushEvent(EventDodge{
			Attacker:      s,
			Defernder:     opponent,
			MoralePenalty: moralePenalty,
		})
	case ErrMissed:
		// Dammit, I missed!
		// Decrease morale by 10%
		moralePenalty := -.1
		s.AddMorale(moralePenalty)
		s.battleLog.PushEvent(EventMiss{
			Attacker:      s,
			Attacked:      opponent,
			MoralePenalty: moralePenalty,
		})
	case nil:
		if killed {
			// F@ck yeah! I killed one!
			// Increase morale by 50%
			moraleBonus := 0.5
			s.AddMorale(moraleBonus)
			s.battleLog.PushEvent(EventKill{
				Attacker:    s,
				Killed:      opponent,
				DamageDealt: damageDealt,
				MoraleBonus: moraleBonus,
			})
		} else {
			// Fine! I dealt some damage!
			// Increase morale by 5%
			moraleBonus := 0.05
			s.AddMorale(moraleBonus)
			s.battleLog.PushEvent(EventHit{
				Attacker:    s,
				Attacked:    opponent,
				DamageDealt: damageDealt,
				MoraleBonus: moraleBonus,
			})
		}
	}
	return
}

// calculateActionDelay calculates the delay for the next action based on
// the current morale percentage
func (s *soldier) calculateActionDelay(
	baseDelay time.Duration,
) time.Duration {
	penalty := time.Duration(float64(baseDelay) * s.status.Morale / 2)
	return baseDelay - penalty
}

// JoinBattle implements the Soldier interface
func (s *soldier) JoinBattle(ctx context.Context) {
	defer func() {
		// Cleanup
		s.actionTicker.Reset(0)
	}()
	s.ResetActionTicker()

LIFE_LOOP:
	for {
		select {
		case <-ctx.Done():
			// The battle was canceled
			break LIFE_LOOP

		case <-s.actionTicker.C():
			// Time to take some action!
			s.takeAction()

		case <-s.endOfLife:
			// Death
			break LIFE_LOOP
		}
	}
}

func (s *soldier) endLife(dueToDeath bool) {
	// End the life-loop in case of a lethal strike
	close(s.endOfLife)

	// Mark the soldier as killed
	if dueToDeath {
		if err := s.battlefield.MarkDead(s); err != nil {
			panic(errors.Wrap(err, "unexpected error during MarkDead"))
		}
	}
}

// TakeDamage implements the Soldier interface
func (s *soldier) TakeDamage(
	from Soldier,
	damage float64,
) (
	damageDealt float64,
	killed bool,
	err error,
) {
	if from == nil {
		return 0, false, errors.New("no opponent to take damage from")
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if luck(random(s.attrs.DodgeChanceMin, s.attrs.DodgeChanceMax)) {
		// Successfully dodged the attack
		// Increase morale by 25%
		s.addMorale(0.25)
		s.stats.Dodges++
		return 0, false, ErrDodged
	}

	s.status.Health -= damage
	s.stats.DamageTaken += damage
	if s.status.Health < 0 {
		// Die
		s.status.Health = 0
		s.endLife(true)
		return damage, true, nil
	}

	// Ouch! Couldn't dodge the attack and took some damage
	// Decrease morale by 15%
	s.addMorale(-.15)

	return damage, false, nil
}

// Attack implements the Soldier interface
func (s *soldier) Attack(opponent Soldier) (
	damageDealt float64,
	killed bool,
	err error,
) {
	if opponent == nil {
		return 0, false, errors.New("no opponent to attack")
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if !luck(random(s.attrs.HitChanceMin, s.attrs.HitChanceMax)) {
		// Miss, no luck
		s.stats.Misses++
		return 0, false, ErrMissed
	}

	potentialDamage := random(
		s.attrs.AttackStrengthMin,
		s.attrs.AttackStrengthMax,
	)
	damageDealt, killed, err = opponent.TakeDamage(s, potentialDamage)
	if err != nil {
		// Opponent dodged the attack
		s.stats.Misses++
		return 0, false, err
	}

	// Hit, damage dealt
	s.stats.Hits++
	if killed {
		s.stats.Kills++
	}
	s.stats.DamageCaused += damageDealt

	return damageDealt, killed, nil
}

// ID implements the Soldier interface
func (s *soldier) ID() SoldierID {
	return s.id
}

// Stats implements the Soldier interface
func (s *soldier) Stats() SoldierStatistics {
	stats := s.stats
	s.lock.Unlock()
	return stats
}

// Status implements the Soldier interface
func (s *soldier) Status() SoldierStatus {
	s.lock.Lock()
	status := s.status
	s.lock.Unlock()
	return status
}

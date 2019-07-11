package battle

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/pkg/errors"
)

// Battlefield allows
type Battlefield interface {
	// FindOpponent returns either an opponent from an opposing faction
	// or an error if case no more opponents are left
	FindOpponent(ownFactionName string) (Soldier, error)

	// MarkDead marks a soldier as dead
	MarkDead(soldier Soldier) error
}

// Faction represents a faction config
type Faction struct {
	Name              string
	ArmySize          uint
	SoldierAttributes SoldierAttributes
}

// Battle represents a battle
type Battle struct {
	lock   *sync.Mutex
	armies map[string][]Soldier
	alive  map[string][]Soldier
	stats  *Statistics
	config Config
}

// Config represents the configuration of a battle
type Config struct {
	BaseActionDelay time.Duration
}

// NewBattle creates a new battle
func NewBattle(
	config Config,
	factions ...Faction,
) (*Battle, error) {
	if len(factions) < 2 {
		return nil, errors.Errorf(
			"invalid number of factions: %d",
			len(factions),
		)
	}

	battle := &Battle{
		lock:   &sync.Mutex{},
		stats:  NewStatistics(),
		config: config,
	}

	armies := make(map[string][]Soldier, len(factions))
	for _, faction := range factions {
		// Generate the faction's army
		names := make(map[SoldierID]struct{}, faction.ArmySize)
		army := make([]Soldier, 0, faction.ArmySize)
		for i := uint(0); i < faction.ArmySize; i++ {
			// Generate unique name
			id := SoldierID{
				Faction: faction.Name,
			}
			for {
				id.Name = randomdata.SillyName()
				if _, alreadyExists := names[id]; !alreadyExists {
					break
				}
			}

			soldier, err := newSoldier(
				id.Name,
				faction.Name,
				faction.SoldierAttributes,
				config,
				Battlefield(battle),
				LogWriter(battle.stats),
			)
			if err != nil {
				return nil, errors.Wrapf(
					err,
					"generating soldier for faction %s",
					faction.Name,
				)
			}
			army = append(army, soldier)
		}
		armies[faction.Name] = army
	}
	battle.armies = armies

	alive := make(map[string][]Soldier, len(factions))
	for factionName, army := range armies {
		cp := make([]Soldier, len(army))
		copy(cp, army)
		alive[factionName] = cp
	}
	battle.alive = alive

	return battle, nil
}

// Statistics returns the battle statistics reader
func (b *Battle) Statistics() StatisticsReader {
	return b.stats
}

// FindOpponent implements the interface Battlefield
func (b *Battle) FindOpponent(ownFactionName string) (Soldier, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	opposingFactions := make([]string, len(b.alive)-1)
	i := 0
	for faction := range b.alive {
		if faction == ownFactionName {
			continue
		}
		opposingFactions[i] = faction
		i++
	}

	randFactionName := opposingFactions[rand.Intn(len(opposingFactions))]
	army := b.alive[randFactionName]

	if len(army) < 1 {
		return nil, ErrNoMoreOpponents
	}

	// Take random opponent
	return army[rand.Intn(len(army))], nil
}

// MarkDead implements the interface Battlefield
func (b *Battle) MarkDead(soldier Soldier) error {
	id := soldier.ID()

	b.lock.Lock()
	defer b.lock.Unlock()

	// Find army
	alive, armyFound := b.alive[id.Faction]
	if !armyFound {
		return errors.Errorf("unknown faction '%s'", id.Faction)
	}

	// Find soldier
	index := -1
	for i, s := range alive {
		if s.ID() == id {
			index = i
			break
		}
	}
	if index < 0 {
		// Dead or not found
		return nil
	}

	// Remove the soldier from the list of the living
	alive[len(alive)-1], alive[index] = alive[index], alive[len(alive)-1]
	b.alive[id.Faction] = alive[:len(alive)-1]

	return nil
}

// Run runs the battle until it's either finished or canceled by the provided
// context
func (b *Battle) Run(ctx context.Context) {
	wg := &sync.WaitGroup{}

	// Register all soldiers in the wait-group
	h := 0
	for _, army := range b.armies {
		h += len(army)
		wg.Add(len(army))
	}

	// Make the soldiers join the battle
	for _, army := range b.armies {
		for _, soldier := range army {
			s := soldier
			go func() {
				defer wg.Done()
				s.JoinBattle(ctx)
			}()
		}
	}

	// Wait for the battle to finish
	// by waiting for all soldiers to finish
	wg.Wait()

	// Stop recorcing battle statistics
	b.stats.StopRecording()

	// Determine the winner
	for factionName, army := range b.alive {
		if len(army) > 0 {
			b.stats.winnerFaction = factionName
		}
	}
}

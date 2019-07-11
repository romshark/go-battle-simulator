package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/romshark/go-battle-simulator/battle"
)

func init()                           { rand.Seed(time.Now().Unix()) }
func random(min, max float64) float64 { return min + rand.Float64()*(max-min) }

/*************************************************************\
	CONF
\*************************************************************/

var confBaseActionDelay = time.Millisecond * 100

var confFactions = []battle.Faction{
	battle.Faction{
		Name:     "A",
		ArmySize: 1,
		SoldierAttributes: battle.SoldierAttributes{
			HealthMin:             random(25, 50),
			HealthMax:             random(50, 70),
			AttackStrengthMin:     random(5, 10),
			AttackStrengthMax:     random(10, 20),
			DodgeChanceMin:        random(.25, .5),
			DodgeChanceMax:        random(.5, .75),
			HitChanceMin:          random(.25, .5),
			HitChanceMax:          random(.5, .75),
			MoraleIncrementFactor: random(1, 1.5),
			MoraleDecrementFactor: random(1, 1.5),
		},
	},
	battle.Faction{
		Name:     "B",
		ArmySize: 1,
		SoldierAttributes: battle.SoldierAttributes{
			HealthMin:             random(25, 50),
			HealthMax:             random(50, 70),
			AttackStrengthMin:     random(5, 10),
			AttackStrengthMax:     random(10, 20),
			DodgeChanceMin:        random(.25, .5),
			DodgeChanceMax:        random(.5, .75),
			HitChanceMin:          random(.25, .5),
			HitChanceMax:          random(.5, .75),
			MoraleIncrementFactor: random(1, 1.5),
			MoraleDecrementFactor: random(1, 1.5),
		},
	},
}

func main() {
	btl, err := battle.NewBattle(
		battle.Config{
			BaseActionDelay: confBaseActionDelay,
		},
		confFactions...,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()

	statistics := btl.Statistics()

	// Start real-time log stream listener
	go func() {
		for battleLogEntry := range statistics.LogStream() {
			tm := battleLogEntry.Time
			log.Printf(
				"%d:%d:%d - %s",
				tm.Hour(),
				tm.Minute(),
				tm.Second(),
				battleLogEntry.Event,
			)
		}
	}()

	log.Print("The battle begins!")
	btl.Run(ctx)
	log.Printf("Battle ended! Faction '%s' wins!", statistics.WinnerFaction())
}

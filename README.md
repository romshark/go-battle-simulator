[![Build Status](https://travis-ci.org/romshark/go-battle-simulator.svg?branch=master)](https://travis-ci.org/romshark/go-battle-simulator)
[![Go Report Card](https://goreportcard.com/badge/github.com/romshark/go-battle-simulator)](https://goreportcard.com/report/github.com/romshark/go-battle-simulator)

# go-battle-simulator
A concurrent goroutine-based battle simulator written in Go.
It simulates individual soldiers of multiple opposing factions on the battlefield. Each soldier has a number of attributes:

- health
- morale
- attack strength (min/max)
- dodge chance (min/max)
- hit chance (min/max)
 - morale factor (inc/dec; WiP)
 
 The morale percentage of a soldier affects his action delay duration (the initiative time interval) - 100% morale is equivalent to only half the base delay duration (which is configurable in the battle settings). Being hit, missing or having an attack dodged decreases the morale (increasing the action delay) while hitting, killing and dodging attacks increases it (decreasing the action delay).
 
 Each faction can be configured to have different attribute traits:
 
```go
var confFactions = []battle.Faction{
  battle.Faction{
    Name:     "Faction A",
    ArmySize: 10,
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
    Name:     "Faction B",
    ArmySize: 10,
    SoldierAttributes: battle.SoldierAttributes{
      HealthMin:             random(20, 30),
      HealthMax:             random(30, 90),
      AttackStrengthMin:     random(2, 6),
      AttackStrengthMax:     random(6, 8),
      DodgeChanceMin:        random(.6, .8),
      DodgeChanceMax:        random(.8, .9),
      HitChanceMin:          random(.1, .3),
      HitChanceMax:          random(.3, .4),
      MoraleIncrementFactor: random(1, 1.5),
      MoraleDecrementFactor: random(1, 1.5),
    },
  },
}

battle.NewBattle(
  battle.Config{
    BaseActionDelay: 1*time.Second,
  },
  confFactions...,
)
```
 
 A battle is over once only a single faction is left on the battlefield. The battlelog is streamed to an event channel which can, for example, be streamed to the console in a stringified form:
 
```
2019/07/11 17:29:31 The battle begins!
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) missed when trying to attack Kangarooboulder (A) (morale penalty: -10.0%)
2019/07/11 17:29:31 17:29:31 - Kangarooboulder (A) missed when trying to attack Knavemaple (B) (morale penalty: -10.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) missed when trying to attack Kangarooboulder (A) (morale penalty: -10.0%)
2019/07/11 17:29:31 17:29:31 - Kangarooboulder (A) missed when trying to attack Knavemaple (B) (morale penalty: -10.0%)
2019/07/11 17:29:31 17:29:31 - Kangarooboulder (A) dodged an attack of Knavemaple (B) (morale penalty for the attacker: -5.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) missed when trying to attack Kangarooboulder (A) (morale penalty: -10.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) dodged an attack of Kangarooboulder (A) (morale penalty for the attacker: -5.0%)
2019/07/11 17:29:31 17:29:31 - Kangarooboulder (A) dodged an attack of Knavemaple (B) (morale penalty for the attacker: -5.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) missed when trying to attack Kangarooboulder (A) (morale penalty: -10.0%)
2019/07/11 17:29:31 17:29:31 - Kangarooboulder (A) missed when trying to attack Knavemaple (B) (morale penalty: -10.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) hit and dealt 11.2 damage to Kangarooboulder (A) (morale bonus: 5.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) dodged an attack of Kangarooboulder (A) (morale penalty for the attacker: -5.0%)
2019/07/11 17:29:31 17:29:31 - Kangarooboulder (A) dodged an attack of Knavemaple (B) (morale penalty for the attacker: -5.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) hit and dealt 11.2 damage to Kangarooboulder (A) (morale bonus: 5.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) missed when trying to attack Kangarooboulder (A) (morale penalty: -10.0%)
2019/07/11 17:29:31 17:29:31 - Kangarooboulder (A) missed when trying to attack Knavemaple (B) (morale penalty: -10.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) missed when trying to attack Kangarooboulder (A) (morale penalty: -10.0%)
2019/07/11 17:29:31 17:29:31 - Kangarooboulder (A) missed when trying to attack Knavemaple (B) (morale penalty: -10.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) hit and dealt 6.4 damage to Kangarooboulder (A) (morale bonus: 5.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) missed when trying to attack Kangarooboulder (A) (morale penalty: -10.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) dodged an attack of Kangarooboulder (A) (morale penalty for the attacker: -5.0%)
2019/07/11 17:29:31 17:29:31 - Knavemaple (B) hit, dealt 9.2 damage and killed Kangarooboulder (A) (morale bonus: 50.0%)
2019/07/11 17:29:31 Battle ended! Faction 'B' wins!
```

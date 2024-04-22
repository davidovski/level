package main

import (
    "fmt"
)

func ReverseLevel(g *Game) {
    fmt.Printf("pframe %d/%d\n", 0, len(g.playerAi))
    g.state = REVERSING
    g.shaderName = "vcr"
}

func playTheEnd(g *Game) {
    g.animStart = g.time
    g.state = END
}

func afterReversed(g *Game) {
    g.shaderName = "none"
    g.ResetAll()
    g.playerAiIdx = 0
    g.TransitionState()
}

func levelStart(g *Game) {

    for _, o := range g.objects {
        o.movable = false
    }
}

func StartLevel1(g *Game ) {
    g.state = IN_GAME

    // when hit end
    g.QueueState(func (g *Game){
        g.animStart = g.time
        playTheEnd(g)
        g.state = END

    })
    // after end
    g.QueueState(ReverseLevel)
    // after reversed
    g.QueueState(afterReversed)
    g.QueueState(StartLevel2)
}

func StartLevel2(g *Game) {
    g.state = PLACING

    g.toPlace = append(g.toPlace, NewLeftSpike(g, 0, 0))

    // after end
    g.QueueState(ReverseLevel)
    // after reversed
    g.QueueState(afterReversed)
    g.QueueState(StartLevel3)
}

func StartLevel3(g *Game) {
    g.state = PLACING

    g.toPlace = append(g.toPlace, NewSpike(g, 0, 0))

    // after end
    g.QueueState(ReverseLevel)
    // after reversed
    g.QueueState(afterReversed)
    g.QueueState(StartLevel3)
}


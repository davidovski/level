package main

import (
    "fmt"
)

func ReverseLevel(g *Game) {
    fmt.Printf("pframe %d/%d\n", 0, len(g.playerAi))
    g.SetReversing()
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
    g.ResetAll()
    g.playerAiIdx = 0
    g.SetInGame()
    for _, o := range g.objects {
        o.movable = false
    }
}

func StartGame(g *Game) {
    StartLevel1(g)
}

func StartLevel1(g *Game ) {
    g.SetInGame()

    tilemap := NewTilemap([][]int{
            {
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 51, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 67, 0,
                0, 50, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 0,
                0, 53, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 0,
                0, 69, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 0,
                0, 69, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 0,
			},
		}, 25)

    g.tilemap = &tilemap
    g.tilemap.UpdateSurface()

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
    g.SetPlacing()

    g.toPlace = append(g.toPlace, NewLeftSpike(g, 0, 0))

    // after end
    g.QueueState(ReverseLevel)
    // after reversed
    g.QueueState(afterReversed)
    g.QueueState(StartLevel3)
}

func StartLevel3(g *Game) {
    g.SetPlacing()

    g.toPlace = append(g.toPlace, NewSpike(g, 0, 0))

    // after end
    g.QueueState(ReverseLevel)
    // after reversed
    g.QueueState(afterReversed)
    g.QueueState(StartLevel4)
}

func StartLevel4(g *Game) {
    g.SetPlacing()
    g.toPlace = append(g.toPlace, NewSpring(g, 0, 0))

    g.ClearAll()
    tilemap := NewTilemap([][]int{
            {
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  35, 36, 34, 35, 36, 34, 35, 36, 34, 67, 0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 0,
                0, 0,  51, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 0,
                0, 50, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 0,
                0, 53, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 0,
                0, 69, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39,
			},
		}, 25)

    g.tilemap = &tilemap
    g.tilemap.UpdateSurface()
    g.exit.startx = 20 * tileSize
    g.exit.starty = 6 * tileSize

    g.ResetAll()
    g.playerAi = g.playerAi[:0]

    g.QueueState(levelStart)
    // after end
    g.QueueState(ReverseLevel)
    // after reversed
    g.QueueState(afterReversed)
    g.QueueState(StartLevel5)
}

func StartLevel5(g *Game) {
    g.SetPlacing()
    g.toPlace = append(g.toPlace, NewSpike(g, 0, 0))
    g.exit.movable = true

    // after end
    g.QueueState(ReverseLevel)
    // after reversed
    g.QueueState(afterReversed)
    g.QueueState(StartLevel5)
}

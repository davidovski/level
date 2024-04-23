package main

func StartGame(g *Game) {
    //g.state = IN_GAME
    g.player = NewPlayer(g, 4 * tileSize, 9 * tileSize)
    g.objects = append(g.objects, g.player)
    g.exit = NewExit(g, 21 * tileSize, 9 * tileSize)
    g.objects = append(g.objects, g.exit)

    g.ResetAll()
    StartLevel1(g)
}

func PauseScreen(g *Game) {
    g.SetPaused()
}

func ReverseLevel(g *Game) {
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
}

func noMoveable(g *Game) {
    for _, o := range g.objects {
        o.movable = false
    }
}


func StartLevel1(g *Game ) {
    //g.SetInGame()

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
                0, 69, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 0,
			},
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
                0, 0, 99, 84, 82, 83, 84, 82, 83, 84, 82, 83, 84, 82, 83, 84, 82, 83, 84, 82, 83, 84, 82, 115, 0,
                0, 98, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
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
    g.QueueState(func (g *Game){
        // What do you mean "is that it?". I kinda uhh ran out of time to finish the actual game but uhhh if you want I can let you have a go at finishing it for me. Just press R to reverse time and we'll go from there
        PauseScreen(g)
        g.audioPlayer.voiceAudio[0].Play()
    })
    // after end
    g.QueueState(func (g *Game){
        g.animStart = g.time - endCardDuration
        ReverseLevel(g)
    })
    // after reversed
    g.QueueState(func (g *Game){
        g.animStart = 0
        afterReversed(g)
    })
    g.QueueState(StartLevel2)
}

func StartLevel2(g *Game) {
    g.SetPlacing()
    //noMoveable(g)

    //  OK so, this level could probably be a bit more diffcult. Here's a spikey thing, just click to place it anywhere, just make sure it kills the player
    g.audioPlayer.voiceAudio[1].Play()
    g.toPlace = append(g.toPlace, NewLeftSpike(g, 0, 0))

    // after end
    g.QueueState(ReverseLevel)

    // after reversed
    g.QueueState(func (g *Game){
        //  Nice, Good job.
    // Ok right I've got another spike for you to place. Remember, you can click on things to move them around if it doesn't quite work out
        g.audioPlayer.voiceAudio[2].Play()
        afterReversed(g)
    })
    g.QueueState(StartLevel3)
}

func StartLevel3(g *Game) {
    g.SetPlacing()
    noMoveable(g)

    //g.audioPlayer.voiceAudio[3].Play()

    g.toPlace = append(g.toPlace, NewSpike(g, 0, 0))

    // after end
    g.QueueState(ReverseLevel)
    // after reversed
    g.QueueState(afterReversed)
    g.QueueState(StartLevel4)
}

func StartLevel4(g *Game) {
    g.SetPlacing()
    noMoveable(g)

    g.toPlace = append(g.toPlace, NewSpring(g, 0, 0))
    // Ok lets try something a bit different. here's a spring this time
    g.audioPlayer.voiceAudio[4].Play()

    g.ClearAll()
    tilemap := NewTilemap([][]int{
            {
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  50, 34,  67,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  37, 69,  70,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  39, 37,  38,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  55, 53,  54,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  35, 36, 34, 35, 36, 34, 35, 71, 69, 70, 0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 0,
                0, 0,  51, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 0,
                0, 50, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 55, 53, 54, 0,
                0, 53, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 71, 69, 70, 0,
                0, 69, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 39, 37, 38, 68,
			},
            {
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  98,  82,  115,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  83, 84, 82, 83, 84, 82, 83, 0,  0,  0,  0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 0,
                0, 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 0,
                0, 0,  99, 84, 82, 83, 84, 82, 83, 84, 82, 83, 84, 82, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 0,
                0, 98, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 0,
                0, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 0,
                0, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 135,
			},
		}, 25)

    g.tilemap = &tilemap
    g.tilemap.UpdateSurface()
    g.exit.startx = 18 * tileSize
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
    //noMoveable(g)

    // HMM thats too easy. maybe Try moving the exit ontop of the hill, that could work
    g.audioPlayer.voiceAudio[5].Play()

    //g.toPlace = append(g.toPlace, NewSpring(g, 0, 0))
    g.exit.movable = true
    g.toPlace = append(g.toPlace, g.exit)
    g.RemoveObject(g.exit)
    g.toPlace = append(g.toPlace, NewSpring(g, 0, 0))
    // after end
    g.QueueState(ReverseLevel)
    // after reversed
    g.QueueState(afterReversed)
    g.QueueState(StartLevel6)
}

// How about this?
func StartLevel6(g *Game) {
    g.SetPlacing()
    //noMoveable(g)

    g.audioPlayer.voiceAudio[6].Play()
    g.toPlace = append(g.toPlace, NewSpring(g, 0, 0))
    g.toPlace = append(g.toPlace, NewRightSideSpring(g, 0, 0))
    g.toPlace = append(g.toPlace, NewLeftSideSpring(g, 0, 0))
    g.toPlace = append(g.toPlace, NewBox(g, 0, 0))
    g.exit.movable = true

    // after end
    g.QueueState(ReverseLevel)
    // after reversed
    g.QueueState(afterReversed)
    g.QueueState(StartLevel6)
}

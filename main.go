package main

import (
	_ "embed"
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

)



const (
    screenWidth = 400
    screenHeight = 240
    tileSize = 16
    playerSpeed = 2.1
    jumpHeight = 4
    rewindSpeed = 2
    gravity = 0.16
    friction = 0.75
    airResistance = 0.98

    exitTransitionWeight = 0.9
    ghostAlpha = 0.5
    hightlightBorder = 2
)

var (
	//go:embed shaders/none.kage
	noneShader_src []byte
	//go:embed shaders/vcr.kage
	vcrShader_src []byte
)

var (
    shaders map[string]*ebiten.Shader
    tilesImage *ebiten.Image
)

type State int

const (
	IN_GAME State = iota
	END
	PLACING
    REVERSING
)

type RecPoint struct {
    x float32
    y float32
    vx float32
    vy float32
    alpha float32
}


type Game struct {
    surface *ebiten.Image
    tilemap *Tilemap
    offsetX int
    offsetY int
    player *GameObject
    startPosition *GameObject
    exit *GameObject
    objects []*GameObject
    time int
    animStart int
    shaderName string
    recording [][]RecPoint
    state State
    toPlace []*GameObject
    whenStateFinished []func(g *Game)

    playerAi [][3]bool
    playerAiIdx int
}

func (g * Game)RecordPoint() {
    points := []RecPoint{}
    for _, object := range g.objects {
        points = append(points, RecPoint{
            x: object.x,
            y: object.y,
            vx: object.vx,
            vy: object.vy,
            alpha: object.alpha,
        })
    }
    g.recording = append(g.recording, points)
}

func (g * Game)ReplayPoint() {
    if len(g.recording) == 0 {
        return
    }

    var points []RecPoint
    points, g.recording = g.recording[len(g.recording)-1], g.recording[:len(g.recording)-1]
    for i, point := range points {
        obj := g.objects[i]
        obj.x = point.x
        obj.y = point.y
        obj.vx = point.vx
        obj.vy = point.vy
        obj.alpha = point.alpha
    }
}

func (g * Game)ResetPlayerAi() {
    g.playerAiIdx = 0
    g.player.x = g.player.startx
    g.player.y = g.player.starty
    g.player.vx = 0
    g.player.vy = 0
    g.player.alpha = ghostAlpha
}

func (g * Game)ReplayPlayerAi() {
    if len(g.playerAi) == 0 {
        return
    }

    var state [3]bool
    if g.playerAiIdx >= len(g.playerAi) {
        state = g.playerAi[len(g.playerAi) - 1]
    } else {
        state = g.playerAi[g.playerAiIdx]
    }

    if state[0] {
        g.player.MoveLeft()
    }

    if state[1] {
        g.player.MoveRight()
    }

    if state[2] {
        g.player.Jump()
    }

    g.playerAiIdx += 1
    if g.playerAiIdx >= len(g.playerAi) * 2 {
        g.KillPlayer()
    }

}
func (g * Game) ClearAll() {
    g.objects = g.objects[:0]
    g.objects = append(g.objects, g.player)
    g.objects = append(g.objects, g.exit)
}

func (g * Game) ResetAll() {
    for _, obj := range g.objects {
        obj.x = obj.startx
        obj.y = obj.starty
        obj.vx = 0
        obj.vy = 0
    }
    g.recording = g.recording[:0];
}

func (g *Game) Init() {
    g.surface = ebiten.NewImage(screenWidth, screenHeight)
    g.shaderName = "none"

    g.player = NewPlayer(g, 4 * tileSize, 9 * tileSize)
    g.objects = append(g.objects, g.player)
    g.exit = NewExit(g, 21 * tileSize, 9 * tileSize)
    g.objects = append(g.objects, g.exit)

    g.ResetAll()
    StartLevel1(g)


	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Reverse")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}

}

func (g *Game) Update() error {
    g.time += 1

    //if ebiten.IsKeyJustPressed(ebiten.KeyR) {
    //    if g.state == IN_GAME {
    //        g.state = REVERSING
    //        g.shaderName = "vcr"
    //    }
    //} else {
    //    if g.state == REVERSING {
    //        g.state = IN_GAME
    //        g.shaderName = "none"
    //    }
    //}

    if g.state == IN_GAME {
        if inpututil.IsKeyJustPressed(ebiten.KeyR) {
            g.SetReversing()
            next := func (g *Game){
                g.SetInGame()
                g.recording = g.recording[:0]
                g.playerAi = g.playerAi[:0]
            }
            g.whenStateFinished = append([]func(*Game){next}, g.whenStateFinished...)
        }

        var currentState [3]bool
        if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
            g.player.MoveLeft()
            currentState[0] = true
        } else {
            currentState[0] = false
        }

        if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
            g.player.MoveRight()
            currentState[1] = true
        } else {
            currentState[1] = false
        }

        if  (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyUp)) {
            g.player.Jump()
            currentState[2] = true
        } else {
            currentState[2] = false
        }
        g.playerAi = append(g.playerAi, currentState)

        for _, obj := range g.objects {
            obj.Update(*g.tilemap, g.objects)
        }

        g.tilemap.Update()
        g.RecordPoint()
    }

    if g.state == REVERSING {
        for x := 0; x < rewindSpeed; x++ {
            g.ReplayPoint()
        }

        if len(g.recording) == 0 {
            g.TransitionState()
            fmt.Printf("end of recording state transition\n")
        }
    }

    if g.state == PLACING {
        g.UpdatePlacing()
    }
    if g.player.y > screenHeight {
        g.KillPlayer()
    }


    return nil
}

func (g *Game) UpdatePlacing() {
    for _, obj := range g.objects {
        obj.Update(*g.tilemap, g.objects)
    }

    g.tilemap.Update()
    g.ReplayPlayerAi()

    cx, cy := ebiten.CursorPosition()

    for _, object := range g.objects {
        object.highlight = object.CollidePoint(float32(cx), float32(cy)) && object.movable && len(g.toPlace) == 0
    }

    if len(g.toPlace) > 0 {
        placeable := g.toPlace[0]
        cx = int(math.Floor(float64(cx)/float64(g.tilemap.tileSize)))*g.tilemap.tileSize
        cy = int(math.Floor(float64(cy)/float64(g.tilemap.tileSize)))*g.tilemap.tileSize

        cx += placeable.offsetX
        cy += placeable.offsetY
        placeable.x = float32(cx)
        placeable.y = float32(cy)
        placeable.alpha = float32(math.Abs(math.Sin(float64(g.time) / 30.0)))
    }

    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
        g.PlaceObject(cx, cy)
    }
}

func (g *Game) PlaceObject(cx, cy int) {
        if len(g.toPlace) == 0 {
            object := GetObjectAt(g.objects, float32(cx), float32(cy))
            if object != nil {
                if object.movable {
                    g.toPlace = append([]*GameObject{object}, g.toPlace...)
                    g.RemoveObject(object)
                }
            }
            return
        }

        placeable := g.toPlace[0]
        if placeable.HasCollision(*g.tilemap, g.objects, NONE) {
            return
        }

        placeable.startx = float32(cx)
        placeable.starty = float32(cy)
        placeable.highlight = false
        placeable.alpha = 1.0

        g.objects = append(g.objects, placeable)

        g.toPlace = g.toPlace[1:len(g.toPlace)]

        if len(g.toPlace) == 0 && len(g.playerAi) == 0 {
            g.TransitionState()
        }

}

func (g *Game) Draw(screen *ebiten.Image) {

    g.surface.Fill(color.Alpha16{0x9ccf})

    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(float64(g.offsetX), float64(g.offsetY))

    g.surface.DrawImage(g.tilemap.surface, op)

    for i := len(g.objects)-1; i >= 0; i-- {
        obj := g.objects[i]
        obj.Draw(g.surface, *g.tilemap)
    }

    if g.state == PLACING {
        if len(g.toPlace) > 0 {
            g.toPlace[0].Draw(g.surface, *g.tilemap)
        }
    }

    if g.state == END {

        // draw THE END
        ebitenutil.DebugPrint(screen, fmt.Sprintf("THE END %d", g.time - g.animStart))

        // AFTER THE END
        if g.time > g.animStart + 60 {
            g.TransitionState()
        }
    }

	shop := &ebiten.DrawRectShaderOptions{}
	shop.Uniforms = map[string]any{
        "Time":   float32(g.time) / 60,
        "NoiseOffset":   float32(g.time) / 60,
	}
	shop.Images[0] = g.surface
	shop.Images[1] = g.surface
	shop.Images[2] = g.surface
	shop.Images[3] = g.surface
	screen.DrawRectShader(screenWidth, screenHeight, shaders[g.shaderName], shop)

    //ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
    //screen.DrawImage(surface, &ebiten.DrawImageOptions{})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int){
	return screenWidth, screenHeight
}

func LoadShaders() error {
    if shaders == nil {
		shaders = map[string]*ebiten.Shader{}
	}
    var err error

    shaders["none"], err = ebiten.NewShader([]byte(noneShader_src))
    if err != nil {
        return err
    }

    shaders["vcr"], err = ebiten.NewShader([]byte(vcrShader_src))
    if err != nil {
        return err
    }

    return nil
}
func (g *Game) KillPlayer() {
    if g.state == IN_GAME {
        g.ResetAll()
        g.playerAi = g.playerAi[:0]
    } else {
        g.playerAiIdx = 0
        g.player.x = g.player.startx
        g.player.y = g.player.starty
        g.player.vx = 0
        g.player.vy = 0

        if len(g.toPlace) == 0 {
            g.ResetAll()
            g.playerAi = g.playerAi[:0]
            g.SetInGame()

            for _, o := range g.objects {
                o.movable = false
            }
        }
    }
}

func (g *Game) EndLevel() {
    if g.state == IN_GAME {
        g.state = END
        g.TransitionState()
    } 
    if g.state == PLACING {
        g.ResetPlayerAi()
    }
}

func (g *Game)RemoveObject(obj *GameObject) {

    i := 0 // output index
    for _, x := range g.objects {
        if x != obj {
            // copy and increment index
            g.objects[i] = x
            i++
        }
    }
    // Prevent memory leak by erasing truncated values 
    // (not needed if values don't contain pointers, directly or indirectly)
    for j := i; j < len(g.objects); j++ {
        g.objects[j] = nil
    }
    g.objects = g.objects[:i]
}

func (g *Game) SetReversing() {
    g.state = REVERSING
    g.shaderName = "vcr"
    g.player.alpha = 1.0
}

func (g *Game) SetInGame() {
    g.state = IN_GAME
    g.shaderName = "none"
    g.player.alpha = 1.0
}

func (g *Game) SetPlacing() {
    g.state = PLACING
    g.shaderName = "none"
    g.player.alpha = ghostAlpha
}

func main() {
    LoadShaders()

    var err error
    tilesImage, _, err = ebitenutil.NewImageFromFile("assets/tiles.png")
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowTitle("Hello, World!")
    game := &Game{}
    game.Init()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) TransitionState() {
    if len(g.whenStateFinished) > 0 {
        var function func(*Game)
        g.whenStateFinished, function = g.whenStateFinished[1:len(g.whenStateFinished)], g.whenStateFinished[0]
        function(g)
    }

}

func (g *Game) QueueState(function func(g *Game)) {
    g.whenStateFinished = append(g.whenStateFinished, function)
}

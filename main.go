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
    jumpHeight = 3.2
    rewindSpeed = 2
    gravity = 0.16
    friction = 0.9
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
    shaderName string
    recording [][]RecPoint
    state State
    toPlace []*GameObject
}

func (g * Game)RecordPoint() {
    points := []RecPoint{}
    for _, object := range g.objects {
        points = append(points, RecPoint{
            x: object.x,
            y: object.y,
            vx: object.vx,
            vy: object.vy,
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
        g.objects[i].x = point.x
        g.objects[i].y = point.y
        g.objects[i].vx = point.vx
        g.objects[i].vy = point.vy
    }
}

func (g * Game)ResetAll() {
    for _, obj := range g.objects {
        obj.x = obj.startx
        obj.y = obj.starty
    }
}

func (g *Game) Init() {
    g.state = PLACING

    g.toPlace = append(g.toPlace, NewBox(0, 0))
    g.toPlace = append(g.toPlace, NewBox(0, 0))
    g.toPlace = append(g.toPlace, NewBox(0, 0))
    g.toPlace = append(g.toPlace, NewSpring(0, 0))

    g.surface = ebiten.NewImage(screenWidth, screenHeight)
    g.shaderName = "none"
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
                0, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3, 0,
                0, 4, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 6, 0,
                0, 7, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 9, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		}, 25)

    g.tilemap = &tilemap

    g.tilemap.UpdateSurface()

    g.player = NewPlayer(4 * tileSize, 8 * tileSize)
    g.objects = append(g.objects, g.player)
    g.exit = NewExit(21 * tileSize, 8 * tileSize)
    g.objects = append(g.objects, g.exit)

    g.ResetAll()


	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Reverse")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}

}

func (g *Game) Update() error {
    g.time += 1

    if inpututil.IsKeyJustPressed(ebiten.KeyR) {
        if g.state == REVERSING {
            g.state = IN_GAME
            g.shaderName = "none"
        }

        if g.state == IN_GAME {
            g.state = REVERSING
            g.shaderName = "vcr"
        }

    }

    if g.state == IN_GAME {
        if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
            g.player.vx = -playerSpeed
        }

        if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
            g.player.vx = playerSpeed
        }

        if g.player.onGround && (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyUp)) {
            g.player.vy += -jumpHeight
        }

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
    }

    if g.state == PLACING {
        if len(g.toPlace) > 0 {
            cx, cy := ebiten.CursorPosition()
            cx = int(math.Floor(float64(cx)/float64(g.tilemap.tileSize)))*g.tilemap.tileSize
            cy = int(math.Floor(float64(cy)/float64(g.tilemap.tileSize)))*g.tilemap.tileSize

            g.toPlace[0].x = float32(cx)
            g.toPlace[0].y = float32(cy)

            if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
                g.PlaceObject(cx, cy)
            }
        }
    }

    return nil
}

func (g *Game)PlaceObject(cx, cy int) {
        g.toPlace[0].startx = float32(cx)
        g.toPlace[0].starty = float32(cy)

        g.objects = append(g.objects, g.toPlace[0])

        g.toPlace = g.toPlace[1:len(g.toPlace)]

        if len(g.toPlace) == 0 {
            g.ResetAll()
            g.state = IN_GAME
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

    cx, cy := ebiten.CursorPosition()

	shop := &ebiten.DrawRectShaderOptions{}
	shop.Uniforms = map[string]any{
        "Time":   float32(g.time) / 60,
		"Cursor": []float32{float32(cx), float32(cy)},
	}
	shop.Images[0] = g.surface
	shop.Images[1] = g.surface
	shop.Images[2] = g.surface
	shop.Images[3] = g.surface
	screen.DrawRectShader(screenWidth, screenHeight, shaders[g.shaderName], shop)

    ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
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

func main() {
    LoadShaders()

    var err error
    tilesImage, _, err = ebitenutil.NewImageFromFile("tiles.png")
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

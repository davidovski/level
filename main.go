package main

import (
	"log"
    "image"
    "image/color"
    _ "embed"
    _ "image/png"

	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
    screenWidth = 400
    screenHeight = 240
    tileSize = 16
    playerSpeed = 0.2
    jumpHeight = 0.2
    rewindSpeed = 2
)

var (
	//go:embed shaders/none.kage
	noneShader_src []byte
	//go:embed shaders/vcr.kage
	vcrShader_src []byte

	//go:embed shaders/reverse.kage
	rewindShader_src []byte
)

var (
    shaders map[string]*ebiten.Shader
)

type State int

const (
	IN_GAME State = iota
	END
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

func (g * Game)InitPlayer() {
    g.player = &GameObject{}

    playerImage, _, err := ebitenutil.NewImageFromFile("Assets/Main Characters/Ninja Frog/Idle (32x32).png")
	if err != nil {
		log.Fatal(err)
	}

    g.player.image = playerImage.SubImage(image.Rect(0, 0, 32, 32)).(*ebiten.Image)
    g.objects = append(g.objects, g.player)

    g.ResetPlayer()
}

func (g * Game)ResetPlayer() {
    g.player.x = g.startPosition.x
    g.player.y = g.startPosition.y - 0.1
}

func (g * Game)InitExit() {
    g.exit = &GameObject{
        x: 21,
        y: 8,
    }

    exitImage, _, err := ebitenutil.NewImageFromFile("tiles.png")
	if err != nil {
		log.Fatal(err)
	}

    g.exit.image = exitImage.SubImage(image.Rect(0, 16, 32, 48)).(*ebiten.Image)
    g.objects = append(g.objects, g.exit)
}

func (g *Game) Init() {
    g.state = IN_GAME

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

    g.startPosition = &GameObject{
        x: 4,
        y: 8,
    }

    g.tilemap = &tilemap

    g.tilemap.UpdateSurface()

    g.InitPlayer()
    g.InitExit()


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
        } else  {
            g.state = REVERSING
            g.shaderName = "vcr"
        }

    }

    if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
        g.player.vx = -playerSpeed
    }

    if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
        g.player.vx = playerSpeed
    }

    if g.player.onGround && (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyUp)) {
        g.player.vy = -jumpHeight
    }

    if g.state == IN_GAME {
        for _, obj := range g.objects {
            obj.Update(*g.tilemap)
        }

        g.tilemap.Update()
        g.RecordPoint()
    }

    if g.state == REVERSING {
        for x := 0; x < rewindSpeed; x++ {
            g.ReplayPoint()
        }
    }

    return nil
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

	ebiten.SetWindowTitle("Hello, World!")
    game := &Game{}
    game.Init()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

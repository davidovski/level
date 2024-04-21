package main

import (
	"log"
    "image"
    _ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
    screenWidth = 240
    screenHeight = 240
    tileSize = 16
    playerSpeed = 0.2
    jumpHeight = 0.2
)

type Game struct {
    tilemap *Tilemap
    offsetX int
    offsetY int
    player *GameObject
}

func (g * Game)InitPlayer() {
    g.player = &GameObject{
        x: 5,
        y: 5,
    }

    playerImage, _, err := ebitenutil.NewImageFromFile("Assets/Main Characters/Ninja Frog/Idle (32x32).png")
	if err != nil {
		log.Fatal(err)
	}

    g.player.image = playerImage.SubImage(image.Rect(0, 0, 32, 32)).(*ebiten.Image)
}

func (g *Game) Init() {
    tilemap := NewTilemap([][]int{
            {
				243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
				243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
				243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
				243, 218, 243, 243, 243, 243, 243, 243, 243, 243, 243, 218, 243, 244, 243,
				243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
				243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
				243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
				243, 243, 244, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
				243, 243, 243, 243, 243, 243, 243, 243, 243, 219, 243, 243, 243, 219, 243,
				243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
				243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
				243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
				243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
				243, 218, 243, 243, 243, 243, 243, 243, 243, 243, 243, 244, 243, 243, 243,
				243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243,
			},
			{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7,
				7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 29,
				29, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 51,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		}, 15)
    g.tilemap = &tilemap

    g.tilemap.UpdateSurface()

    g.InitPlayer()


	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Tiles (Ebitengine Demo)")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}

}

func (g *Game) Update() error {
    if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
        g.player.vx = -playerSpeed
    }

    if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
        g.player.vx = playerSpeed
    }

    if g.player.onGround && (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyUp)) {
        g.player.vy = -jumpHeight
    }

    g.player.Update(*g.tilemap)
    return g.tilemap.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(float64(g.offsetX), float64(g.offsetY))

    screen.DrawImage(g.tilemap.surface, op)
    g.player.Draw(screen, *g.tilemap)
	ebitenutil.DebugPrint(screen, "Hello, World!")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int){
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowTitle("Hello, World!")
    game := &Game{}
    game.Init()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

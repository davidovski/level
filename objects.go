package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
    SPRING_FORCE = 10
)

type Direction int

const (
	NONE Direction = iota
	LEFT
	RIGHT
    UP
    DOWN
)

type GameObject struct {
    game *Game
    startx, starty float32
    x, y float32
    vx, vy float32

    offsetX, offsetY int
    alpha float32
    highlight bool

    friction float32
    resistance float32
    image *ebiten.Image
    onGround bool
    onCollideUp func(this, other *GameObject) bool
    onCollideDown func(this, other *GameObject) bool
    onCollideLeft func(this, other *GameObject) bool
    onCollideRight func(this, other *GameObject) bool
    Draw func (o * GameObject, screen *ebiten.Image, tilemap Tilemap)
    movable bool
}


type Player struct {
    GameObject
}

func (o * GameObject) Update(tilemap Tilemap, others []*GameObject) {
    var direction Direction
    o.vy += gravity

    o.vx *= o.resistance
    o.vy *= o.resistance

    o.x += o.vx

    if o.vx > 0 {
        direction = RIGHT
    } else {
        direction = LEFT
    }

    if o.HasCollision(tilemap, others, direction) {
        o.x -= o.vx
        o.vx = 0
    }

    o.y += o.vy

    if o.vy > 0 {
        direction = UP
    } else {
        direction = DOWN
    }

    if o.HasCollision(tilemap, others, direction) {
        o.onGround = true;
        o.vx *= o.friction

        o.y -= o.vy
        o.vy = 0
    } else {
        o.onGround = false;
    }
}
func GetObjectAt(objects []*GameObject, x, y float32) *GameObject {

    for _, object := range objects {
        if object.CollidePoint(x, y) {
            return object
        }
    }
    return nil
}

func (object * GameObject) CollidePoint(x, y float32) bool {
        maxX := object.x + float32(object.image.Bounds().Dx())
        maxY := object.y + float32(object.image.Bounds().Dy())
        minX := object.x
        minY := object.y
        return x >= minX && x < maxX && y >= minY && y < maxY
    }

func (o * GameObject) HasCollision(tilemap Tilemap, others []*GameObject, dir Direction) bool {
    if tilemap.CollideObject(o) {
        return true
    }
    for _, obj := range others {
        if obj.Collide(o) {
            var f func(this, other *GameObject) bool
            switch dir {
                case UP:
                    f = obj.onCollideUp
                case DOWN:
                    f = obj.onCollideDown
                case LEFT:
                    f = obj.onCollideLeft
                case RIGHT:
                    f = obj.onCollideRight
            }

            if f == nil{
                return true
            }
            return f(obj, o)
        }
    }

    return false
}

func DrawObject(o * GameObject, screen *ebiten.Image, tilemap Tilemap) {
    op := &ebiten.DrawImageOptions{}
    op.ColorScale.ScaleAlpha(o.alpha)
    op.GeoM.Translate(float64(o.x), float64(o.y))
    screen.DrawImage(o.image, op)

    if o.highlight {
        vector.StrokeRect(screen, o.x, o.y, float32(o.image.Bounds().Dx()), float32(o.image.Bounds().Dy()), hightlightBorder, color.RGBA{255, 100, 100, 255}, false)
    }
}

func (object * GameObject) Collide(other *GameObject) bool {
    maxX1 := object.x + float32(object.image.Bounds().Dx())
    maxY1 := object.y + float32(object.image.Bounds().Dy())
    minX1 := object.x
    minY1 := object.y

    maxX2 := other.x + float32(other.image.Bounds().Dx())
    maxY2 := other.y + float32(other.image.Bounds().Dy())
    minX2 := other.x
    minY2 := other.y

    if (minX1 == minX2 && maxX1 == maxX2 && minY1 == minY2 && maxY1 == maxY2) {
        return false
    }

    return ! ( minX2 >= maxX1 || maxX2 <= minX1 || minY2 >= maxY1 || maxY2 <= minY1)
}

func (object * GameObject) Jump() {
    if object.onGround {
        object.vy += -jumpHeight
    }
}

func (object * GameObject) MoveLeft() {
    object.vx = -playerSpeed
}
func (object * GameObject) MoveRight() {
    object.vx = playerSpeed
}


func (object * GameObject) ToRect() image.Rectangle {
    width := object.image.Bounds().Dx()
    height := object.image.Bounds().Dy()
    x := int(object.x)
    y := int(object.y)
    return image.Rect(x, y, x+width, y+height)
}

func OnCollideGeneric(this, other *GameObject) bool {
    return true
}

func NewObject(game *Game, x, y float32) *GameObject{
    return &GameObject{
        game: game,
        startx: x,
        starty: y,
        alpha: 1.0,
        highlight: false,
        friction: friction,
        resistance: airResistance,
        x: x,
        y: y,
        movable: true,
        Draw: DrawObject,
    }
}

func NewPlayer(game *Game, x, y float32) *GameObject{
    player := NewObject(game, x, y)

    playerImage := ebiten.NewImageFromImage(characterImage)

    player.image = playerImage.SubImage(image.Rect(4, 8, 27, 32)).(*ebiten.Image)
    player.Draw = DrawObject

    player.movable = false
    return player
}

func NewExit(game *Game, x, y float32) *GameObject{
    exit := NewObject(game, x, y)

    exit.image = tilesImage.SubImage(image.Rect(0, 16, 32, 48)).(*ebiten.Image)
    exit.onCollideUp = OnCollideExit
    exit.onCollideDown = OnCollideExit
    exit.onCollideLeft = OnCollideExit
    exit.onCollideRight = OnCollideExit

    exit.movable = false
    return exit
}

func NewBox(game *Game, x, y float32) *GameObject{
    box := NewObject(game, x, y)

    box.image = tilesImage.SubImage(image.Rect(160, 0, 176, 16)).(*ebiten.Image)

    return box
}

func NewSpring(game *Game, x, y float32) *GameObject{
    spring := NewObject(game, x, y)

    spring.image = tilesImage.SubImage(image.Rect(176, 0, 192, 16)).(*ebiten.Image)
    spring.onCollideUp = OnCollideSpring

    return spring
}

func NewSpike(game *Game, x, y float32) *GameObject{
    spike := NewObject(game, x, y)

    spike.offsetY = 3
    spike.image = tilesImage.SubImage(image.Rect(192, 3, 208, 16)).(*ebiten.Image)
    spike.onCollideUp = OnCollideSpike

    return spike
}

func NewLeftSpike(game *Game, x, y float32) *GameObject{
    spike := NewObject(game, x, y)

    spike.offsetX = 3
    spike.image = tilesImage.SubImage(image.Rect(195, 16, 208, 32)).(*ebiten.Image)
    spike.onCollideRight = OnCollideSpike

    return spike
}

func OnCollideExit(this, other *GameObject) bool {
    if other == this.game.player {
        g := other.game
        g.player.x = (exitTransitionWeight)*g.player.x + (1-exitTransitionWeight)*g.exit.x
        g.player.y = (exitTransitionWeight)*g.player.y + (1-exitTransitionWeight)*g.exit.y
        g.player.alpha *= exitTransitionWeight

        if g.player.alpha < 0.001 {
            g.EndLevel()
        }
    }
    return false
}

func OnCollideSpring(this, other *GameObject) bool {
    other.vy -= SPRING_FORCE
    other.onGround = true
    return false
}

func OnCollideSpike(this, other *GameObject) bool {
    if other == this.game.player {
        other.game.KillPlayer()
    }
    return true
}


package main

import (
	"log"
    "image"
    "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
    SPRING_FORCE = 8
)

type GameObject struct {
    startx, starty float32
    x, y float32
    vx, vy float32

    offsetX, offsetY int

    friction float32
    resistance float32
    image *ebiten.Image
    onGround bool
    onCollideX func(this, other *GameObject) bool
    onCollideY func(this, other *GameObject) bool
}


type Player struct {
    GameObject
}

func (o * GameObject) Update(tilemap Tilemap, others []*GameObject) {
    o.vy += gravity

    o.vx *= o.resistance
    o.vy *= o.resistance

    o.x += o.vx
    if o.HasCollision(tilemap, others, true) {
        o.x -= o.vx
        o.vx = 0
    }

    o.y += o.vy
    if o.HasCollision(tilemap, others, false) {
        o.onGround = true;
        o.vx *= o.friction

        o.y -= o.vy
        o.vy = 0
    } else {
        o.onGround = false;
    }
}

func (o * GameObject) HasCollision(tilemap Tilemap, others []*GameObject, x bool) bool {
    if tilemap.CollideObject(o) {
        return true
    }
    for _, obj := range others {
        if obj.Collide(o) {
            var f func(this, other *GameObject) bool
            if x {
                f = obj.onCollideX
            } else {
                f = obj.onCollideY
            }

            if f == nil{
                return true
            }
            return f(obj, o)
        }
    }
    return false
}

func (o * GameObject) Draw(screen *ebiten.Image, tilemap Tilemap) {
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(float64(o.x), float64(o.y))
    screen.DrawImage(o.image, op)
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

func NewObject(x, y float32) *GameObject{
    obj := &GameObject{
        startx: x,
        starty: y,
    }

    obj.friction = friction
    obj.resistance = airResistance
    obj.x = obj.startx
    obj.y = obj.starty
    return obj
}

func NewPlayer(x, y float32) *GameObject{
    player := NewObject(x, y)

    playerImage, _, err := ebitenutil.NewImageFromFile("Assets/Main Characters/Ninja Frog/Idle (32x32).png")
	if err != nil {
		log.Fatal(err)
	}

    player.image = playerImage.SubImage(image.Rect(4, 8, 27, 32)).(*ebiten.Image)

    return player
}

func NewExit(x, y float32) *GameObject{
    exit := NewObject(x, y)

    exit.image = tilesImage.SubImage(image.Rect(0, 16, 32, 48)).(*ebiten.Image)

    return exit
}

func NewBox(x, y float32) *GameObject{
    box := NewObject(x, y)

    box.image = tilesImage.SubImage(image.Rect(160, 0, 176, 16)).(*ebiten.Image)

    return box
}

func NewSpring(x, y float32) *GameObject{
    spring := NewObject(x, y)

    spring.image = tilesImage.SubImage(image.Rect(176, 0, 192, 16)).(*ebiten.Image)
    spring.onCollideY = OnCollideSpring

    return spring
}

func NewSpike(x, y float32) *GameObject{
    spike := NewObject(x, y)

    spike.offsetY = 3
    spike.image = tilesImage.SubImage(image.Rect(192, 3, 208, 16)).(*ebiten.Image)
    spike.onCollideY = OnCollideSpike

    return spike
}

func NewLeftSpike(x, y float32) *GameObject{
    spike := NewObject(x, y)

    spike.offsetX = 3
    spike.image = tilesImage.SubImage(image.Rect(195, 16, 208, 32)).(*ebiten.Image)
    spike.onCollideX = OnCollideSpike

    return spike
}

func OnCollideSpring(this, other *GameObject) bool {
    other.vy -= SPRING_FORCE
    other.onGround = true
    return false
}

func OnCollideSpike(this, other *GameObject) bool {
    other.x = other.startx
    other.y = other.starty
    return true
}


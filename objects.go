package main

import (
	"math"
	"image"
	"image/color"

    "math/rand/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
    SPRING_FORCE = 8
    SIDE_SPRING_FORCE = 8
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
    images []*ebiten.Image
    onGround bool
    onCollideUp func(this, other *GameObject) bool
    onCollideDown func(this, other *GameObject) bool
    onCollideLeft func(this, other *GameObject) bool
    onCollideRight func(this, other *GameObject) bool
    Draw func (o * GameObject, screen *ebiten.Image, tilemap Tilemap)
    UpdateFunc func (o * GameObject, tilemap Tilemap, others []*GameObject) bool
    movable bool
    state int
    delta int
}


type Player struct {
    GameObject
}

func UpdateHPlatform(o * GameObject, tilemap Tilemap, others []*GameObject) bool {
    o.delta += 1
    if int(o.delta / 128) % 2 == 0 {
        o.vx = -0.4
    } else {
        o.vx = 0.4
    }
    o.x += o.vx
    o.vy = 0
    return true
}

func UpdateVPlatform(o * GameObject, tilemap Tilemap, others []*GameObject) bool {
    o.delta += 1
    if int(o.delta / 128) % 2 == 0 {
        o.vy = -0.4
    } else {
        o.vy = 0.4
    }
    o.y += o.vy
    return true
}

func UpdateSpring(o * GameObject, tilemap Tilemap, others []*GameObject) bool {
    o.delta += 1
    if o.delta % 3 == 0 {
        if o.state == 1 {
            o.state = 0
        } else if o.state == 2 {
            o.state = 1
        }
    }
    return false
}

func UpdatePlayer(o * GameObject, tilemap Tilemap, others []*GameObject) bool {
    var d float32 = 8.0 
    var t float32 = 1.0
    // 0
    if o.onGround {
        if o.vx > t {
            o.state = (int(o.x/d) % 6) + 1
        } else if o.vx < -t {
            o.state = (int(o.x/d) % 6) + 16
        } else {
            if o.vx < 0 {
                o.state = 15
            } else {
                o.state = 0
            }
        }
    } else {
        if o.vx == 0 {

            if o.state >= 15 {
                o.state = (int(math.Abs(float64(o.vy))) % 5) + 25
            } else {
                o.state = (int(math.Abs(float64(o.vy))) % 5) + 10
            }
        } else if o.vx > 0 {
            o.state = (int(math.Abs(float64(o.vy))) % 5) + 10
        } else {
            o.state = (int(math.Abs(float64(o.vy))) % 5) + 25 
        }
    }
    return false
}

func (o * GameObject) Update(tilemap Tilemap, others []*GameObject) {
    var direction Direction
    o.vy += gravity

    o.vx *= o.resistance
    o.vy *= o.resistance

    if o.UpdateFunc != nil {
        if o.UpdateFunc(o, tilemap, others) {
            return
        }
    }

    o.x += o.vx
    

    if o.vx < 0 {
        direction = LEFT
    } else {
        direction = RIGHT
    }

    if o.HasCollision(tilemap, others, direction) {
        o.x -= o.vx
        o.vx = 0
    }

    o.y += o.vy

    if o.vy > 0 {
        direction = DOWN
    } else {
        direction = UP
    }

    if o.HasCollision(tilemap, others, direction) {
        if ! o.onGround && o.vy > gravity*12 {
            o.PlayLand()
        }
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
        maxX := object.x + float32(object.images[0].Bounds().Dx())
        maxY := object.y + float32(object.images[0].Bounds().Dy())
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
                case DOWN:
                    f = obj.onCollideUp
                case UP:
                    f = obj.onCollideDown
                case RIGHT:
                    f = obj.onCollideLeft
                case LEFT:
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
func ShadowDraw(screen *ebiten.Image, image *ebiten.Image, x, y float32, alpha float32) {
    ShadowDrawOffset(screen, image, x, y, alpha, shadowOffset)
    }


func ShadowDrawOffset(screen *ebiten.Image, image *ebiten.Image, x, y, alpha, offset float32) {
    op := &ebiten.DrawImageOptions{}
    if alpha > 0 {
        op = &ebiten.DrawImageOptions{}
        op.ColorScale.ScaleAlpha(alpha)
        op.ColorScale.Scale(0, 0, 0, 1);
        op.GeoM.Translate(float64(x+offset), float64(y + offset))
        screen.DrawImage(image, op)
    }
    op = &ebiten.DrawImageOptions{}
    op.ColorScale.ScaleAlpha(alpha)
    op.GeoM.Translate(float64(x), float64(y))
    screen.DrawImage(image, op)
}

func DrawObject(o * GameObject, screen *ebiten.Image, tilemap Tilemap) {
    image := o.images[o.state]
    ShadowDraw(screen, image, o.x, o.y, o.alpha)

    if o.highlight {
        vector.StrokeRect(screen, o.x, o.y, float32(image.Bounds().Dx()), float32(image.Bounds().Dy()), hightlightBorder, color.RGBA{255, 100, 100, 255}, false)
    }
}

func (object * GameObject) Collide(other *GameObject) bool {
    maxX1 := object.x + float32(object.images[0].Bounds().Dx())
    maxY1 := object.y + float32(object.images[0].Bounds().Dy())
    minX1 := object.x
    minY1 := object.y

    maxX2 := other.x + float32(other.images[0].Bounds().Dx())
    maxY2 := other.y + float32(other.images[0].Bounds().Dy())
    minX2 := other.x
    minY2 := other.y

    if (minX1 == minX2 && maxX1 == maxX2 && minY1 == minY2 && maxY1 == maxY2) {
        return false
    }

    return ! ( minX2 >= maxX1 || maxX2 <= minX1 || minY2 >= maxY1 || maxY2 <= minY1)
}

func (object *GameObject) PlayLand() {
    jumpid := rand.IntN(2)
    object.game.audioPlayer.landAudio[jumpid].Rewind()
    object.game.audioPlayer.landAudio[jumpid].Play()
}

func (object *GameObject) playJump() {
    jumpid := rand.IntN(2)
    object.game.audioPlayer.jumpAudio[jumpid].Rewind()
    object.game.audioPlayer.jumpAudio[jumpid].Play()
}

func (object * GameObject) Jump() {
    if object.onGround {
        object.playJump()
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
    width := object.images[0].Bounds().Dx()
    height := object.images[0].Bounds().Dy()
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

    player.images = []*ebiten.Image{
        playerImage.SubImage(image.Rect(0, 0, 17, 28)).(*ebiten.Image),

        // run
        playerImage.SubImage(image.Rect(0, 30, 17, 57)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(21, 30, 38, 58)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(42, 30, 59, 58)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(63, 30, 81, 57)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(84, 30, 101, 58)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(105, 30, 122, 58)).(*ebiten.Image),

        // jump
        playerImage.SubImage(image.Rect(0, 60, 17, 88)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(21, 60, 41, 86)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(42, 60, 63, 86)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(63, 60, 80, 90)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(84, 60, 102, 89)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(105, 60, 125, 89)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(126, 60, 148, 87)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(148, 60, 166, 86)).(*ebiten.Image),

        //left
        playerImage.SubImage(image.Rect(0, 90, 17, 118)).(*ebiten.Image),

        // run
        playerImage.SubImage(image.Rect(0, 120, 17, 147)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(21, 120, 38, 148)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(42, 120, 59, 148)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(63, 120, 81, 147)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(84, 120, 101, 148)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(105, 120, 122, 148)).(*ebiten.Image),

        // jump
        playerImage.SubImage(image.Rect(0, 150, 17, 178)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(21, 150, 41, 176)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(42, 150, 63, 176)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(63, 150, 80, 180)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(84, 150, 102, 179)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(105, 150, 125, 179)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(126, 150, 148, 177)).(*ebiten.Image),
        playerImage.SubImage(image.Rect(148, 150, 166, 176)).(*ebiten.Image),
    }

    player.Draw = DrawObject
    player.UpdateFunc = UpdatePlayer

    player.movable = false
    return player
}

func NewExit(game *Game, x, y float32) *GameObject{
    exit := NewObject(game, x, y)

    exit.images = []*ebiten.Image{
        tilesImage.SubImage(image.Rect(0, 16, 32, 48)).(*ebiten.Image),
    }
    exit.onCollideUp = OnCollideExit
    exit.onCollideDown = OnCollideExit
    exit.onCollideLeft = OnCollideExit
    exit.onCollideRight = OnCollideExit

    exit.movable = false
    return exit
}

func NewBox(game *Game, x, y float32) *GameObject{
    box := NewObject(game, x, y)

    box.images = []*ebiten.Image{
        tilesImage.SubImage(image.Rect(160, 0, 176, 16)).(*ebiten.Image),
    }

    return box
}

func NewSpring(game *Game, x, y float32) *GameObject{
    spring := NewObject(game, x, y)

    spring.images = []*ebiten.Image{
        tilesImage.SubImage(image.Rect(176, 0, 192, 16)).(*ebiten.Image),
        tilesImage.SubImage(image.Rect(176, 16, 192, 32)).(*ebiten.Image),
        tilesImage.SubImage(image.Rect(176, 32, 192, 48)).(*ebiten.Image),
    }
    spring.onCollideUp = OnCollideTopSpring
    spring.UpdateFunc = UpdateSpring

    return spring
}

func NewRightSideSpring(game *Game, x, y float32) *GameObject{
    spring := NewObject(game, x, y)

    spring.images = []*ebiten.Image{
        tilesImage.SubImage(image.Rect(176, 48, 192, 64)).(*ebiten.Image),
        tilesImage.SubImage(image.Rect(176, 64, 192, 80)).(*ebiten.Image),
        tilesImage.SubImage(image.Rect(176, 80, 192, 96)).(*ebiten.Image),
    }
    spring.onCollideRight = OnCollideRightSideSpring
    spring.UpdateFunc = UpdateSpring

    return spring
}

func NewLeftSideSpring(game *Game, x, y float32) *GameObject{
    spring := NewObject(game, x, y)

    spring.images = []*ebiten.Image{
        tilesImage.SubImage(image.Rect(176, 96, 192, 112)).(*ebiten.Image),
        tilesImage.SubImage(image.Rect(176, 112, 192, 128)).(*ebiten.Image),
        tilesImage.SubImage(image.Rect(176, 128, 192, 144)).(*ebiten.Image),
    }
    spring.onCollideLeft = OnCollideLeftSideSpring
    spring.UpdateFunc = UpdateSpring

    return spring
}

func NewSpike(game *Game, x, y float32) *GameObject{
    spike := NewObject(game, x, y)

    spike.offsetY = 3
    spike.images = []*ebiten.Image{
        tilesImage.SubImage(image.Rect(192, 3, 208, 16)).(*ebiten.Image),
    }

    spike.onCollideUp = OnCollideSpike

    return spike
}

func NewLeftSpike(game *Game, x, y float32) *GameObject{
    spike := NewObject(game, x, y)

    spike.offsetX = 3
    spike.images = []*ebiten.Image{
        tilesImage.SubImage(image.Rect(195, 16, 208, 32)).(*ebiten.Image),
    }
    spike.onCollideLeft = OnCollideSpike

    return spike
}

func NewVPlatform(game *Game, x, y float32) *GameObject{
    platform := NewObject(game, x, y)

    platform.images = []*ebiten.Image{
        tilesImage.SubImage(image.Rect(208, 0, 224, 5)).(*ebiten.Image),
    }
    platform.onCollideUp = OnCollideVPlatformTop
    platform.onCollideDown = OnCollideVPlatformBottom
    platform.UpdateFunc = UpdateVPlatform

    return platform
}

func NewHPlatform(game *Game, x, y float32) *GameObject{
    platform := NewObject(game, x, y)

    platform.images = []*ebiten.Image{
        tilesImage.SubImage(image.Rect(208, 0, 224, 5)).(*ebiten.Image),
    }
    platform.onCollideUp = OnCollideVPlatformTop
    platform.UpdateFunc = UpdateHPlatform

    return platform
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
    other.onGround = true
    this.state = 2
    this.delta = 1
    this.game.audioPlayer.springAudio.Rewind()
    this.game.audioPlayer.springAudio.Play()
    return false
}

func OnCollideTopSpring(this, other *GameObject) bool {
    other.vy = -SPRING_FORCE
    other.y += other.vy

    return OnCollideSpring(this, other)
}

func OnCollideLeftSideSpring(this, other *GameObject) bool {
    other.vx = -SPRING_FORCE
    other.x += other.vx
    return OnCollideSpring(this, other)
}

func OnCollideRightSideSpring(this, other *GameObject) bool {
    other.vx = SPRING_FORCE
    other.x += other.vx
    return OnCollideSpring(this, other)
}

func OnCollideSpike(this, other *GameObject) bool {
    if other == this.game.player {
        other.game.KillPlayer()
    }
    return true
}

func OnCollideVPlatformTop(this, other *GameObject) bool {
    other.y = this.y - float32(other.images[0].Bounds().Dy()) + this.vy
    other.vy = this.vy
    other.x += this.vx
    return true
}
func OnCollideVPlatformBottom(this, other *GameObject) bool {
    other.y = this.y + float32(this.images[0].Bounds().Dy())
    other.vy = this.vy
    return false
}

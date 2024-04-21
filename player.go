package main

import (
    "github.com/hajimehoshi/ebiten/v2"
)

const (
    gravity = 0.01
    friction = 0.6
)

type GameObject struct {
    x, y float32
    vx, vy float32
    image *ebiten.Image
    onGround bool
}


type Player struct {
    GameObject
}

func (o * GameObject) Update(tilemap Tilemap) {
    o.vy += gravity

    o.x += o.vx
    if tilemap.CollideObject(o) {
        o.x -= o.vx
        o.vx = 0
    }

    o.y += o.vy
    if (tilemap.CollideObject(o)) {

        o.onGround = true;
        o.vx *= friction


        o.y -= o.vy
        o.vy = 0
    } else {
        o.onGround = false;
    }
}

func (o * GameObject) Draw(screen *ebiten.Image, tilemap Tilemap) {
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(float64(o.x * float32(tilemap.tileSize)), float64(o.y * float32(tilemap.tileSize)))
    screen.DrawImage(o.image, op)
}



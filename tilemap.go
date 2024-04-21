package main

import (
    "github.com/hajimehoshi/ebiten/v2"
    "log"
    "image"
    _ "image/png"

)

type Tilemap struct {
    tilesImage *ebiten.Image
    surface *ebiten.Image
    mapWidth int
    tileSize int
    layers [][]int
    collisions []image.Rectangle
}


// take in map coord and make it real life coord
func (t *Tilemap) Translate(x, y int) (int, int) {
    return x * t.tileSize, y*t.tileSize
}

func (* Tilemap) Update() error {
	return nil
}

func (tm * Tilemap) Draw(screen *ebiten.Image) {
}

func (tm *Tilemap) UpdateSurface() {
    w := tm.tilesImage.Bounds().Dx()
	tileXCount := w / tileSize

	// Draw each tile with each DrawImage call.
	// As the source images of all DrawImage calls are always same,
	// this rendering is done very efficiently.
	// For more detail, see https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#Image.DrawImage
	for _, l := range tm.layers {
		for i, t := range l {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64((i%tm.mapWidth)*tm.tileSize), float64((i/tm.mapWidth)*tm.tileSize))

			sx := (t % tileXCount) * tileSize
			sy := (t / tileXCount) * tileSize
			tm.surface.DrawImage(tm.tilesImage.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image), op)
		}
	}

    //ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func NewTilemap(layers [][]int, mapWidth int) Tilemap {
    tilemap := Tilemap{
        tileSize: 16,
        mapWidth: 25,
    }

    tilemap.layers = layers

    tilemap.surface = ebiten.NewImage(mapWidth*tilemap.tileSize, len(layers[0])/mapWidth*tilemap.tileSize)

    var err error
	tilemap.tilesImage = tilesImage

	if err != nil {
		log.Fatal(err)
	}

    tilemap.CalculateCollisions()

    return tilemap
}

func (tm * Tilemap) CalculateCollisions() {
    for i, t := range tm.layers[0] {
        if t != 0 {
            x := i%tm.mapWidth * tm.tileSize
            y := i/tm.mapWidth * tm.tileSize

            w, h := tm.tileSize, tm.tileSize
            rect := image.Rect(x, y, x+w, y+h)
            tm.collisions = append(tm.collisions, rect)
            
        }
    }
}


func Collide(r1 image.Rectangle, r2 image.Rectangle) bool {
    return ! ( r2.Min.X > r1.Max.X || r2.Max.X < r1.Min.X || r2.Min.Y > r1.Max.Y || r2.Max.Y < r1.Min.Y)
}

func (t * Tilemap) Collide(x, y, width, height int) bool {
    r1 :=  image.Rect(
        x,
        y,
        x+width,
        y+height,
    )

    for _, r2 := range t.collisions {
        if ! ( r2.Min.X > r1.Max.X || r2.Max.X < r1.Min.X || r2.Min.Y > r1.Max.Y || r2.Max.Y < r1.Min.Y) {
            return true

        }
    }
    return false
    
}


func (t * Tilemap) CollideObject(object *GameObject) bool {
    width := object.image.Bounds().Dx()
    height := object.image.Bounds().Dy()
    x := int(object.x)
    y := int(object.y)
    return t.Collide(x, y, width, height)
}


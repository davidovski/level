package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
    "github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
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

    endCardDuration = 200

    exitTransitionWeight = 0.8
    ghostAlpha = 0.5
    hightlightBorder = 2
    audioFadeIn = 0.999

    sampleRate = 44100
    shadowOffset = 1
)

var (
	//go:embed shaders/none.kage
	noneShader_src []byte
	//go:embed shaders/vcr.kage
	vcrShader_src []byte
	//go:embed shaders/clouds.kage
	cloudShader_src []byte
	//go:embed shaders/bloom.kage
	bloomShader_src []byte

	//go:embed assets/tiles.png
	tilesPng_src []byte

	//go:embed assets/character.png
	characterPng_src []byte

	//go:embed assets/rewind.wav
	rewindWav_src []byte

	//go:embed assets/stop.wav
	stopWav_src []byte

	//go:embed assets/start.wav
	startWav_src []byte

	//go:embed assets/ambient.ogg
	ambientOgg_src []byte
)

var (
    shaders map[string]*ebiten.Shader
    tilesImage *ebiten.Image
    characterImage *ebiten.Image
    fontFaceSource *text.GoTextFaceSource
)

type State int

const (
	IN_GAME State = iota
	END
    PLACING
	PAUSED
    REVERSING
)

type RecPoint struct {
    x float32
    y float32
    vx float32
    vy float32
    alpha float32
}

type AudioPlayer struct {
    audioContext *audio.Context
    rewindAudio  *audio.Player
    stopAudio  *audio.Player
    startAudio  *audio.Player
    ambientAudio  *audio.Player
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

    audioPlayer *AudioPlayer
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
    g.shaderName = "sky"

    g.player = NewPlayer(g, 4 * tileSize, 9 * tileSize)
    g.objects = append(g.objects, g.player)
    g.exit = NewExit(g, 21 * tileSize, 9 * tileSize)
    g.objects = append(g.objects, g.exit)

    g.ResetAll()
    StartGame(g)
    g.audioPlayer.ambientAudio.SetVolume(0)


	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Reverse")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}

}

func (g *Game) Update() error {
    g.time += 1


    if g.state == IN_GAME || g.state == PLACING {
        if ! g.audioPlayer.ambientAudio.IsPlaying() {
            g.audioPlayer.ambientAudio.Play()
        }
        volume := g.audioPlayer.ambientAudio.Volume()
        volume = 1-((1-volume)*audioFadeIn)
        g.audioPlayer.ambientAudio.SetVolume(volume)
    }

    if g.state == PAUSED && inpututil.IsKeyJustPressed(ebiten.KeyR) {
        g.TransitionState()
    }

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
        g.time -= 1 + rewindSpeed;
        if g.time < g.animStart || g.animStart <= 0{
            for x := 0; x < rewindSpeed; x++ {
                g.ReplayPoint()
            }

            if len(g.recording) == 0 {
                g.TransitionState()
            }
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

func DrawBackground(screen *ebiten.Image, time int)  {
	shop := &ebiten.DrawRectShaderOptions{}
	shop.Uniforms = map[string]any{
        "Time":   float32(time) / 60,
	}
	screen.DrawRectShader(screenWidth, screenHeight, shaders["sky"], shop)
}

func PostProcess(screen *ebiten.Image, shaderName string, time int) {
    w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
    for _, shader := range []string{shaderName} {
        out := ebiten.NewImage(w, h)
        shop := &ebiten.DrawRectShaderOptions{}

        shop.Uniforms = map[string]any{
            "Time":   float32(time) / 60,
            "NoiseOffset":   float32(time) / 60,
        }
        shop.Images[0] = screen
        shop.Images[1] = screen
        shop.Images[2] = screen
        shop.Images[3] = screen
        out.DrawRectShader(w, h, shaders[shader], shop)

        op := &ebiten.DrawImageOptions{}
        screen.DrawImage(out, op)
    }

}

func (g *Game) DrawTheEnd(surface *ebiten.Image, alpha float32) {
    textSize := 30.0

    msg := fmt.Sprintf("THE END")
    textOp := &text.DrawOptions{}
    textOp.GeoM.Translate((screenWidth - textSize*7 ) / 2, (screenHeight - textSize) / 2)
    textOp.ColorScale.ScaleWithColor(color.White)
    textOp.ColorScale.ScaleAlpha(alpha)
    text.Draw(surface, msg, &text.GoTextFace{
        Size:   textSize,
        Source: fontFaceSource,
    }, textOp)

}

func (g *Game) IsShowTheEnd() bool {
    return g.animStart > 0 && (g.state == END || g.state == PAUSED || g.state == REVERSING)

}

func (g *Game) Draw(screen *ebiten.Image) {

    DrawBackground(screen, g.time)

    g.surface.Fill(color.RGBA{0, 0, 0, 0})
    g.tilemap.Draw(g.surface, 0, -2)

    for i := len(g.objects)-1; i >= 0; i-- {
        obj := g.objects[i]
        obj.Draw(obj, g.surface, *g.tilemap)
    }

    if g.state == PLACING {
        if len(g.toPlace) > 0 {
            g.toPlace[0].Draw(g.toPlace[0], g.surface, *g.tilemap)
        }
    }

    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(float64(g.offsetX), float64(g.offsetY-2))
    if g.IsShowTheEnd() {

        // draw THE END
        a :=  float64(endCardDuration - (g.time - g.animStart)) / float64(endCardDuration);
        a = 1 - float64(math.Pow(float64(a), 10))

        if g.state == PAUSED {
            a = 10.0
        }

        if a < 0.0 {
            a = 0.0
        }

        op.GeoM.Translate(float64(g.offsetX), float64(g.offsetY) + a*screenHeight/2)

        screen.DrawImage(g.surface, op)
        g.DrawTheEnd(screen, float32(a))

        // AFTER THE END
        if g.state == END && g.time - g.animStart  > endCardDuration {
            g.TransitionState()
        }
    } else {
        screen.DrawImage(g.surface, op)
    }


    PostProcess(screen, g.shaderName, g.time)

    ebitenutil.DebugPrint(screen, fmt.Sprintf("tps: %.4f", ebiten.ActualFPS()))
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
    shaders["bloom"], err = ebiten.NewShader([]byte(bloomShader_src))
    if err != nil {
        return err
    }

    shaders["sky"], err = ebiten.NewShader([]byte(cloudShader_src))
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

func (g *Game) SetPaused() {
    g.state = PAUSED
    g.shaderName = "vcr"
    //g.player.alpha = 1.0
    g.audioPlayer.startAudio.Rewind()
    g.audioPlayer.startAudio.Play()
    g.audioPlayer.ambientAudio.Pause()
}
func (g *Game) SetReversing() {
    g.state = REVERSING
    g.shaderName = "vcr"
    g.player.alpha = 1.0
    g.audioPlayer.rewindAudio.Rewind()
    g.audioPlayer.rewindAudio.Play()
    g.audioPlayer.ambientAudio.Pause()
}

func (g *Game) StopRewinding() {
    g.shaderName = "none"
    if g.audioPlayer.rewindAudio.IsPlaying() {
        g.audioPlayer.rewindAudio.Pause()
        g.audioPlayer.startAudio.Rewind()
        g.audioPlayer.startAudio.Play()
        g.audioPlayer.ambientAudio.Play()
    }

}

func (g *Game) SetInGame() {
    g.state = IN_GAME
    g.player.alpha = 1.0
    g.StopRewinding()
}

func (g *Game) SetPlacing() {
    g.state = PLACING
    g.player.alpha = ghostAlpha
    g.StopRewinding()
}

func loadAudioVorbis(oggFile []byte, audioContext *audio.Context) *audio.Player {

    var err error
    sound, err := vorbis.DecodeWithoutResampling(bytes.NewReader(oggFile))
    if err != nil {
        return nil
    }

    p, err := audioContext.NewPlayer(sound)

    if err != nil {
        return nil
    }
    return p
}
func loadAudio(wavFile []byte, audioContext *audio.Context) *audio.Player {

    var err error
    sound, err := wav.DecodeWithoutResampling(bytes.NewReader(wavFile))
    if err != nil {
        return nil
    }

    p, err := audioContext.NewPlayer(sound)

    if err != nil {
        return nil
    }
    return p
}

func (g *Game) LoadAudio() {
    g.audioPlayer = &AudioPlayer{}

    g.audioPlayer.audioContext = audio.NewContext(sampleRate)
    g.audioPlayer.rewindAudio = loadAudio(rewindWav_src, g.audioPlayer.audioContext)
    g.audioPlayer.stopAudio = loadAudio(stopWav_src, g.audioPlayer.audioContext)
    g.audioPlayer.startAudio = loadAudio(startWav_src, g.audioPlayer.audioContext)
    g.audioPlayer.ambientAudio = loadAudioVorbis(ambientOgg_src, g.audioPlayer.audioContext)

    if g.audioPlayer.ambientAudio == nil {
        fmt.Printf("AUDIO NUL")
    }
}

func (g *Game) LoadImages() {
    s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))
	if err != nil {
		log.Fatal(err)
	}
	fontFaceSource = s

    img, _, err := image.Decode(bytes.NewReader(characterPng_src))
	if err != nil {
		log.Fatal(err)
	}

    characterImage = ebiten.NewImageFromImage(img)

    img, _, err = image.Decode(bytes.NewReader(tilesPng_src))
	if err != nil {
		log.Fatal(err)
	}

    tilesImage = ebiten.NewImageFromImage(img)
}

func main() {
    err := LoadShaders()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowTitle("Hello, World!")
    game := &Game{}
    game.LoadAudio()
    game.LoadImages()
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

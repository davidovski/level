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
	_ "github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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
    gravity = 0.16
    friction = 0.75
    airResistance = 0.98

    endCardDuration = 240

    exitTransitionWeight = 0.7
    ghostAlpha = 0.5
    hightlightBorder = 2
    audioFadeIn = 0.99
    musicVolume = 0.2

    sampleRate = 44100
    shadowOffset = 1

    killPlayerAfter = 80
    musicLoopLength = 230

    menuFadeInTime = 80
    buttonOffset = 4.0
    buttonOffsetPressed = 2.0
)

var (
	//go:embed assets/font.otf
	fontOtf_src []byte

	//go:embed shaders/none.kage
	noneShader_src []byte
	//go:embed shaders/vcr.kage
	vcrShader_src []byte
	//go:embed shaders/clouds.kage
	cloudShader_src []byte
	//go:embed shaders/bloom.kage
	bloomShader_src []byte
	//go:embed shaders/vortex.kage
	vortexShader_src []byte

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

	//go:embed assets/spring.ogg
	springOgg_src []byte

	//go:embed assets/jump1.ogg
	jump1Ogg_src []byte

	//go:embed assets/jump2.ogg
	jump2Ogg_src []byte

	//go:embed assets/land1.ogg
	land1Ogg_src []byte

	//go:embed assets/land2.ogg
	land2Ogg_src []byte

	//go:embed assets/vo/voice1.ogg
	voice1Ogg_src []byte
	//go:embed assets/vo/voice2.ogg
	voice2Ogg_src []byte
	//go:embed assets/vo/voice3.ogg
	voice3Ogg_src []byte
	//go:embed assets/vo/voice4.ogg
	voice4Ogg_src []byte
	//go:embed assets/vo/voice5.ogg
	voice5Ogg_src []byte
	//go:embed assets/vo/voice6.ogg
	voice6Ogg_src []byte
	//go:embed assets/vo/voice7.ogg
	voice7Ogg_src []byte
	//go:embed assets/vo/voice8.ogg
	voice8Ogg_src []byte
	//go:embed assets/vo/voice9.ogg
	voice9Ogg_src []byte
	//go:embed assets/vo/voice10.ogg
	voice10Ogg_src []byte
	//go:embed assets/vo/voice11.ogg
	voice11Ogg_src []byte
)

var (
    shaders map[string]*ebiten.Shader
    tilesImage *ebiten.Image
    characterImage *ebiten.Image
    fontFaceSource *text.GoTextFaceSource

    rewindSpeed = 2
    bx, by, bw, bh float32
    bo float32 = buttonOffset 
)

type State int

const (
	MENU State = iota
	IN_GAME
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
    delta int
    state int
}

type AudioPlayer struct {
    audioContext *audio.Context
    rewindAudio  *audio.Player
    stopAudio  *audio.Player
    startAudio  *audio.Player
    ambientAudio  *audio.Player
    springAudio  *audio.Player
    jumpAudio  []*audio.Player
    landAudio  []*audio.Player
    voiceAudio  []*audio.Player
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
            delta: object.delta,
            state: object.state,
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
        obj.delta = point.delta
        obj.state = point.state
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
    if g.playerAiIdx >= len(g.playerAi) + killPlayerAfter {
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
        obj.delta = 0
        obj.highlight = false
    }
    g.recording = g.recording[:0];
}

func (g *Game) Init() {
    g.surface = ebiten.NewImage(screenWidth, screenHeight)
    g.shaderName = "sky"
    g.state = MENU
    //StartGame(g)// TODO
    g.audioPlayer.ambientAudio.SetVolume(0)


	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Reverse")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}

}
func (g *Game) SetEditingMode() {
        g.SetPlacing()
        //g.playerAi = g.playerAi[:0]
        g.ResetAll()
        g.playerAiIdx = 0
    }

func (g *Game) Update() error {
    g.time += 1

    if g.state == MENU {
        if float32(g.time) > menuFadeInTime {
            cx, cy := ebiten.CursorPosition()
            onButton :=  float32(cx) > bx && float32(cy) > by && float32(cx) < bx + bw && float32(cy) < by + bh
            if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
                if onButton {
                    bo = buttonOffsetPressed
                } else {
                    bo = buttonOffset
                }
            }
            if inpututil.IsMouseButtonJustReleased(ebiten.MouseButton0) {
                bo = buttonOffset
                if onButton && g.animStart <= 0{
                    g.animStart = g.time + 60
                    g.audioPlayer.landAudio[0].Rewind()
                    g.audioPlayer.landAudio[0].Play()
                    StartGame(g)
                }
            }
        }
    }


    if g.state == MENU || g.state == IN_GAME || g.state == PLACING {
        if g.audioPlayer.ambientAudio.Position().Seconds() > musicLoopLength {
            g.audioPlayer.ambientAudio.Rewind()
        }

        if ! g.audioPlayer.ambientAudio.IsPlaying() {
            g.audioPlayer.ambientAudio.Play()
        }
        volume := g.audioPlayer.ambientAudio.Volume()
        volume = musicVolume-((musicVolume-volume)*audioFadeIn)
        g.audioPlayer.ambientAudio.SetVolume(volume)
    }

    if g.state == PAUSED && inpututil.IsKeyJustPressed(ebiten.KeyR) {
        g.TransitionState()
    }

    if g.state == IN_GAME {
        if inpututil.IsKeyJustPressed(ebiten.KeyE) {
            g.SetEditingMode()
        }

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
            obj.highlight = false
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
    if g.player != nil && g.player.y > screenHeight {
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
    } else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButton0) {
        g.PlaceObject(cx, cy)
    }
}

func (g *Game) PlaceObject(cx, cy int) {

        if len(g.toPlace) == 0 {
            return
        }

        placeable := g.toPlace[0]
        if placeable.HasCollision(*g.tilemap, g.objects, NONE) {
            return
        }

        placeable.delta = 0
        placeable.startx = float32(cx)
        placeable.starty = float32(cy)
        placeable.highlight = false
        placeable.alpha = 1.0

        g.objects = append(g.objects, placeable)

        g.toPlace = g.toPlace[1:len(g.toPlace)]
        placeable.PlayLand()
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

func PostProcess(screen *ebiten.Image, shaderName string, time int, game *Game) {
    w, h := screen.Bounds().Dx(), screen.Bounds().Dy()

    //out := ebiten.NewImage(w, h)
    //shop := &ebiten.DrawRectShaderOptions{}

    //shop.Uniforms = map[string]any{
    //    "Time":   float32(time) / 60,
    //    "Ex":   game.exit.x,
    //    "Ey":   game.exit.y,
    //}
    //shop.Images[0] = screen
    //shop.Images[1] = screen
    //shop.Images[2] = screen
    //shop.Images[3] = screen
    //out.DrawRectShader(w, h, shaders["vortex"], shop)

    //op := &ebiten.DrawImageOptions{}
    //screen.DrawImage(out, op)

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

func (g *Game) DrawVCRControls(surface *ebiten.Image) {
    textSize := 15.0

    var msg string
    if g.state == REVERSING {
        msg = fmt.Sprintf("<<")
    } else {
        msg = fmt.Sprintf("||")
    }
    textOp := &text.DrawOptions{}
    textOp.GeoM.Translate(15, 15)
    textOp.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
    text.Draw(surface, msg, &text.GoTextFace{
        Size:   textSize,
        Source: fontFaceSource,
    }, textOp)

}

func (g *Game) DrawTheEnd(surface *ebiten.Image, alpha float32) {
    textSize := 30.0

    msg := fmt.Sprintf("THE END")
    textOp := &text.DrawOptions{}
    textOp.GeoM.Translate((screenWidth - textSize*7 ) / 2, (screenHeight - textSize) / 2)
    textOp.ColorScale.ScaleWithColor(color.RGBA{55, 53, 53, 255})
    textOp.ColorScale.ScaleAlpha(alpha)
    text.Draw(surface, msg, &text.GoTextFace{
        Size:   textSize,
        Source: fontFaceSource,
    }, textOp)
}

func (g *Game) DrawTitle(surface *ebiten.Image, alpha float32) {
    textSize := 30.0
    tmp := ebiten.NewImage(screenWidth, screenHeight)

    msg := fmt.Sprintf("LEVÅÆ")
    textOp := &text.DrawOptions{}
    textOp.GeoM.Translate((screenWidth - textSize*5 ) / 2, (screenHeight - textSize) / 3)
    textOp.ColorScale.ScaleWithColor(color.RGBA{216, 211, 210, 255})
    text.Draw(tmp, msg, &text.GoTextFace{
        Size:   textSize,
        Source: fontFaceSource,
    }, textOp)
    ShadowDraw(surface, tmp, 0, 0, alpha)
}
func (g *Game) DrawStart(surface *ebiten.Image, alpha float32) {
    msg := fmt.Sprintf("play")
    tmp := ebiten.NewImage(screenWidth, screenHeight)
    var textSize float32 = 15.0

    var x, y float32 = (screenWidth - textSize*5 ) / 2 + bo/2, (screenHeight + textSize) / 3 + screenHeight / 3 + bo/2
    var padding float32 = 5.0

    c := color.RGBA{55, 53, 53, 255}
    bx, by, bw, bh =  x - padding, y - padding, (2*padding + float32(len(msg))*textSize) , (2.0*padding + textSize)
    vector.DrawFilledRect(tmp, bx, by, bw, bh, c, false)

    c = color.RGBA{79, 78, 78, 255}
    vector.DrawFilledRect(tmp, bx-bo, by-bo, bw, bh, c, false)

    textOp := &text.DrawOptions{}
    textOp.GeoM.Translate(float64(x-bo), float64(y-bo))
    textOp.ColorScale.ScaleWithColor(color.RGBA{216, 211, 210, 255})
    //textOp.ColorScale.ScaleAlpha(alpha)
    textOp.Filter = ebiten.FilterNearest
    text.Draw(tmp, msg, &text.GoTextFace{
        Size:   float64(textSize),
        Source: fontFaceSource,
    }, textOp)

    op := &ebiten.DrawImageOptions{}
    op.ColorScale.ScaleAlpha(alpha)
    surface.DrawImage(tmp, op)
}

func (g *Game) IsShowTheEnd() bool {
    return g.animStart > 0 && (g.state == END || g.state == PAUSED || g.state == REVERSING)

}

func (g *Game) Draw(screen *ebiten.Image) {

    DrawBackground(screen, g.time)

    if g.state == MENU {
        var scale float32 = 1.0
        delta := float32(g.time)
        if g.animStart > 0 {
            delta = float32((g.animStart+menuFadeInTime) - g.time)
            scale = 0.5
        }

        if g.tilemap != nil {
            polation := float32(math.Pow(float64(delta*scale)/menuFadeInTime, 3))*screenHeight
            g.tilemap.Draw(screen, 0, polation - 2)
        }

        alpha := float32(delta) * scale / (menuFadeInTime/2.0)
        if alpha > 1.0 {
            alpha = 1.0
        }
        g.DrawTitle(screen, alpha)

        alpha = float32(delta - (menuFadeInTime/2.0)) * scale / (menuFadeInTime/2.0)
        if alpha > 1.0 {
            alpha = 1.0
        }
        g.DrawStart(screen, alpha)


        if g.animStart > 0 && delta <= 0 {
            g.animStart = 0
            g.SetInGame()
        }
        return
    }

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
    op.GeoM.Translate(float64(g.offsetX), float64(g.offsetY))
    if g.IsShowTheEnd() {

        // draw THE END
        a :=  float64(endCardDuration - (g.time - g.animStart)) / float64(endCardDuration);
        a = 1 - float64(math.Pow(float64(a), 2))

        if g.state == PAUSED {
            a = 1.0
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

    if g.state == REVERSING || g.state == PAUSED {
        g.DrawVCRControls(screen)
    }


    PostProcess(screen, g.shaderName, g.time, g)

    //ebitenutil.DebugPrint(screen, fmt.Sprintf("tps: %.4f", ebiten.ActualFPS()))
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
    shaders["vortex"], err = ebiten.NewShader([]byte(vortexShader_src))
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
    g.audioPlayer.voiceAudio[0].Pause()
    rewindSpeed = 1 + int(len(g.recording) / 120)
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

func loadAudiosVorbis(oggFiles [][]byte, audioContext *audio.Context) []*audio.Player {
    var players = make([]*audio.Player, 0)
    for _, x := range oggFiles {
        var err error
        sound, err := vorbis.DecodeWithoutResampling(bytes.NewReader(x))
        if err != nil {
            return nil
        }

        p, err := audioContext.NewPlayer(sound)

        if err != nil {
            return nil
        }
        players = append(players, p)
    }
    return players
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
    g.audioPlayer.springAudio = loadAudioVorbis(springOgg_src, g.audioPlayer.audioContext)
    g.audioPlayer.jumpAudio = loadAudiosVorbis([][]byte{jump1Ogg_src, jump2Ogg_src}, g.audioPlayer.audioContext)
    g.audioPlayer.landAudio = loadAudiosVorbis([][]byte{land1Ogg_src, land2Ogg_src}, g.audioPlayer.audioContext)

    g.audioPlayer.voiceAudio = loadAudiosVorbis([][]byte{
        voice1Ogg_src,
        voice2Ogg_src,
        voice3Ogg_src,
        voice4Ogg_src,
        voice5Ogg_src,
        voice6Ogg_src,
        voice7Ogg_src,
        voice8Ogg_src,
        voice9Ogg_src,
        voice10Ogg_src,
        voice11Ogg_src,
    }, g.audioPlayer.audioContext)

}

func (g *Game) LoadImages() {
    s, err := text.NewGoTextFaceSource(bytes.NewReader(fontOtf_src))
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

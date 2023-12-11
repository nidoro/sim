package main

import (
    "github.com/nidoro/sim"
    "log"
    "fmt"
    "image/color"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "github.com/hajimehoshi/ebiten/v2/text"
    "github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
    "github.com/hajimehoshi/ebiten/v2/vector"
    
    "golang.org/x/image/font"
    "golang.org/x/image/font/opentype"
)

type Truck struct {
    sim.EntityBase
    TerminalId  string
}

type TruckSource struct {
    sim.EntitySourceBase
    TerminalId      string
}

func (source *TruckSource) Generate() sim.Entity {
    env := source.GetEnvironment()
    truck := &Truck{
        TerminalId: "Terminal",
    }
    
    env.AddEntity("Truck", truck)
    env.ForwardTo(truck, fmt.Sprintf("CLAS %s", source.TerminalId))
    
    return truck
}

type Drawable struct {
    Image *ebiten.Image
    X float64
    Y float64
    ScaleX float64
    ScaleY float64
}

func NewDrawable(w int, h int) *Drawable {
    return &Drawable{
        Image: ebiten.NewImage(w, h),
        ScaleX: 1,
        ScaleY: 1,
    }
}

type Global struct {
    Canvas *ebiten.Image
    CanvasWidth float64
    CanvasHeight float64
    FirstUpdate bool
    
    Camera *ebiten.Image
    CameraX float64
    CameraY float64
    CameraWidth float64
    CameraHeight float64
    CameraScale float64
    CameraOuterWidth float64
    CameraOuterHeight float64
    CameraAspectRatio float64
    
    CanvasDrawables []*Drawable
    
    IsDragging bool
    DragCursorStartX float64
    DragCursorStartY float64
    DragCameraStartX float64
    DragCameraStartY float64
    
    
    FontFace font.Face
    Env *sim.Environment
}

var g Global

func GetCursorPositionNormalized() (float64, float64) {
    x, y := ebiten.CursorPosition()
    nx := float64(x) / g.CameraOuterWidth
    ny := float64(y) / g.CameraOuterHeight
    return nx, ny
}

func GetCursorCanvasPosition() (float64, float64) {
    nx, ny := GetCursorPositionNormalized()
    canx := g.CameraX + nx * g.CameraOuterWidth / g.CameraScale
    cany := g.CameraY + ny * g.CameraOuterHeight / g.CameraScale
    return canx, cany
}

func (g *Global) Update() error {
    g.Env.Advance()
    
    winW, winH := ebiten.WindowSize()
    
    if g.FirstUpdate {
        ebiten.MaximizeWindow()
        winW, winH = ebiten.WindowSize()
        g.FirstUpdate = false
        
        g.CameraOuterWidth = float64(winW)
        g.CameraOuterHeight = float64(winH)
        g.CameraWidth = g.CanvasWidth
        g.CameraHeight = g.CanvasHeight
        g.Camera = ebiten.NewImage(int(g.CameraOuterWidth), int(g.CameraOuterHeight))
        g.CameraAspectRatio = g.CameraOuterHeight / g.CameraOuterWidth
    }
    
    if ebiten.IsFullscreen() {
        winW, winH = ebiten.ScreenSizeInFullscreen()
    }
    
    if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
        g.CameraX += 0.02*g.CameraOuterWidth
    }
    
    if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
        g.CameraX -= 0.02*g.CameraOuterWidth
    }
    
    if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
        g.CameraY -= 0.02*g.CameraOuterHeight
    }
    
    if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
        g.CameraY += 0.02*g.CameraOuterHeight
    }
    
    if ebiten.IsKeyPressed(ebiten.KeyControl) && ebiten.IsKeyPressed(ebiten.KeyEqual) {
        g.CameraX += 0.02*0.5*g.CameraWidth
        g.CameraY += 0.02*0.5*g.CameraHeight
        g.CameraWidth *= 0.98
    }
    
    if ebiten.IsKeyPressed(ebiten.KeyControl) && ebiten.IsKeyPressed(ebiten.KeyMinus) {
        g.CameraX -= 0.02*0.5*g.CameraWidth
        g.CameraY -= 0.02*0.5*g.CameraHeight
        g.CameraWidth *= 1.02
    }
    
    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        g.IsDragging = true
        g.DragCursorStartX, g.DragCursorStartY = GetCursorPositionNormalized()
        g.DragCameraStartX = g.CameraX
        g.DragCameraStartY = g.CameraY
    }
    
    if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
        g.IsDragging = false
    }
    
    if g.IsDragging {
        x, y := GetCursorPositionNormalized()
        dx := g.DragCursorStartX - x
        dy := g.DragCursorStartY - y
        g.CameraX = g.DragCameraStartX + dx*g.CameraWidth
        g.CameraY = g.DragCameraStartY + dy*g.CameraHeight
    }
    
    _, wheelDeltaY := ebiten.Wheel()
    if wheelDeltaY > 0 {
        nx, ny := GetCursorPositionNormalized()
        g.CameraX -= 0.02*nx*g.CameraWidth
        g.CameraY -= 0.02*ny*g.CameraHeight
        g.CameraWidth *= 1.02
    } else if wheelDeltaY < 0 {
        nx, ny := GetCursorPositionNormalized()
        g.CameraX += 0.02*nx*g.CameraWidth
        g.CameraY += 0.02*ny*g.CameraHeight
        g.CameraWidth *= 0.98
    }
    
    g.CameraWidth = min(g.CameraWidth, g.CanvasWidth)
    g.CameraHeight = min(g.CameraWidth * g.CameraAspectRatio, g.CanvasHeight)
    
    g.CameraX = min(g.CanvasWidth - g.CameraWidth, g.CameraX)
    g.CameraY = min(g.CanvasHeight - g.CameraHeight, g.CameraY)
    g.CameraX = max(0, g.CameraX)
    g.CameraY = max(0, g.CameraY)
    
    g.CameraScale = g.CameraOuterWidth / g.CameraWidth

    return nil
}

func (g *Global) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{230, 230, 230, 255})
    
    g.Canvas.Fill(color.RGBA{0, 230, 230, 255})
    vector.StrokeRect(g.Canvas, 10, 10, float32(g.CanvasWidth-20), float32(g.CanvasHeight-20), 20, color.RGBA{255, 0, 0, 255}, true)
    text.Draw(g.Canvas, fmt.Sprintf("%d", len(g.Env.GetProcessBase("UNTK Terminal").Queue)), g.FontFace, 0, 50, color.Black)
    vector.DrawFilledRect(g.Canvas, 100, 100, 100, 100, color.Black, true)
    vector.StrokeRect(g.Canvas, 100, 100, 100, 100, 10, color.RGBA{255, 0, 0, 1}, true)
    
    for _, dwb := range g.CanvasDrawables {
        op := &ebiten.DrawImageOptions{}
        op.GeoM.Scale(dwb.ScaleX, dwb.ScaleY)
        op.GeoM.Translate(dwb.X, dwb.Y)
        g.Canvas.DrawImage(dwb.Image, op)
    }
    
    /*
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Scale(g.CanvasScale, g.CanvasScale)
    op.GeoM.Translate(g.CanvasX, g.CanvasY)
    
    screen.DrawImage(g.Canvas, op)
    */
    
    g.Camera.Fill(color.RGBA{0, 0, 0, 255})
    
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(-g.CameraX, -g.CameraY)
    op.GeoM.Scale(g.CameraScale, g.CameraScale)
    g.Camera.DrawImage(g.Canvas, op)
    
    op = &ebiten.DrawImageOptions{}
    screen.DrawImage(g.Camera, op)
    
    ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s", sim.GetHumanTime(g.Env.Now)), 0, 0)
    ebitenutil.DebugPrintAt(screen, fmt.Sprintf("x: %.2f y: %.2f", g.CameraX, g.CameraY), 0, 24)
    ebitenutil.DebugPrintAt(screen, fmt.Sprintf("scale: %.2f", g.CameraScale), 0, 48)
    
    canx, cany := GetCursorCanvasPosition()
    
    ebitenutil.DebugPrintAt(screen, fmt.Sprintf("canx: %.2f, cany: %.2f", canx, cany), 0, 120)
}

func (g *Global) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return outsideWidth, outsideHeight
}

func main() {
    tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
    if err != nil {
        log.Fatal(err)
    }

    const dpi = 72
    g.FontFace, err = opentype.NewFace(tt, &opentype.FaceOptions{
        Size:    48,
        DPI:     dpi,
        Hinting: font.HintingFull,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    g.CanvasWidth = 1600
    g.CanvasHeight = 1600
    g.Canvas = ebiten.NewImage(int(g.CanvasWidth), int(g.CanvasHeight))
    g.FirstUpdate = true
    
    g.CameraOuterWidth = 800
    g.CameraOuterHeight = 9.0/16.0 * g.CameraOuterWidth
    g.CameraWidth = g.CanvasWidth
    g.CameraHeight = g.CanvasHeight
    g.CameraX = 0
    g.CameraY = 0
    g.CameraScale = 1
    g.Camera = ebiten.NewImage(int(g.CameraOuterWidth), int(g.CameraOuterHeight))
    
    dwb := NewDrawable(200, 100)
    dwb.X = g.CanvasWidth/2
    dwb.Y = g.CanvasHeight/2
    
    dwb.Image.Fill(color.RGBA{0, 0, 230, 255})
    text.Draw(dwb.Image, "UNTK Terminal", g.FontFace, 0, 48, color.Black)
    
    g.CanvasDrawables = append(g.CanvasDrawables, dwb)
    
    g.Env = sim.NewEnvironment()
    
    interval := sim.Days(30) / 10000
    
    g.Env.AddEntitySource(&TruckSource{
        EntitySourceBase: sim.EntitySourceBase{
            Id: "Terminal", 
            RNG: sim.NewRNGExponential(1/interval),
        },
        TerminalId: "Terminal",
    })
    
    tid := "Terminal"
    
    g.Env.AddResource(&sim.ResourceBase{Id: fmt.Sprintf("CLAS %s", tid), Amount: 1})
    g.Env.AddProcess(
        sim.ProcessBase{
            Id: fmt.Sprintf("CLAS %s", tid),
            Groups: []string{"CLAS", tid},
            Needs: map[string]float64{
                fmt.Sprintf("CLAS %s", tid): 1,
            },
            RNG: sim.NewRNGTriangular(sim.Minutes(1), sim.Minutes(5), sim.Minutes(3)),
            NextProcess: fmt.Sprintf("UNTK %s", tid),
        },
    )
    
    g.Env.AddResource(&sim.ResourceBase{Id: fmt.Sprintf("UNTK %s", tid), Amount: 1})
    g.Env.AddProcess(
        sim.ProcessBase{
            Id: fmt.Sprintf("UNTK %s", tid),
            Groups: []string{"UNTK", tid},
            Needs: map[string]float64{
                fmt.Sprintf("UNTK %s", tid): 1,
            },
            RNG: sim.NewRNGLogNormal(sim.Minutes(4.3), sim.Minutes(0.6)),
        },
    )
    
    g.Env.LogLevel = 0
    g.Env.EndDate = sim.Days(30)
    
    g.Env.Begin()
    
    ebiten.SetWindowTitle("Hello, World!")
    ebiten.SetTPS(30)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
    
    if err := ebiten.RunGame(&g); err != nil {
        log.Fatal(err)
    }
    
}














package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	x float64
}

func (g *Game) Update() error {
	// 每秒移动 100 像素
	g.x += 100.0 / ebiten.CurrentTPS()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 绘制一个红色矩形，矩形的位置每秒右移 100 像素
	screen.Fill(color.RGBA{0xff, 0x00, 0x00, 0xff})
	ebitenutil.DrawRect(screen, g.x, 100, 50, 50, color.RGBA{0xff, 0xff, 0xff, 0xff})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(320, 240)
	ebiten.SetWindowTitle("Animation Example")
	if err := ebiten.RunGame(&Game{}); err != nil {
		panic(err)
	}
}

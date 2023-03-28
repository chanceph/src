package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth  = 520
	screenHeight = 640
	gameWidth    = 320
	gameHeight   = 640
	infoWidth    = 200
	blockSize    = 32
	scoreX       = 340
	scoreY       = 32
	gameTipX     = 340
	gameTipY     = 320
)

var (
	colors = []color.RGBA{
		{255, 0, 0, 255},   // 红色
		{0, 255, 0, 255},   // 绿色
		{0, 0, 255, 255},   // 蓝色
		{255, 255, 0, 255}, // 黄色
		{255, 0, 255, 255}, // 紫色
		{0, 255, 255, 255}, // 青色
		{255, 165, 0, 255}, // 橙色
		{128, 128, 128, 255}, // 灰色
	}

	blocks = [][][]int{
		// I 形方块
		{
			{0, 0, 0, 0},
			{1, 1, 1, 1},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
		// J 形方块
		{
			{2, 0, 0},
			{2, 2, 2},
			{0, 0, 0},
		},
		// L 形方块
		{
			{0, 0, 3},
			{3, 3, 3},
			{0, 0, 0},
		},
		// O 形方块
		{
			{4, 4},
			{4, 4},
		},
		// S 形方块
		{
			{0, 5, 5},
			{5, 5, 0},
			{0, 0, 0},
		},
		// T 形方块
		{
			{0, 6, 0},
			{6, 6, 6},
			{0, 0, 0},
		},
		// Z 形方块
		{
			{7, 7, 0},
			{0, 7, 7},
			{0, 0, 0},
		},
	}

	game     [][]int // 当前游戏状态，0 表示空，1-7 表示各种颜色的方块
	curBlock [][]int // 当前方块状态，0 表示空，1-7 表示各种颜色的方块
	curColor int     // 当前方块的颜色
	curX     int     // 当前方块左上角的横坐标
	curY     int     // 当前方块左上角的纵坐标

	nextBlock [][]int // 当前方块状态，0 表示空，1-7 表示各种颜色的方块
	nextColor int     // 当前方块的颜色
	nextX     int     // 当前方块左上角的横坐标
	nextY     int     // 当前方块左上角的纵坐标

	totalScore   int
	frameCount   int
	fallInterval = 10
	blockImages  []*ebiten.Image
	fontGame     font.Face
	gameOver     = false
	paused       = false
)

type block struct {
	shape [4][4]int
	color color.Color
	x, y  int
}

func getRandomBlock() ([][]int, int) {
	rand.Seed(time.Now().UnixNano())
	randInt := rand.Intn(len(blocks))
	return blocks[randInt], randInt
}

func moveLeft() {
	if canMoveLeft() {
		curX--
	}
}

func moveRight() {
	if canMoveRight() {
		curX++
	}
}

func rotate() {
	rotatedBlock := make([][]int, len(curBlock[0]))
	for i := range rotatedBlock {
		rotatedBlock[i] = make([]int, len(curBlock))
	}
	for i := range curBlock {
		for j := range curBlock[i] {
			rotatedBlock[j][len(curBlock)-1-i] = curBlock[i][j]
		}
	}
	if canRotate(rotatedBlock) {
		curBlock = rotatedBlock
	}
}

func canMoveLeft() bool {
	for i := range curBlock {
		for j := range curBlock[i] {
			if curBlock[i][j] != 0 && (curX+j <= 0 || game[curY+i][curX+j-1] != 0) {
				return false
			}
		}
	}
	return true
}

func canMoveRight() bool {
	for i := range curBlock {
		for j := range curBlock[i] {
			if curBlock[i][j] != 0 && (curX+j >= gameWidth/blockSize-1 || game[curY+i][curX+j+1] != 0) {
				return false
			}
		}
	}
	return true
}

func canRotate(rotatedBlock [][]int) bool {
	for i := range rotatedBlock {
		for j := range rotatedBlock[i] {
			if rotatedBlock[i][j] != 0 {
				if curX+j < 0 || curX+j >= gameWidth/blockSize || curY+i >= gameHeight/blockSize || game[curY+i][curX+j] != 0 {
					return false
				}
			}
		}
	}
	return true
}

func canMoveDown() bool {
	for i := range curBlock {
		for j := range curBlock[i] {
			if curBlock[i][j] != 0 && (curY+i >= gameHeight/blockSize-1 || game[curY+i+1][curX+j] != 0) {
				return false
			}
		}
	}
	return true
}

func moveDown() {
	if canMoveDown() {
		curY++
	} else {
		//新方块出来就是0且无法下坠, 游戏结束
		if curY == 0 {
			gameOver = true
		}

		// 将当前方块放入游戏状态中
		for i := range curBlock {
			for j := range curBlock[i] {
				if curBlock[i][j] != 0 {
					game[curY+i][curX+j] = curColor + 1
				}
			}
		}

		// 随机生成下一个方块状态
		curBlock = nextBlock
		curColor = nextColor
		curX = (gameWidth/blockSize - len(curBlock[0])) / 2
		curY = 0

		nextBlock, nextColor = getRandomBlock()
	}
}

type Game struct {
}

func (g *Game) Update() error {

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		initGame()
		gameOver = false
		paused = false
	}

	if gameOver {
		// 如果游戏已经结束，则停止更新游戏画面
		// ebiten.SetMaxTPS(0)
		return nil
	}
	// 处理用户输入
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		paused = !paused
	}
	if paused {
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		moveLeft()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		moveRight()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		rotate()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		moveDown()
	}

	// 方块下落
	if frameCount%fallInterval == 0 {
		moveDown()
	}
	frameCount++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	// 绘制背景色
	ebitenutil.DrawRect(screen, 0, 0, gameWidth, gameHeight, color.White)

	// 计算分数
	score := 0
	for i := range game {
		fullLine := true
		for j := range game[i] {
			if game[i][j] == 0 {
				fullLine = false
				break
			}
		}
		if fullLine {
			copy(game[1:i+1], game[:i])
			game[0] = make([]int, gameWidth/blockSize)
			score += 10
		}
	}
	totalScore += score

	// 绘制游戏区域
	for i := range game {
		for j := range game[i] {
			if game[i][j] != 0 {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(j*blockSize), float64(i*blockSize))
				screen.DrawImage(blockImages[7], op)
			}
		}
	}

	// 绘制当前方块
	for i := range curBlock {
		for j := range curBlock[i] {
			if curBlock[i][j] != 0 {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64((curX+j)*blockSize), float64((curY+i)*blockSize))
				screen.DrawImage(blockImages[curColor], op)
			}
		}
	}

	// 绘制下一个方块
	for i := range nextBlock {
		for j := range nextBlock[i] {
			if nextBlock[i][j] != 0 {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64((nextX+j)*blockSize), float64((nextY+i)*blockSize))
				screen.DrawImage(blockImages[nextColor], op)
			}
		}
	}
	text.Draw(screen, fmt.Sprintf("Score: %d", totalScore), fontGame, scoreX, scoreY, color.White)

	if gameOver || paused {
		tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
		if err != nil {
			log.Fatal(err)
		}
		fontGame1, _ := opentype.NewFace(tt, &opentype.FaceOptions{
			Size:    32,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		str := ""
		if gameOver {
			str = "GAME OVER"
		}
		if paused {
			str = "PAUSED"
		}
		text.Draw(screen, str, fontGame1, gameTipX, gameTipY, color.White)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Tetris")

	if err := ebiten.RunGame(&Game{}); err != nil {
		panic(err)
	}
}

func init() {

	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	fontGame, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    12,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	initGame()
}

func initGame() {
	// 初始化游戏状态
	game = make([][]int, gameHeight/blockSize)
	for i := range game {
		game[i] = make([]int, gameWidth/blockSize)
	}
	// 随机生成当前方块状态
	curBlock, curColor = getRandomBlock()
	curX = (gameWidth/blockSize - len(curBlock[0])) / 2
	curY = 0

	// 随机生成下一个方块状态
	nextBlock, nextColor = getRandomBlock()
	nextX = (gameWidth / blockSize) + ((infoWidth)/blockSize-len(curBlock[0]))/2
	nextY = 2

	// 加载方块图片资源
	for i := 0; i < 8; i++ {
		img := ebiten.NewImage(blockSize, blockSize)
		img.Fill(colors[i])
		blockImages = append(blockImages, img)
	}
	ebiten.NewImage(screenHeight, screenWidth)
}

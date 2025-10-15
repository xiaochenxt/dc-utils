package captcha

import (
	"bytes"
	"errors"
	"fmt"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"os"
)

const (
	width             = 200                // 验证码宽度
	height            = 70                 // 验证码高度
	charCount         = 4                  // 字符数量
	fontSize          = 50                 // 基础字体大小
	mainLineThickness = 3                  // 主干扰线粗细
	dotCount          = 15                 // 噪点数量
	fontPath          = "./ttf/epilog.ttf" // 字体路径
)

var (
	customFont *opentype.Font // 加载的字体
)

func init() {
	loadFont()
}

// Captcha 封装验证码答案和图片数据
type Captcha struct {
	Answer string
	Image  []byte
}

// Generate 生成验证码
func Generate() (*Captcha, error) {
	if customFont == nil {
		return nil, errors.New("字体加载失败")
	}
	code := generateRandomCode()
	img := drawCaptcha(code)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return &Captcha{
		Answer: code,
		Image:  buf.Bytes(),
	}, nil
}

func generateRandomCode() string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, charCount)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func loadFont() {
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		fmt.Printf("读取字体文件失败: %v\n", err)
		return
	}
	fontObj, err := opentype.Parse(fontData)
	if err != nil {
		fmt.Printf("解析字体失败: %v\n", err)
		return
	}
	customFont = fontObj
}

func drawCaptcha(code string) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	drawCurvedDisturbLine(img, code)
	drawChars(img, code)
	drawNoiseDots(img)

	return img
}

func drawCurvedDisturbLine(img *image.RGBA, code string) {
	charPositions := getCharCenterPositions(code)
	controlPoints := buildControlPoints(charPositions)

	r := uint8(rand.Intn(150) + 50)
	g := uint8(rand.Intn(150) + 50)
	b := uint8(rand.Intn(150) + 50)
	lineColor := color.RGBA{r, g, b, 200}

	drawBezierCurve(img, controlPoints, lineColor, mainLineThickness)
}

func getCharCenterPositions(code string) []image.Point {
	avgCharWidth := width / (charCount + 1)
	positions := make([]image.Point, charCount)
	for i := range positions {
		x := avgCharWidth*(i+1) - avgCharWidth/2
		y := height/2 + rand.Intn(10) - 5
		positions[i] = image.Point{x, y}
	}
	return positions
}

func buildControlPoints(charPoints []image.Point) []image.Point {
	if len(charPoints) == 0 {
		return nil
	}
	controlPoints := make([]image.Point, 0)
	startX := rand.Intn(20)
	startY := charPoints[0].Y
	controlPoints = append(controlPoints, image.Point{startX, startY})

	for _, p := range charPoints {
		controlPoints = append(controlPoints, p)
	}

	endX := width - rand.Intn(20)
	endY := charPoints[len(charPoints)-1].Y
	controlPoints = append(controlPoints, image.Point{endX, endY})
	return controlPoints
}

// 线性插值函数，修复math.Lerp的问题
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func drawBezierCurve(img *image.RGBA, points []image.Point, clr color.RGBA, thickness int) {
	if len(points) < 2 {
		return
	}

	// 使用贝塞尔曲线算法绘制平滑曲线
	for i := 0; i < len(points)-1; i++ {
		p0 := points[i]
		p1 := points[i+1]

		// 绘制两点之间的线段
		for t := 0.0; t <= 1.0; t += 0.01 {
			x := int(lerp(float64(p0.X), float64(p1.X), t))
			y := int(lerp(float64(p0.Y), float64(p1.Y), t))

			// 绘制粗线
			for dy := -thickness; dy <= thickness; dy++ {
				for dx := -thickness; dx <= thickness; dx++ {
					if dx*dx+dy*dy <= thickness*thickness {
						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < width && ny >= 0 && ny < height {
							img.Set(nx, ny, clr)
						}
					}
				}
			}
		}
	}
}

func drawChars(img *image.RGBA, code string) {
	if customFont == nil {
		return
	}
	avgCharWidth := width / (charCount + 1)
	for i, c := range code {
		size := float64(fontSize - 4 + rand.Intn(10))
		face, err := opentype.NewFace(customFont, &opentype.FaceOptions{
			Size:    size,
			DPI:     72,
			Hinting: xfont.HintingFull,
		})
		if err != nil {
			continue
		}

		r := uint8(rand.Intn(200) + 30)
		g := uint8(rand.Intn(200) + 30)
		b := uint8(rand.Intn(200) + 30)
		charColor := color.RGBA{r, g, b, 255}

		x := avgCharWidth*(i+1) - avgCharWidth/2 + rand.Intn(10) - 5
		y := height/2 + 10 + rand.Intn(10) - 5
		angle := float64(rand.Intn(20)-10) * math.Pi / 180

		drawRotatedChar(img, string(c), face, charColor, x, y, angle)
	}
}

func drawRotatedChar(img *image.RGBA, char string, face xfont.Face, clr color.RGBA, x, y int, angle float64) {
	charImg := image.NewRGBA(image.Rect(0, 0, width, height))
	drawer := &xfont.Drawer{
		Dst:  charImg,
		Src:  image.NewUniform(clr),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	drawer.DrawString(char)

	rotateAndMerge(img, charImg, angle)
}

func rotateAndMerge(dst, src *image.RGBA, angle float64) {
	sinTheta := math.Sin(angle)
	cosTheta := math.Cos(angle)
	bounds := dst.Bounds()
	cx, cy := bounds.Max.X/2, bounds.Max.Y/2

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			xRel := x - cx
			yRel := y - cy
			xRot := int(float64(xRel)*cosTheta-float64(yRel)*sinTheta) + cx
			yRot := int(float64(xRel)*sinTheta+float64(yRel)*cosTheta) + cy
			if xRot >= 0 && xRot < width && yRot >= 0 && yRot < height {
				srcColor := src.At(xRot, yRot)
				r, g, b, a := srcColor.RGBA()
				if a > 0 {
					dstColor := dst.At(x, y)
					dr, dg, db, _ := dstColor.RGBA()
					alpha := float64(a) / 65535.0
					rNew := uint8(float64(r>>8)*alpha + float64(dr>>8)*(1-alpha))
					gNew := uint8(float64(g>>8)*alpha + float64(dg>>8)*(1-alpha))
					bNew := uint8(float64(b>>8)*alpha + float64(db>>8)*(1-alpha))
					dst.Set(x, y, color.RGBA{rNew, gNew, bNew, 255})
				}
			}
		}
	}
}

func drawNoiseDots(img *image.RGBA) {
	for i := 0; i < dotCount; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)

		r := uint8(rand.Intn(200) + 30)
		g := uint8(rand.Intn(200) + 30)
		b := uint8(rand.Intn(200) + 30)

		img.Set(x, y, color.RGBA{r, g, b, 255})
	}
}

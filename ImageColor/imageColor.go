package ImageColor

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"math"
	"math/rand"
	"time"
)

type IM struct {
	Im image.Image
}

func NewIm(reader io.Reader) (*IM, error) {
	//reader, err := os.Open("test.jpg")
	//if err != nil {
	//	log.Fatal(err)
	//}
	im, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return &IM{
		Im: im,
	}, nil
}

func (im *IM) ExtractColor() []float64 {
	// sample pixels
	bounds := im.Im.Bounds()
	stepX := (bounds.Max.X - bounds.Min.X)/16
	stepY := (bounds.Max.Y - bounds.Min.Y)/16
	var histogram = make([]float64, 16*16*3)
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y+=stepY {
		for x := bounds.Min.X; x < bounds.Max.X; x+=stepX {
			if idx >= 16*16*3 {
				break
			}
			r, g, b, _ := im.Im.At(x, y).RGBA()
			histogram[idx] = float64(r>>12)
			histogram[idx+1] = float64(g>>12)
			histogram[idx+2] = float64(b>>12)
			idx += 3
		}
	}
	return Norm(histogram)
}

func Norm(vec []float64) []float64 {
	var lenVec float64
	for i := 0; i < len(vec); i++ {
		lenVec += vec[i] * vec[i]
	}
	lenVec = math.Sqrt(lenVec)
	for i := 0; i < len(vec); i++ {
		vec[i] /= lenVec
	}
	return vec
}

func kMeans(points [][]float64, dim int, k int, maxIter int)  {
	// random select cluster center
	var centers = make([][]float64, k)
	rand.Seed(time.Now().UnixNano())
	idx := -1
	for i := range centers {
		add := rand.Intn(len(points) - idx - 2) + 1
		idx += add
		centers[i] = points[idx]
	}
	// K-means iteration
	var label = make([]int, len(points))
	var count = make([]int, len(centers))
	for i := 0; i < maxIter; i++ {
		// 1.assign points
		for j := 0; j < len(points); j++ {
			minDis := math.MaxFloat64
			minDisIdx := -1
			for k := 0; k < len(centers); k++ {
				dis := getDis(points[j], centers[k])
				if dis < minDis {
					minDis = dis
					minDisIdx = k
				}
			}
			label[j] = minDisIdx
		}
		// 2.re-calculate center
		count = make([]int, len(centers))
		for k := 0; k < len(centers); k++ {
			centers[k] = make([]float64, dim)
		}
		for j := 0; j < len(points); j++ {
			for d := 0; d < dim; d++{
				centers[label[j]][d] += points[j][d]
			}
			count[label[j]] += 1
		}
		for k := 0; k < len(centers); k++ {
			for d := 0; d < dim; d++{
				centers[k][d] /= float64(count[k])
			}
		}
	}
	fmt.Println(centers)
}

func getDis(pst1 []float64, pst2 []float64) (dst float64) {
	for i := 0; i < len(pst1); i++ {
		dst += math.Pow(pst1[i] - pst2[i], 2)
	}
	dst = math.Sqrt(dst)
	return
}


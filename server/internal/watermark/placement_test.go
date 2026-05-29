package watermark

import (
	"image"
	"reflect"
	"testing"
)

func TestCenterPositions(t *testing.T) {
	tests := []struct {
		name string
		base image.Point
		mark image.Point
		want []image.Point
	}{
		{"sisa genap", image.Pt(100, 80), image.Pt(20, 20), []image.Point{image.Pt(40, 30)}},
		{"sisa ganjil membulat ke bawah", image.Pt(101, 101), image.Pt(20, 20), []image.Point{image.Pt(40, 40)}},
		{"mark lebih besar dari base jadi negatif", image.Pt(10, 10), image.Pt(30, 30), []image.Point{image.Pt(-10, -10)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Center{}.positions(tt.base, tt.mark)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("positions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatternPositions(t *testing.T) {
	base := image.Pt(50, 50)
	mark := image.Pt(10, 10)
	want := []image.Point{
		image.Pt(0, 0), image.Pt(26, 0),
		image.Pt(0, 26), image.Pt(26, 26),
	}

	t.Run("grid tegak deterministik", func(t *testing.T) {
		got := Pattern{Angle: 0}.positions(base, mark)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("positions() = %v, want %v", got, want)
		}
	})

	t.Run("angle tidak memengaruhi layout", func(t *testing.T) {
		for _, angle := range []int{0, 45, 90} {
			got := Pattern{Angle: angle}.positions(base, mark)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("angle %d: positions() = %v, want %v", angle, got, want)
			}
		}
	})
}

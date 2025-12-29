package internal

type Frame struct {
	Left  float32
	Right float32
}

func (f *Frame) Clamp() {
	f.Left = max(-1.0, min(f.Left, 1.0))
	f.Right = max(-1.0, min(f.Right, 1.0))
}

func (f *Frame) Add(other Frame) {
	f.Left += other.Left
	f.Right += other.Right
}

func (f *Frame) Scale(factor float32) {
	f.Left *= factor
	f.Right *= factor
}

type FrameList []Frame

func (l FrameList) Add(other FrameList) {
	for i := range l {
		l[i].Add(other[i])
	}
}

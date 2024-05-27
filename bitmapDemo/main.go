package main

type Bitmap struct {
	bits []byte
	vmax uint
}

func NewBitmap(maxVal ...uint) *Bitmap {
	var max uint = 8192
	if len(maxVal) > 0 && maxVal[0] > 0 {
		max = maxVal[0]
	}
	bm := &Bitmap{
		vmax: max,
		bits: make([]byte, (max + 7) / 8),
	}
	return bm
}

func (bm *Bitmap) Set(num uint) {
	// 扩容
	if num > bm.vmax {
		bm.vmax += 1024
		if num > bm.vmax {
			bm.vmax = num
		}
		dd := int(num + 7) / 8 - len(bm.bits)
		if dd > 0 {
			tmp := make([]byte, dd)
			bm.bits = append(bm.bits, tmp...)
		}
	}
	bm.bits[num/8] |= 1 << (num%8)
}

func (bm *Bitmap) Remove(num uint) {
	if num > bm.vmax { return }
	bm.bits[num/8] &^= 1 << (num%8)
}

func (bm *Bitmap) Check(num uint) bool {
	if num > bm.vmax { return false }
	return bm.bits[num/8] & (1 << (num % 8)) != 0
}

func main() {

}
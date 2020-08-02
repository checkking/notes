package main

import "fmt"

type Bitmap struct {
    bulkets    []uint64
}

func NewBitmap() *Bitmap {
    return &Bitmap{}
}

func (bm *Bitmap) Has(num int) bool {
    bulket, bit := num / 64, uint(num % 64)
    return bulket < len(bm.bulkets) && bm.bulkets[bulket] & (1 << bit) != 0
}

func (bm *Bitmap) Add(num int) {
    bulket, bit := num / 64, uint(num % 64)
    if bulket >= len(bm.bulkets) {
        oldLen := len(bm.bulkets)
        newBulkes := make([]uint64, bulket + 1 + 10)
        copy(newBulkes[0:oldLen], bm.bulkets)
        bm.bulkets = newBulkes
    }
    bm.bulkets[bulket] |= (1 << bit)
}

func main() {
    bm := NewBitmap()
    bm.Add(1)
    bm.Add(3)
    bm.Add(1024)
    fmt.Printf("has 1:%v\n", bm.Has(1))
    fmt.Printf("has 3:%v\n", bm.Has(3))
    fmt.Printf("has 4:%v\n", bm.Has(4))
    fmt.Printf("has 1024:%v\n", bm.Has(1024))
}


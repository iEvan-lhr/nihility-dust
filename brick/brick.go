package brick

import "github.com/iEvan-lhr/nihility-dust/dust"

/*
一砖一瓦，构建地球
The earth can be built brick by brick
*/

// Brick
//不仅可靠更便于移植
//Not only reliable, but also easy to transplant
type Brick struct {
	//构建自灰尘
	Material []*dust.Dust
}

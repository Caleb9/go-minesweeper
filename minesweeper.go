package main

import (
	"fmt"
	. "github.com/caleb9/go-minesweeper/internal"
)

func main() {
	fmt.Println(Help)
	fmt.Println()
	fmt.Println(`
Secret Service reports that there are 6 mines on that meadow... but where?
Uncover all non-mine fields before someone steps on a wrong one. Beware though!
Minesweeper's first mistake is also their last...`)
	fmt.Println()

	rows, cols := 6, 6
	minesCount := 6
	minefield, _ := NewGrid(rows, cols, NewMines(rows, cols, minesCount))
	Game(minefield)
}

package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

type Grid interface {
	Rows() int
	Cols() int
	Mines() int
	TryDefuse(row, col int) (liveAnotherDay bool, err error)
	Flag(row, col int) error
	Reveal(alive bool) string
}

type field struct {
	IsMine        bool
	IsDefused     bool
	AdjacentMines int
	IsFlagged     bool
}

type grid [][]field

func (g grid) Rows() int {
	if g == nil {
		return 0
	}
	return len(g)
}

func (g grid) Cols() int {
	if g == nil || len(g) < 1 {
		return 0
	}
	return len(g[0])
}

func (g grid) Mines() int {
	mines := 0
	for _, rowFields := range g {
		for _, field := range rowFields {
			if field.IsMine {
				mines++
			}
		}
	}
	return mines
}

func (g grid) TryDefuse(row, col int) (liveAnotherDay bool, err error) {
	if row < 0 || g.Rows() <= row || col < 0 || g.Cols() <= col {
		return true, errors.New("invalid row or column")
	}
	if g[row][col].IsDefused {
		return true, errors.New("already defused")
	}
	g[row][col].IsDefused = true
	if g[row][col].IsMine {
		return false, nil
	}
	for _, gridRow := range g[max(row-1, 0):min(row+2, len(g))] {
		for _, field := range gridRow[max(col-1, 0):min(col+2, len(gridRow))] {
			if field.IsMine {
				g[row][col].AdjacentMines++
			}
		}
	}
	return true, nil
}

func (g grid) Flag(row, col int) error {
	if row < 0 || g.Rows() <= row || col < 0 || g.Cols() <= col {
		return errors.New("invalid row or column")
	}
	g[row][col].IsFlagged = !g[row][col].IsFlagged
	return nil
}

func (g grid) Reveal(alive bool) string {
	for row, rowFields := range g {
		for col := range rowFields {
			if g[row][col].IsMine {
				g[row][col].IsDefused = true
				g[row][col].IsFlagged = alive
			}
		}
	}
	return g.String()
}

func (g grid) String() string {
	var buffer strings.Builder
	buffer.WriteString(" ")
	for col := range g[0] {
		buffer.WriteString(fmt.Sprintf(" %c", rowColLabelRune(col)))
	}
	buffer.WriteString("\n")
	for row, rowFields := range g {
		buffer.WriteString(fmt.Sprintf("%c ", rowColLabelRune(row)))
		for _, field := range rowFields {
			if field.IsFlagged {
				buffer.WriteString("🚩")
			} else if !field.IsDefused {
				buffer.WriteString("🍀")
			} else if field.IsMine {
				buffer.WriteString("💣")
			} else {
				buffer.WriteString(fmt.Sprintf("%c", adjacentMinesRune(field.AdjacentMines)))
			}
		}
		buffer.WriteRune('\n')
	}
	return buffer.String()
}

func rowColLabelRune(num int) rune {
	return rune(int('\u2488') + num)
}

func adjacentMinesRune(num int) rune {
	return rune(int('\uff10') + num)
}

type Coordinate struct {
	row, col int
}

func NewGrid(rows, cols int, mines []Coordinate) (Grid, error) {
	const RowMin, RowMax, ColMin, ColMax = 3, 9, 3, 9
	if rows < RowMin || RowMax < rows || cols < ColMin || ColMax < cols {
		return nil, errors.New("invalid grid size")
	}
	g := make(grid, rows)
	for i := 0; i < rows; i++ {
		g[i] = make([]field, cols)
	}
	for _, mine := range mines {
		row, col := mine.row, mine.col
		if row < 0 || g.Rows() <= row || col < 0 || g.Cols() <= col {
			return nil, errors.New("wrong mine")
		}
		g[row][col].IsMine = true
	}
	return g, nil
}

func NewMines(rows, cols, count int) []Coordinate {
	mines := make([]Coordinate, rows*cols)
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			mines[row*cols+col] = Coordinate{row, col}
		}
	}
	rand.Shuffle(len(mines), func(i, j int) {
		mines[i], mines[j] = mines[j], mines[i]
	})
	return mines[0:count]
}

const Help = `
In each step, type ROW and COLUMN, confirm with [ENTER]

To flag a mine, add 'f' at the end

Examples:
22  - defuse field in row 2 and column 2
13f - flag field in row 1 and column 3 as mine`

func readAndParseInput(inputScanner *bufio.Scanner) (row, col int, flag bool, err error) {
	fmt.Print("❓ ")
	if !inputScanner.Scan() {
		os.Exit(0)
	}
	err = inputScanner.Err()
	if err != nil {
		return 0, 0, false, err
	}
	invalidInputErr := errors.New("invalid input")
	input := inputScanner.Text()
	if len(input) < 2 || 3 < len(input) {
		return 0, 0, false, invalidInputErr
	}
	row, rowErr := strconv.Atoi(input[0:1])
	col, colErr := strconv.Atoi(input[1:2])
	if rowErr != nil || colErr != nil {
		return 0, 0, false, invalidInputErr
	}
	if len(input) == 3 {
		if input[2] != 'f' {
			return 0, 0, false, invalidInputErr
		}
		flag = true
	}
	/* Subtract 1 because rows and columns are labeled with 1-indexed sequence */
	return row - 1, col - 1, flag, nil
}

func Game(minefield Grid) {
	alive, defusedFields := true, 0
	noMineFields := (minefield.Rows() * minefield.Cols()) - minefield.Mines()
	inputScanner := bufio.NewScanner(os.Stdin)
	for alive && defusedFields < noMineFields {
		fmt.Println()
		fmt.Println(minefield)
		row, col, flag, err := readAndParseInput(inputScanner)
		if err != nil {
			fmt.Println(err)
			fmt.Println(Help)
			continue
		}
		if flag {
			err = minefield.Flag(row, col)
			if err != nil {
				fmt.Println(err)
				fmt.Println(Help)
			}
			continue
		}
		still_alive, err := minefield.TryDefuse(row, col)
		if err != nil {
			fmt.Println(err)
			continue
		}
		alive = still_alive
		defusedFields++
	}
	fmt.Printf("\n%v\n", minefield.Reveal(alive))
	if alive {
		fmt.Println("\nYOU WIN! 🥵")
	} else {
		fmt.Println("\nYOU DIE! 🪦")
	}
}

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

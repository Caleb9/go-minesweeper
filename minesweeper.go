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
	Reveal(coord Coordinate) (isStillAlive bool, err error)
	Flag(coord Coordinate) error
	RevealAll(isVictory bool) string
}

type field struct {
	IsMine        bool
	IsRevealed    bool
	AdjacentMines int
	IsFlagged     bool
}

type grid [][]field

func (g grid) Rows() int {
	return len(g)
}

func (g grid) Cols() int {
	if len(g) < 1 {
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

func (g grid) Reveal(coord Coordinate) (isStillAlive bool, err error) {
	row, col := coord.row, coord.col
	if row < 0 || g.Rows() <= row || col < 0 || g.Cols() <= col {
		return true, errors.New("invalid row or column")
	}
	if g[row][col].IsRevealed {
		return true, errors.New("already defused")
	}
	g[row][col].IsRevealed = true
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

func (g grid) Flag(coord Coordinate) error {
	row, col := coord.row, coord.col
	if row < 0 || g.Rows() <= row || col < 0 || g.Cols() <= col {
		return errors.New("invalid row or column")
	}
	g[row][col].IsFlagged = !g[row][col].IsFlagged
	return nil
}

func (g grid) RevealAll(isVictory bool) string {
	for row, rowFields := range g {
		for col := range rowFields {
			if g[row][col].IsMine {
				g[row][col].IsRevealed = true
				g[row][col].IsFlagged = isVictory
			}
		}
	}
	return g.String()
}

func (g grid) String() string {
	var buffer strings.Builder
	buffer.WriteString(" ")
	for col := range g[0] {
		glyph, _ := rowColLabelGlyph(col)
		buffer.WriteString(fmt.Sprintf(" %c", glyph))
	}
	buffer.WriteString("\n")
	for row, rowFields := range g {
		glyph, _ := rowColLabelGlyph(row)
		buffer.WriteString(fmt.Sprintf("%c ", glyph))
		for _, field := range rowFields {
			if field.IsFlagged {
				buffer.WriteRune('🚩')
			} else if !field.IsRevealed {
				buffer.WriteRune('🟩')
			} else if field.IsMine {
				buffer.WriteRune('💣')
			} else {
				glyph, _ := adjacentMinesGlyph(field.AdjacentMines)
				buffer.WriteRune(glyph)
			}
		}
		buffer.WriteRune('\n')
	}
	return buffer.String()
}

const RowMax, ColMax = 9, 9

func rowColLabelGlyph(num int) (rune, error) {
	if num < 0 || RowMax <= num || ColMax <= num {
		return 0, errors.New("invalid num")
	}
	const One = int('\u2488')
	return rune(One + num), nil
}

func adjacentMinesGlyph(num int) (rune, error) {
	if num < 0 || RowMax <= num || ColMax <= num {
		return 0, errors.New("invalid num")
	}
	const Zero = int('\uff10')
	return rune(Zero + num), nil
}

type Coordinate struct {
	row, col int
}

func NewGrid(rows, cols int, mines []Coordinate) (Grid, error) {
	if mines == nil {
		return nil, errors.New("invalid mines")
	}
	const RowMin, ColMin = 3, 3
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
			return nil, errors.New("invalid mine")
		}
		g[row][col].IsMine = true
	}
	return g, nil
}

func NewMines(rows, cols, count int) []Coordinate {
	mines := make([]Coordinate, rows*cols)
	for row := range rows {
		for col := range cols {
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

func readAndParseInput(inputScanner *bufio.Scanner) (coord Coordinate, flag bool, err error) {
	fmt.Print("❓ ")
	if !inputScanner.Scan() {
		os.Exit(0)
	}
	err = inputScanner.Err()
	errCoord, invalidInputErr := Coordinate{-1, -1}, errors.New("invalid input")
	if err != nil {
		return errCoord, false, err
	}
	input := inputScanner.Text()
	if len(input) < 2 || 3 < len(input) {
		return errCoord, false, invalidInputErr
	}
	row, rowErr := strconv.Atoi(input[0:1])
	col, colErr := strconv.Atoi(input[1:2])
	if rowErr != nil || colErr != nil {
		return errCoord, false, invalidInputErr
	}
	if len(input) == 3 {
		if input[2] != 'f' {
			return errCoord, false, invalidInputErr
		}
		flag = true
	}
	/* Subtract 1 because rows and columns are labeled with 1-indexed sequence */
	return Coordinate{row - 1, col - 1}, flag, nil
}

func Game(minefield Grid) {
	isAlive, revealedFields := true, 0
	noMineFields := minefield.Rows()*minefield.Cols() - minefield.Mines()
	inputScanner := bufio.NewScanner(os.Stdin)
	for isAlive && revealedFields < noMineFields {
		fmt.Println()
		fmt.Println(minefield)
		coord, flag, err := readAndParseInput(inputScanner)
		if err != nil {
			fmt.Println(err)
			fmt.Println(Help)
			continue
		}
		if flag {
			err = minefield.Flag(coord)
			if err != nil {
				fmt.Println(err)
				fmt.Println(Help)
			}
			continue
		}
		isStillAlive, err := minefield.Reveal(coord)
		if err != nil {
			fmt.Println(err)
			continue
		}
		isAlive = isStillAlive
		revealedFields++
	}
	fmt.Printf("\n%v\n", minefield.RevealAll(isAlive))
	if isAlive {
		fmt.Println("\n🥵 YOU WIN!")
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

package internal

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
	rows() int
	cols() int
	mines() int
	reveal(coord Coordinate) (isStillAlive bool, err error)
	flag(coord Coordinate) error
	revealAll(isVictory bool) string
}

type field struct {
	isMine        bool
	isRevealed    bool
	adjacentMines int
	isFlagged     bool
}

type grid struct {
	fields [][]field
}

func (g *grid) rows() int {
	return len(g.fields)
}

func (g *grid) cols() int {
	if len(g.fields) < 1 {
		return 0
	}
	return len(g.fields[0])
}

func (g *grid) mines() int {
	mines := 0
	for _, rowFields := range g.fields {
		for _, field := range rowFields {
			if field.isMine {
				mines++
			}
		}
	}
	return mines
}

// reveal returns true if minesweeper is still alive, or an error
func (g *grid) reveal(coord Coordinate) (bool, error) {
	row, col := coord.row, coord.col
	if row < 0 || g.rows() <= row || col < 0 || g.cols() <= col {
		return false, errors.New("invalid row or column")
	}
	if g.fields[row][col].isRevealed {
		return false, errors.New("already defused")
	}
	g.fields[row][col].isRevealed = true
	if g.fields[row][col].isMine {
		return false, nil
	}
	for _, gridRow := range g.fields[max(row-1, 0):min(row+2, len(g.fields))] {
		for _, field := range gridRow[max(col-1, 0):min(col+2, len(gridRow))] {
			if field.isMine {
				g.fields[row][col].adjacentMines++
			}
		}
	}
	return true, nil
}

func (g *grid) flag(coord Coordinate) error {
	row, col := coord.row, coord.col
	if row < 0 || g.rows() <= row || col < 0 || g.cols() <= col {
		return errors.New("invalid row or column")
	}
	g.fields[row][col].isFlagged = !g.fields[row][col].isFlagged
	return nil
}

func (g *grid) revealAll(isVictory bool) string {
	for _, rowFields := range g.fields {
		for _, field := range rowFields {
			if field.isMine {
				field.isRevealed = true
				field.isFlagged = isVictory
			}
		}
	}
	return g.String()
}

func (g *grid) String() string {
	var buffer strings.Builder
	buffer.WriteString(" ")
	for col := range g.fields[0] {
		glyph, _ := rowColLabelGlyph(col)
		buffer.WriteString(fmt.Sprintf(" %c", glyph))
	}
	buffer.WriteString("\n")
	for row, rowFields := range g.fields {
		glyph, _ := rowColLabelGlyph(row)
		buffer.WriteString(fmt.Sprintf("%c ", glyph))
		for _, field := range rowFields {
			switch {
			case field.isFlagged:
				buffer.WriteRune('🚩')
			case !field.isRevealed:
				buffer.WriteRune('🟩')
			case field.isMine:
				buffer.WriteRune('💣')
			default:
				glyph, _ := adjacentMinesGlyph(field.adjacentMines)
				buffer.WriteRune(glyph)
			}
		}
		buffer.WriteRune('\n')
	}
	return buffer.String()
}

const (
	colMax = 9
	rowMax = 9
)

func rowColLabelGlyph(num int) (rune, error) {
	if num < 0 || rowMax <= num || colMax <= num {
		return 0, errors.New("invalid num")
	}
	const One = int('\u2488')
	return rune(One + num), nil
}

func adjacentMinesGlyph(num int) (rune, error) {
	if num < 0 || rowMax <= num || colMax <= num {
		return 0, errors.New("invalid num")
	}
	const Zero = int('\uff10')
	return rune(Zero + num), nil
}

type Coordinate struct {
	row, col int
}

func NewGrid(rows, cols int, mines []Coordinate) (Grid, error) {
	const RowMin, ColMin = 3, 3
	if rows < RowMin || rowMax < rows || cols < ColMin || colMax < cols {
		return nil, errors.New("invalid grid size")
	}
	g := grid{make([][]field, rows)}
	for i := 0; i < rows; i++ {
		g.fields[i] = make([]field, cols)
	}
	for _, mine := range mines {
		row, col := mine.row, mine.col
		if row < 0 || g.rows() <= row || col < 0 || g.cols() <= col {
			return nil, errors.New("invalid mine")
		}
		g.fields[row][col].isMine = true
	}
	return Grid(&g), nil
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

// readAndParseInput returns user input field coordinate, whether it's a flag, or an error
func readAndParseInput(inputScanner *bufio.Scanner) (Coordinate, bool, error) {
	fmt.Print("❓ ")
	if !inputScanner.Scan() {
		os.Exit(0)
	}
	err := inputScanner.Err()
	zeroCoord, invalidInputErr := Coordinate{}, errors.New("invalid input")
	if err != nil {
		return zeroCoord, false, err
	}
	input := inputScanner.Text()
	if len(input) < 2 || 3 < len(input) {
		return zeroCoord, false, invalidInputErr
	}
	row, rowErr := strconv.Atoi(input[0:1])
	col, colErr := strconv.Atoi(input[1:2])
	if rowErr != nil || colErr != nil {
		return zeroCoord, false, invalidInputErr
	}
	var flag bool
	if len(input) == 3 {
		if input[2] != 'f' {
			return zeroCoord, false, invalidInputErr
		}
		flag = true
	}
	/* Subtract 1 because rows and columns are labeled with 1-indexed sequence */
	return Coordinate{row - 1, col - 1}, flag, nil
}

func Game(minefield Grid) {
	isAlive, revealedFields := true, 0
	noMineFields := minefield.rows()*minefield.cols() - minefield.mines()
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
			err = minefield.flag(coord)
			if err != nil {
				fmt.Println(err)
				fmt.Println(Help)
			}
			continue
		}
		isStillAlive, err := minefield.reveal(coord)
		if err != nil {
			fmt.Println(err)
			continue
		}
		isAlive = isStillAlive
		revealedFields++
	}
	fmt.Printf("\n%v\n", minefield.revealAll(isAlive))
	if isAlive {
		fmt.Println("\n🥵 YOU WIN!")
	} else {
		fmt.Println("\nYOU DIE! 🪦")
	}
}

package lang_test

import (
	"strings"
	"testing"

	"image/color"

	"github.com/ukrustacean/figure-display/painter"
	"github.com/ukrustacean/figure-display/painter/lang"
)

func TestParser_Parse(t *testing.T) {
	testInput := `
white
green
update
bgrect 25 35 75 85
figure 120 140
move 90 110
reset
invalid
bgrect wrong args
`

	p := lang.Parser{}
	operations, parseErr := p.Parse(strings.NewReader(testInput))
	
	if parseErr != nil {
		t.Fatalf("parsing failed unexpectedly: %v", parseErr)
	}

	expectedOpCount := 7
	if actualCount := len(operations); actualCount != expectedOpCount {
		t.Fatalf("operation count mismatch: expected %d, actual %d", expectedOpCount, actualCount)
	}

	// Verify white color operation
	colorFillOp, isColorFill := operations[0].(painter.ColorFill)
	if !isColorFill || colorFillOp.Color != color.White {
		t.Errorf("first operation should be white ColorFill, found %+v", operations[0])
	}

	// Verify green color operation
	greenColorOp, isGreenColorFill := operations[1].(painter.ColorFill)
	expectedGreen := color.RGBA{G: 0xff, A: 0xff}
	if !isGreenColorFill || greenColorOp.Color != expectedGreen {
		t.Errorf("second operation should be green ColorFill, found %+v", operations[1])
	}

	// Verify update operation
	if operations[2] != painter.UpdateOp {
		t.Errorf("third operation should be UpdateOp, found %+v", operations[2])
	}

	// Verify background rectangle operation
	bgRectOp, isBgRect := operations[3].(painter.BgRect)
	if !isBgRect || bgRectOp.X1 != 25 || bgRectOp.Y1 != 35 || bgRectOp.X2 != 75 || bgRectOp.Y2 != 85 {
		t.Errorf("fourth operation should be BgRect(25,35,75,85), found %+v", operations[3])
	}

	// Verify figure operation
	figureOp, isFigure := operations[4].(painter.Figure)
	if !isFigure || figureOp.X != 120 || figureOp.Y != 140 {
		t.Errorf("fifth operation should be Figure(120,140), found %+v", operations[4])
	}

	// Verify move operation
	moveOp, isMove := operations[5].(painter.Move)
	if !isMove || moveOp.X != 90 || moveOp.Y != 110 {
		t.Errorf("sixth operation should be Move(90,110), found %+v", operations[5])
	}

	// Verify reset operation
	if _, isReset := operations[6].(painter.Reset); !isReset {
		t.Errorf("seventh operation should be Reset, found %+v", operations[6])
	}
}
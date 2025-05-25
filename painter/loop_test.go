package painter

import (
	"image"
	"image/color"
	"image/draw"
	"reflect"
	"sync"
	"testing"
	"time"

	"golang.org/x/exp/shiny/screen"
)

type testExecutor struct {
	t             *testing.T
	loop          *Loop
	receiver      *mockReceiver
	executedOps   []string
	executionMutex sync.Mutex
}

func newTestExecutor(t *testing.T) *testExecutor {
	receiver := &mockReceiver{}
	executor := &testExecutor{
		t:        t,
		receiver: receiver,
		loop: &Loop{
			Receiver: receiver,
			state:    TextureState{Background: color.Black},
		},
	}
	return executor
}

func (te *testExecutor) recordOperation(name string) Operation {
	return customOperation(func(state TextureState) TextureState {
		te.executionMutex.Lock()
		te.executedOps = append(te.executedOps, name)
		te.executionMutex.Unlock()
		return state
	})
}

func (te *testExecutor) loggedOperation(description string, operation Operation) Operation {
	return customOperation(func(state TextureState) TextureState {
		te.t.Log(description, reflect.TypeOf(operation))
		return operation.Do(state)
	})
}

func (te *testExecutor) executeTest() {
	te.loop.Start(testScreen{})
	defer te.loop.StopAndWait()

	// Execute primary operations
	te.loop.Post(te.loggedOperation("executing white fill", ColorFill{Color: color.White}))
	te.loop.Post(te.loggedOperation("executing green fill", ColorFill{Color: color.RGBA{G: 0xff, A: 0xff}}))
	te.loop.Post(UpdateOp)

	// Execute concurrent operations
	for i := 0; i < 3; i++ {
		go te.loop.Post(te.loggedOperation("executing concurrent green fill", ColorFill{Color: color.RGBA{G: 0xff, A: 0xff}}))
	}

	// Execute ordered test operations
	te.loop.Post(te.recordOperation("op 1"))
	te.loop.Post(te.recordOperation("op 2"))
	te.loop.Post(te.recordOperation("op 3"))

	time.Sleep(100 * time.Millisecond)
}

func (te *testExecutor) validateResults() {
	te.validateTextureUpdate()
	te.validateColorSequence()
	te.validateOperationOrder()
}

func (te *testExecutor) validateTextureUpdate() {
	if te.receiver.lastTexture == nil {
		te.t.Fatal("Texture was not updated")
	}

	mockTex, isCorrectType := te.receiver.lastTexture.(*testTexture)
	if !isCorrectType {
		te.t.Fatal("Unexpected texture type", te.receiver.lastTexture)
	}

	if len(mockTex.Colors) == 0 {
		te.t.Error("No colors were filled in mockTexture, but updates should have occurred.")
		return
	}

	te.validateFinalColor(mockTex)
}

func (te *testExecutor) validateFinalColor(texture *testTexture) {
	finalColor := texture.Colors[len(texture.Colors)-1]
	expectedGreen := color.RGBA{G: 0xff, A: 0xff}

	rgbaColor, isRGBA := finalColor.(color.RGBA)
	if !isRGBA || rgbaColor != expectedGreen {
		te.t.Errorf("Final color validation failed: got %+v, expected %+v. All colors: %+v", 
			finalColor, expectedGreen, texture.Colors)
	}
}

func (te *testExecutor) validateColorSequence() {
	mockTex := te.receiver.lastTexture.(*testTexture)
	if len(mockTex.Colors) > 0 {
		lastColor := mockTex.Colors[len(mockTex.Colors)-1]
		expectedColor := color.RGBA{G: 0xff, A: 0xff}
		
		if rgbaColor, ok := lastColor.(color.RGBA); !ok || rgbaColor != expectedColor {
			te.t.Errorf("Color sequence validation: expected final color %+v, got %+v", expectedColor, lastColor)
		}
	}
}

func (te *testExecutor) validateOperationOrder() {
	expectedOrder := []string{"op 1", "op 2", "op 3"}
	
	te.executionMutex.Lock()
	actualOrder := te.executedOps
	te.executionMutex.Unlock()

	if !reflect.DeepEqual(actualOrder, expectedOrder) {
		te.t.Error("Operation execution order validation failed:", actualOrder)
	}
}

func TestLoop_Post(t *testing.T) {
	executor := newTestExecutor(t)
	executor.executeTest()
	executor.validateResults()
}

// Supporting types and functions
type customOperation func(state TextureState) TextureState

func (op customOperation) Do(state TextureState) TextureState {
	return op(state)
}

type mockReceiver struct {
	lastTexture screen.Texture
	updateMutex sync.Mutex
}

func (r *mockReceiver) Update(texture screen.Texture) {
	r.updateMutex.Lock()
	r.lastTexture = texture
	r.updateMutex.Unlock()
}

type testScreen struct{}

func (s testScreen) NewBuffer(size image.Point) (screen.Buffer, error) {
	panic("NewBuffer not implemented for test")
}

func (s testScreen) NewTexture(size image.Point) (screen.Texture, error) {
	return &testTexture{Colors: []color.Color{}}, nil
}

func (s testScreen) NewWindow(opts *screen.NewWindowOptions) (screen.Window, error) {
	panic("NewWindow not implemented for test")
}

type testTexture struct {
	Colors      []color.Color
	colorsMutex sync.Mutex
}

func (t *testTexture) Release() {}

func (t *testTexture) Size() image.Point { 
	return size 
}

func (t *testTexture) Bounds() image.Rectangle {
	return image.Rectangle{Max: t.Size()}
}

func (t *testTexture) Upload(dp image.Point, src screen.Buffer, sr image.Rectangle) {}

func (t *testTexture) Fill(dr image.Rectangle, src color.Color, op draw.Op) {
	t.colorsMutex.Lock()
	t.Colors = append(t.Colors, src)
	t.colorsMutex.Unlock()
}
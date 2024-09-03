/*
 Copyright 2024 The Knative Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package tui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/multierr"
	"knative.dev/client/pkg/output"
	"knative.dev/client/pkg/output/term"
)

const speedInterval = time.Second / 5

type Progress interface {
	Runnable[ProgressControl]
}

type ProgressControl interface {
	io.Writer
	Error(err error)
}

func (w *widgets) NewProgress(totalSize int, message Message) Progress {
	return &BubbleProgress{
		InputOutput: output.PrinterFrom(w.ctx),
		TotalSize:   totalSize,
		Message:     message,
	}
}

type Message struct {
	Text        string
	PaddingSize int
}

func (m Message) BoundingBoxSize() int {
	mSize := m.TextSize()
	if mSize < m.PaddingSize {
		mSize = m.PaddingSize
	}
	return mSize
}

func (m Message) TextSize() int {
	return len(m.Text)
}

type BubbleProgress struct {
	output.InputOutput
	Message
	TotalSize  int
	FinalPause time.Duration

	prog       progress.Model
	tea        *tea.Program
	downloaded int
	speed      int
	prevSpeed  []int
	err        error
	quitChan   chan struct{}
	teaErr     error
}

func (b *BubbleProgress) With(fn func(ProgressControl) error) error {
	b.start()
	err := func() error {
		defer b.stop()
		return fn(b)
	}()
	return multierr.Combine(err, b.err, b.teaErr)
}

func (b *BubbleProgress) Error(err error) {
	b.err = err
	b.tea.Send(tea.Quit())
}

func (b *BubbleProgress) Write(bytes []byte) (int, error) {
	if b.err != nil {
		return 0, b.err
	}
	noOfBytes := len(bytes)
	b.downloaded += noOfBytes
	b.speed += noOfBytes
	if b.TotalSize > 0 {
		percent := float64(b.downloaded) / float64(b.TotalSize)
		b.onProgress(percent)
	}
	return noOfBytes, nil
}

func (b *BubbleProgress) Init() tea.Cmd {
	return b.tickSpeed()
}

func (b *BubbleProgress) View() string {
	return b.display(b.prog.View()) +
		"\n" + helpStyle("Press Ctrl+C to cancel")
}

func helpStyle(str string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(str)
}

func (b *BubbleProgress) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	handle := bubbleProgressHandler{b}
	switch event := msg.(type) {
	case tea.WindowSizeMsg:
		return handle.windowSize(event)

	case tea.KeyMsg:
		return handle.keyPressed(event)

	case speedChange:
		return handle.speedChange()

	case percentChange:
		return handle.percentChange(event)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		return handle.progressFrame(event)

	default:
		return b, nil
	}
}

type bubbleProgressHandler struct {
	*BubbleProgress
}

func (b bubbleProgressHandler) windowSize(event tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	const percentLen = 4
	b.prog.Width = event.Width - len(b.display("")) + percentLen
	return b, nil
}

func (b bubbleProgressHandler) keyPressed(event tea.KeyMsg) (tea.Model, tea.Cmd) {
	if event.Type == tea.KeyCtrlC {
		b.err = context.Canceled
		return b, tea.Quit
	}
	return b, nil
}

func (b bubbleProgressHandler) speedChange() (tea.Model, tea.Cmd) {
	b.prevSpeed = append(b.prevSpeed, b.speed)
	const speedAvgAmount = 4
	if len(b.prevSpeed) > speedAvgAmount {
		b.prevSpeed = b.prevSpeed[1:]
	}
	b.speed = 0
	if b.downloaded < b.TotalSize {
		return b, b.tickSpeed()
	}
	return b, nil
}

func (b bubbleProgressHandler) percentChange(event percentChange) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0, 1)
	cmds = append(cmds, b.prog.SetPercent(float64(event)))

	if event >= 1.0 {
		cmds = append(cmds, b.quitSignal())
	}

	return b, tea.Batch(cmds...)
}

func (b *BubbleProgress) quitSignal() tea.Cmd {
	// The final pause is to give the progress bar a chance to finish its
	// animation before quitting. Otherwise, it ends abruptly, and the user
	// might not see the progress bar at 100%.
	return tea.Sequence(b.finalPause(), tea.Quit)
}

func (b bubbleProgressHandler) progressFrame(event progress.FrameMsg) (tea.Model, tea.Cmd) {
	progressModel, cmd := b.prog.Update(event)
	if m, ok := progressModel.(progress.Model); ok {
		b.prog = m
	}
	return b, cmd
}

func (b *BubbleProgress) display(bar string) string {
	const padding = 2
	const pad = " ⋮ "
	paddingLen := padding + b.Message.BoundingBoxSize() - b.Message.TextSize()
	titlePad := strings.Repeat(" ", paddingLen)
	total := humanizeBytes(float64(b.TotalSize), "")
	totalFmt := fmt.Sprintf("%6.2f %-3s", total.num, total.unit)
	return b.Message.Text + titlePad + bar + pad + b.speedFormatted() +
		pad + totalFmt
}

func (b *BubbleProgress) speedFormatted() string {
	s := humanizeBytes(b.speedPerSecond(), "/s")
	return fmt.Sprintf("%6.2f %-5s", s.num, s.unit)
}

func (b *BubbleProgress) speedPerSecond() float64 {
	speed := 0.
	for _, s := range b.prevSpeed {
		speed += float64(s)
	}
	if len(b.prevSpeed) > 0 {
		speed /= float64(len(b.prevSpeed))
	}
	return speed / float64(speedInterval.Microseconds()) *
		float64(time.Second.Microseconds())
}

func (b *BubbleProgress) tickSpeed() tea.Cmd {
	return tea.Every(speedInterval, func(ti time.Time) tea.Msg {
		return speedChange{}
	})
}

func (b *BubbleProgress) start() {
	b.prog = progress.New(progress.WithDefaultGradient())
	b.tea = tea.NewProgram(b, ioProgramOptions(b.InputOutput)...)
	b.quitChan = make(chan struct{})
	go func() {
		t := b.tea
		if _, err := t.Run(); err != nil {
			b.teaErr = err
		}
		close(b.quitChan)
	}()
}

func (b *BubbleProgress) stop() {
	if b.tea == nil {
		return
	}

	b.tea.Send(b.quitSignal())
	<-b.quitChan

	if term.IsWriterTerminal(b.OutOrStdout()) && b.teaErr == nil {
		b.teaErr = b.tea.ReleaseTerminal()
	}

	b.tea = nil
	b.quitChan = nil
}

func (b *BubbleProgress) onProgress(percent float64) {
	b.tea.Send(percentChange(percent))
}

func (b *BubbleProgress) finalPause() tea.Cmd {
	pause := b.FinalPause
	if pause == 0 {
		pause = speedInterval * 3
	}
	return tea.Tick(pause, func(_ time.Time) tea.Msg {
		return nil
	})
}

type percentChange float64

type speedChange struct{}

type humanByteSize struct {
	num  float64
	unit string
}

func humanizeBytes(bytes float64, unitSuffix string) humanByteSize {
	num := bytes
	units := []string{
		"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB",
	}
	i := 0
	const kilo = 1024
	for num > kilo && i < len(units)-1 {
		num /= kilo
		i++
	}
	return humanByteSize{
		num:  num,
		unit: units[i] + unitSuffix,
	}
}

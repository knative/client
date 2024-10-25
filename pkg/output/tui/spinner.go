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
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/multierr"
	"knative.dev/client/pkg/output"
	"knative.dev/client/pkg/output/term"
)

const spinnerColor = lipgloss.Color("205")

type Spinner interface {
	Runnable[SpinnerControl]
}

func (w *widgets) NewSpinner(message string) Spinner {
	return &BubbleSpinner{
		InputOutput: output.PrinterFrom(w.ctx),
		Message:     Message{Text: message},
	}
}

type BubbleSpinner struct {
	output.InputOutput
	Message

	*updater
	spin     spinner.Model
	tea      *tea.Program
	quitChan chan struct{}
	teaErr   error
}

// SpinnerControl allows one to control the spinner, for example, to change the
// message.
type SpinnerControl interface {
	UpdateMessage(message string)
}

// With will start the spinner and perform long operation within the
// provided fn. The spinner will be automatically shutdown when the provided
// function exits.
func (b *BubbleSpinner) With(fn func(SpinnerControl) error) error {
	b.start()
	err := func() error {
		defer b.stop()
		return fn(b.updater)
	}()
	return multierr.Combine(err, b.teaErr)
}

func (b *BubbleSpinner) Init() tea.Cmd {
	return b.spin.Tick
}

func (b *BubbleSpinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, c := b.spin.Update(msg)
	b.spin = m
	return b, c
}

func (b *BubbleSpinner) View() string {
	select {
	case m := <-b.updater.messages:
		// nil on channel close
		if m != nil {
			b.Message.Text = *m
		}
	default:
		// nothing
	}
	return fmt.Sprintf("%s %s", b.Message.Text, b.spin.View())
}

func (b *BubbleSpinner) start() {
	b.updater = &updater{make(chan *string)}
	b.spin = spinner.New(
		spinner.WithSpinner(spinner.Meter),
		spinner.WithStyle(spinnerStyle()),
	)
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

func (b *BubbleSpinner) stop() {
	if b.tea == nil {
		return
	}

	close(b.updater.messages)
	b.tea.Quit()
	<-b.quitChan

	if term.IsWriterTerminal(b.OutOrStdout()) && b.teaErr == nil {
		b.teaErr = b.tea.ReleaseTerminal()
	}

	b.tea = nil
	b.quitChan = nil
	endMsg := fmt.Sprintf("%s %s\n",
		b.Message.Text, spinnerStyle().Render("Done"))
	_, _ = b.OutOrStdout().Write([]byte(endMsg))
}

func spinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(spinnerColor)
}

type updater struct {
	messages chan *string
}

func (u updater) UpdateMessage(message string) {
	u.messages <- &message
}

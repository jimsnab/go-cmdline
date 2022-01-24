package cmdline

import (
	"fmt"
	"math"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

type terminalData interface {
	IsTerminal(fd int) bool
	GetSize(fd int) (width int, height int, err error)
}

type defaultTerminal struct {
}

func (t *defaultTerminal) IsTerminal(fd int) bool {
	return term.IsTerminal(fd)
}

func (t *defaultTerminal) GetSize(fd int) (int, int, error) {
	return term.GetSize(fd)
}

var xterm = terminalData(&defaultTerminal{})

type Printer interface {
	Status(text string)
	Clear()
	ChattyStatus(text string)
	SetCounterMax(text string, max int)
	UpdateCountStatus(extraStatusText string)
	Count()
	PauseStatus()
	ResumeStatus()
	Println(text string)
	BeginPrint(text string)
	ContinuePrint(text string)
	EndPrint(text string)
	EndPrintIfStarted()
	DateRangeStatus(from time.Time, to time.Time, purpose string)
}

var Prn Printer = &defaultPrinter{}

func SetPrinter(printer Printer) Printer {
	priorPrinter := Prn
	Prn = printer
	return priorPrinter
}

const simpleTimeFormat = "2006-01-02 15:04:05 MST"

type defaultPrinter struct {
	pauseCount            int
	lastStatus            time.Time
	lastStatusText        string
	lastPrintedStatusText string
	storedStatus          string
	counterText           string
	counter               int
	maxCounter            int
	nestedPrint           bool
}

func (dp *defaultPrinter) Status(text string) {
	dp.lastStatusText = text // lastStatusText is the true last status message, printed or not

	if dp.pauseCount > 0 {
		dp.storedStatus = text
		return
	}

	if !xterm.IsTerminal(syscall.Stdout) {
		return
	}

	width, _, err := xterm.GetSize(syscall.Stdout)
	if err != nil {
		return
	}

	textLength := utf8.RuneCountInString(text)
	if textLength >= width {
		textLength = width - 1
		text = substr(text, 0, textLength)
	}

	maxBase := utf8.RuneCountInString(dp.lastPrintedStatusText)
	if textLength < maxBase {
		maxBase = textLength
	}

	textRune := []rune(text)
	lastRune := []rune(dp.lastPrintedStatusText)

	// determine the common base between what was printed and what is to be printed
	var baseLength int
	for baseLength = 0; baseLength < maxBase; baseLength++ {
		if textRune[baseLength] != lastRune[baseLength] {
			break
		}
	}

	// from the end of the last text, backspace until the new shorter base is reached
	if baseLength < len(lastRune) {
		fmt.Print(strings.Repeat("\b", len(lastRune)-baseLength))
	}

	// print the part of the new text that is different from the last
	fmt.Print(substr(text, baseLength, textLength-baseLength))

	// if new text is shorter than the last text, erase extra right side characters
	eraseLength := 0
	if textLength < len(lastRune) {
		eraseLength = len(lastRune) - textLength
		fmt.Print(strings.Repeat(" ", eraseLength))
		fmt.Print(strings.Repeat("\b", eraseLength))
	}

	dp.lastPrintedStatusText = text
	dp.lastStatus = time.Now()
}

func (dp *defaultPrinter) Clear() {
	dp.SetCounterMax("", 0)
	dp.Status("")
}

func (dp *defaultPrinter) ChattyStatus(text string) {
	secondAgo := time.Now().Add(-1 * time.Second)
	if dp.lastStatus.Before(secondAgo) {
		dp.Status(text)
	}
	dp.lastStatusText = text // lastStatusText changes even if not printed
}

func (dp *defaultPrinter) SetCounterMax(text string, max int) {
	dp.counterText = text
	dp.counter = 0
	dp.maxCounter = max
}

func (dp *defaultPrinter) count(extraStatusText string) {
	if dp.maxCounter > 0 {
		dp.counter++

		c := dp.counter
		if c > dp.maxCounter {
			c = dp.maxCounter
		}
		f := (float64(c) * 100.0) / float64(dp.maxCounter)
		percentage := int(math.Round(f))
		text := fmt.Sprintf("%s %d of %d %d%%", dp.counterText, c, dp.maxCounter, percentage)

		if percentage < 99 {
			lastQuarterSecond := time.Now().Add(-250 * time.Millisecond)
			if dp.lastStatus.After(lastQuarterSecond) {
				dp.lastStatusText = text // lastStatusText changes even if not printed
				return
			}
		}

		if len(extraStatusText) == 0 {
			dp.Status(text)
		} else {
			dp.Status(text + " " + extraStatusText)
		}
	}
}

func (dp *defaultPrinter) UpdateCountStatus(extraStatusText string) {
	if dp.maxCounter > 0 {
		dp.counter-- // decrement, then increment in dp.count(), for a net zero counter change
		dp.count(extraStatusText)
	}
}

func (dp *defaultPrinter) Count() {
	dp.count("")
}

func (dp *defaultPrinter) PauseStatus() {
	if dp.pauseCount == 0 {
		dp.storedStatus = dp.lastStatusText
		dp.Status("")
	}
	dp.pauseCount++
}

func (dp *defaultPrinter) ResumeStatus() {
	if dp.pauseCount == 0 {
		return
	}

	dp.pauseCount--
	if dp.pauseCount == 0 {
		dp.Status(dp.storedStatus)
	}
}

func (dp *defaultPrinter) Println(text string) {
	if dp.nestedPrint {
		panic(fmt.Errorf("in a nested print"))
	}

	dp.PauseStatus()
	fmt.Println(text)
	dp.ResumeStatus()
}

func (dp *defaultPrinter) BeginPrint(text string) {
	if dp.nestedPrint {
		panic(fmt.Errorf("in a nested print"))
	}
	dp.PauseStatus()
	if len(text) > 0 {
		fmt.Print(text)
	}
	dp.nestedPrint = true
}

func (dp *defaultPrinter) ContinuePrint(text string) {
	if !dp.nestedPrint {
		panic(fmt.Errorf("segmented printing didn't begin yet"))
	}
	if len(text) > 0 {
		fmt.Print(text)
	}
}

func (dp *defaultPrinter) EndPrint(text string) {
	if !dp.nestedPrint {
		panic(fmt.Errorf("segmented printing didn't begin yet"))
	}
	fmt.Println(text)
	dp.ResumeStatus()
	dp.nestedPrint = false
}

func (dp *defaultPrinter) EndPrintIfStarted() {
	if dp.nestedPrint {
		dp.EndPrint("")
	}
}

func (dp *defaultPrinter) DateRangeStatus(from time.Time, to time.Time, purpose string) {
	if from.Equal(to) {
		dp.Status(purpose + " for " + from.Format(simpleTimeFormat))
	} else {
		dp.Status(purpose + " between " + from.Format(simpleTimeFormat) + " and " + to.Format(simpleTimeFormat))
	}
}

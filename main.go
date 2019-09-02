package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"io"
	"net"
	"os"
	"strings"
)

var (
	exitCode = 0
	conn net.Conn
	connected = false
	history []string
)

func main() {
	err := mainCode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	os.Exit(exitCode)
}

func mainCode() error {
	app := tview.NewApplication()

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetDoneFunc(func(key tcell.Key) {
		})

	input := tview.NewInputField()

	// state for input bar
	inputState := "READY"
	historyPointer := 0

	input.SetDoneFunc(func(key tcell.Key) {

		// mini state machine to control input box
		switch(inputState) {
		case "READY":
			if key == tcell.KeyUp {
				if len(history) > 0 {
					inputState = "HISTORY_SCROLL"
					historyPointer = len(history) - 1
					input.SetText(history[historyPointer])
				}
			}
		case "HISTORY_SCROLL":
			if key == tcell.KeyUp {
				if historyPointer >= 1 && historyPointer <= len(history) - 1 {
					historyPointer = historyPointer - 1
					input.SetText(history[historyPointer])
				}
			} else if key == tcell.KeyDown {
				if historyPointer >= 0 && historyPointer <= len(history) - 2 {
					historyPointer = historyPointer + 1
					input.SetText(history[historyPointer])
				}
			} else {
				inputState ="READY"
			}
		}

		if key == tcell.KeyEnter {
			buffer := input.GetText()

			if len(buffer) == 0 {
				if connected {
					fmt.Fprintf(conn,"\r\n")
				}
				return
			}

			if len(history) > 0 {
				if buffer != history[len(history)-1] {
					history = append(history, buffer)
				}
			} else {
				history = append(history, buffer)
			}

			if buffer[0] == '/' {
				parts := strings.Split(buffer, " ")
				switch parts[0] {
				case "/quit":
					if connected {
						conn.Close()
					}
					app.Stop()
				case "/connect":
					fmt.Fprintf(textView, "[green]Connecting to %s[white]\n", parts[1])
					host := parts[1]
					if len(parts) == 2 {
						go func(){
							var err error
							conn, err = net.Dial("tcp", host)
							if err != nil {
								fmt.Fprintf(textView, "[red]%s[white]\n", err)
								app.Draw()
								return
							}
							fmt.Fprintf(textView, "[green]Connected![white]")
							connected = true
							app.Draw()
							buf := make([]byte, 10000)

							for {
								n, err := conn.Read(buf)
								if err == io.EOF {
									fmt.Fprintf(textView, "[red]disconnected[white]")
									app.Draw()
									return
								}
								if err != nil {
									fmt.Fprintf(textView, "[red]error: %s[white]", err)
									app.Draw()
									return
								}
								if n > 0 {
									outLine := strings.Replace(string(buf[:n]), "\xff\xf9", "", -1)
									outLine = strings.Replace(outLine, "\xff\xfd", "", -1)
									outLine = strings.Replace(outLine, "\r", "", -1)
									outLine = strings.Replace(outLine, "\u001b[30m", "[black]", -1)
									outLine = strings.Replace(outLine, "\u001b[31m", "[red]", -1)
									outLine = strings.Replace(outLine, "\u001b[32m", "[green]", -1)
									outLine = strings.Replace(outLine, "\u001b[33m", "[yellow]", -1)
									outLine = strings.Replace(outLine, "\u001b[34m", "[blue]", -1)
									outLine = strings.Replace(outLine, "\u001b[35m", "[magenta]", -1)
									outLine = strings.Replace(outLine, "\u001b[36m", "[cyan]", -1)
									outLine = strings.Replace(outLine, "\u001b[37m", "[white]", -1)
									outLine = strings.Replace(outLine, "\u001b[40m", "[:black:]", -1)
									outLine = strings.Replace(outLine, "\u001b[1m", "[::b]", -1)

									// for some reason reset doesn't work properly
									outLine = strings.Replace(outLine, "\u001b[0m", "", -1)
									fmt.Fprintf(textView, "%+v", outLine)
									app.Draw()
								}
							}
						}()
					}
				default:
				}
			} else {
				fmt.Fprintf(textView, "%s\n", buffer)
				if connected {
					fmt.Fprintf(conn,"%s\r\n", buffer)
				}
			}

			input.SetText("")
		}
	})

	textView.SetBorder(false).SetTitle("GoMUD")

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(input, 1, 1, true)

	if err := app.SetRoot(layout, true).Run(); err != nil {
		exitCode = 1
		return err
	}

	return nil
}

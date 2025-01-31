package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	cfm "github.com/0l1v3rr/cli-file-manager/pkg"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/skratchdot/open-golang/open"
)

const VERSION = "v1.2.0"

var (
	path       string
	l               = widgets.NewList()
	p               = widgets.NewParagraph()
	p2              = widgets.NewParagraph()
	p3              = widgets.NewParagraph()
	showHidden bool = true
	showEx     bool = true
	catMode    bool = false
)

func main() {

	defaultPath, err := os.Getwd()
	if err != nil {
		fmt.Println("An error occurred while reading the current path.")
		defaultPath = "/"
	}

	if len(os.Args) > 1 && os.Args[1] != "" {
		path = os.Args[1]
	} else {
		path = defaultPath
	}

	err2 := ui.Init()
	if err2 != nil {
		fmt.Println("Failed to initialize termui.")
		return
	}
	defer ui.Close()

	initWidgets()
}

func initWidgets() {
	pText := `[↑](fg:green) - Scroll Up
	[↓](fg:green) - Scroll Down
	[q](fg:green) - Quit
	[Enter](fg:green) - Open
	[m](fg:green) - Memory Usage
	[f](fg:green) - Disk Information
	[^D (2 times)](fg:green) - Remove file
	[^F](fg:green) - Create file
	[^N](fg:green) - Create folder
	[^R](fg:green) - Rename file
	[^L](fg:green) - Duplicate file
	[^V](fg:green) - Launch VS Code
	[^T](fg:green) - Read file content
	[C](fg:green) - Copy file
	[h](fg:green) - Hide hidden files
	[e](fg:green) - Hide file extension
	`

	disk := cfm.DiskUsage("/")
	initItems(pText, disk)

	selectedRowNum := 0
	copyPath := ""
	previousKey := ""
	inputField := ""
	originalName := ""
	fileCreatingInProgress := false
	dirCreatingInProgress := false
	renameInProgress := false
	uiEvents := ui.PollEvents()

	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "<Down>":
			if !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				l.ScrollDown()
				if !catMode {
					p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
				}
			}
		case "<Up>":
			if !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				l.ScrollUp()
				if !catMode {
					p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
				}
			}
		case "<Home>":
			if !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				l.ScrollTop()
				if !catMode {
					p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
				}
			}
		case "<End>":
			if !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				l.ScrollBottom()
				if !catMode {
					p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
				}
			}
		case "<Space>":
			if !catMode {
				p2.Text = cfm.GetFileInformationsWithSize(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
			}
		case "<C-l>":
			if !catMode && !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				selected := getFileName(l.SelectedRow)
				if selected[len(selected)-1] != '/' {
					err := cfm.Duplicate(fmt.Sprintf("%s/%s", path, selected), path)
					if err != nil {
						errorMsg("An unknown error occurred while duplicating the file.")
					}
					l.Rows = cfm.ReadFiles(path, showHidden)
					ui.Render(l, p, p2, p3)
				}
			}
		case "<C-t>":
			selected := getFileName(l.SelectedRow)
			selectedRowNum = l.SelectedRow
			if selected[len(selected)-1] != '/' {
				catMode = true
				cat(path, selected)
			}
		case "<C-d>":
			if !catMode && !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				if previousKey == "<C-d>" {
					selected := getFileName(l.SelectedRow)
					if selected != ".." && selected != "../" {
						filePath := ""
						if path[len(path)-1] == '/' || selected[0] == '/' {
							filePath = fmt.Sprintf("%v%v", path, selected)
						} else {
							filePath = fmt.Sprintf("%v/%v", path, selected)
						}
						err := os.Remove(filePath)
						if err == nil {
							l.Rows = cfm.ReadFiles(path, showHidden)
							l.SelectedRow = 0
							p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
						} else {
							err2 := os.RemoveAll(filePath)
							if err2 == nil {
								l.Rows = cfm.ReadFiles(path, showHidden)
								l.SelectedRow = 0
								p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
							} else {
								errorMsg("An unknown error occurred while deleting the file.")
							}
						}
					} else {
						errorMsg("You can't delete this!")
					}
				}
			}
		case "<C-f>":
			if !catMode && !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				fileCreatingInProgress = true
				l.Rows = append(l.Rows, fmt.Sprintf("[?]: %v", inputField))
				l.SelectedRow = len(l.Rows) - 1
				textFieldStyle()
				p2.Text = cfm.EmptyFileInfo()
				p.Text = "[Esc](fg:green) - Cancel\n[Enter](fg:green) - Apply Changes\n"
			}
		case "<C-n>":
			if !catMode && !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				dirCreatingInProgress = true
				l.Rows = append(l.Rows, fmt.Sprintf("[$]: %v", inputField))
				l.SelectedRow = len(l.Rows) - 1
				textFieldStyle()
				p2.Text = cfm.EmptyFileInfo()
				p.Text = "[Esc](fg:green) - Cancel\n[Enter](fg:green) - Apply Changes\n"
			}
		case "<C-r>":
			if !catMode && !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				renameInProgress = true
				originalName = l.Rows[l.SelectedRow]
				inputField = getFileName(l.SelectedRow)
				if inputField[len(inputField)-1] == '/' {
					inputField = inputField[:len(inputField)-1]
				}
				l.Rows[l.SelectedRow] = fmt.Sprintf("[#]: %v", inputField)
				textFieldStyle()
				p.Text = "[Esc](fg:green) - Cancel\n[Enter](fg:green) - Apply Changes\n"
			}
		case "<Escape>":
			resetColors()
			if copyPath != "" {
				copyPath = ""
				p.Text = pText
			} else if catMode {
				backToCfm(pText, disk, selectedRowNum)
			} else {
				if fileCreatingInProgress {
					fileCreatingInProgress = false
					inputField = ""
					l.SelectedRow = 0
					l.Rows = l.Rows[:len(l.Rows)-1]
					p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
					p.Text = pText
				} else if dirCreatingInProgress {
					dirCreatingInProgress = false
					inputField = ""
					l.SelectedRow = 0
					l.Rows = l.Rows[:len(l.Rows)-1]
					p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
					p.Text = pText
				} else if renameInProgress {
					renameInProgress = false
					inputField = ""
					l.Rows[l.SelectedRow] = originalName
					originalName = ""
					p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
					p.Text = pText
				}
			}
		case "m":
			if !catMode && !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				p3.Title = "Memory Usage"
				p3.Text = cfm.ReadMemStats()
			}
		case "f":
			if !catMode && !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				p3.Title = "Disk Information"
				p3.Text = fmt.Sprintf(`[All: ](fg:green) - %.2f GB
				[Used:](fg:green) - %.2f GB
				[Free:](fg:green) - %.2f GB
				`, float64(disk.All)/float64(1024*1024*1024), float64(disk.Used)/float64(1024*1024*1024), float64(disk.Free)/float64(1024*1024*1024))
			}
		case "c":
			if !catMode {
				if getFileName(l.SelectedRow)[len(getFileName(l.SelectedRow))-1] != '/' {
					copyPath = fmt.Sprintf("%s/%s", path, getFileName(l.SelectedRow))
					p.Text = fmt.Sprintf(`[↑](fg:green) - Scroll Up
					[↓](fg:green) - Scroll Down
					[q](fg:green) - Quit
					[Enter](fg:green) - Open
					[m](fg:green) - Memory Usage
					[f](fg:green) - Disk Information
					[^D (2 times)](fg:green) - Remove file
					[^F](fg:green) - Create file
					[^N](fg:green) - Create folder
					[^R](fg:green) - Rename file
					[^L](fg:green) - Duplicate file
					[^V](fg:green) - Launch VS Code
					[C](fg:cyan) - Copied to clipboard ([%s](fg:cyan))
					[V](fg:green) - Paste
					`, getFileName(l.SelectedRow))
				}
			}
		case "v":
			if !catMode {
				if copyPath != "" {
					p.Text = `[↑](fg:green) - Scroll Up
					[↓](fg:green) - Scroll Down
					[q](fg:green) - Quit
					[Enter](fg:green) - Open
					[m](fg:green) - Memory Usage
					[f](fg:green) - Disk Information
					[^D (2 times)](fg:green) - Remove file
					[^F](fg:green) - Create file
					[^N](fg:green) - Create folder
					[^R](fg:green) - Rename file
					[^L](fg:green) - Duplicate file
					[^V](fg:green) - Launch VS Code
					[C](fg:cyan) - Copying...
					`
					cfm.Copy(copyPath, path)
					p.Text = pText
					copyPath = ""
					l.Rows = cfm.ReadFiles(path, showHidden)
					ui.Render(l, p, p2, p3)
				} else {
					p.Text = pText
				}
			}
		case "<Resize>":
			if !catMode {
				l.SetRect(0, 0, cfm.GetCliWidth()/2, int(float64(cfm.GetCliHeight())*0.73))
				p.SetRect(cfm.GetCliWidth()/2, 0, cfm.GetCliWidth(), int(float64(cfm.GetCliHeight())*0.58))
				p2.SetRect(0, cfm.GetCliHeight(), cfm.GetCliWidth(), int(float64(cfm.GetCliHeight())*0.73))
				p3.SetRect(cfm.GetCliWidth()/2, int(float64(cfm.GetCliHeight())*0.73), cfm.GetCliWidth(), int(float64(cfm.GetCliHeight())*0.58))
				ui.Render(l, p, p2, p3)
			} else {
				p.SetRect(0, 0, cfm.GetCliWidth(), 3)
				l.SetRect(0, 3, cfm.GetCliWidth(), int(float64(cfm.GetCliHeight())))
				p2.SetRect(cfm.GetCliWidth()+1, cfm.GetCliHeight()+1, cfm.GetCliWidth()+1, cfm.GetCliHeight()+1)
				p3.SetRect(cfm.GetCliWidth()+1, cfm.GetCliHeight()+1, cfm.GetCliWidth()+1, cfm.GetCliHeight()+1)
			}
		case "<C-v>":
			if !catMode && !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				cmd := exec.Command("code", path)
				err := cmd.Run()
				if err != nil {
					errorMsg("Unable to open VS Code.")
				}
			}
		case "h":
			if !catMode && !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress && copyPath == "" {
				if showHidden {
					p.Text = `[↑](fg:green) - Scroll Up
					[↓](fg:green) - Scroll Down
					[q](fg:green) - Quit
					[Enter](fg:green) - Open
					[m](fg:green) - Memory Usage
					[f](fg:green) - Disk Information
					[^D (2 times)](fg:green) - Remove file
					[^F](fg:green) - Create file
					[^N](fg:green) - Create folder
					[^R](fg:green) - Rename file
					[^L](fg:green) - Duplicate file
					[^V](fg:green) - Launch VS Code
					[^T](fg:green) - Read file content
					[C](fg:green) - Copy file
					[h](fg:green) - Show hidden files
					[e](fg:green) - Hide file extension
					`
					showEx = true
				} else {
					p.Text = pText
				}
				showHidden = !showHidden
				l.Rows = cfm.ReadFiles(path, showHidden)
			}
		case "e":
			if !catMode && !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress && copyPath == "" {
				showEx = !showEx
				if showEx {
					l.Rows = cfm.ReadFiles(path, showHidden)
					p.Text = pText
				} else {
					l.Rows = cfm.NoEx(path)
					p.Text = `[↑](fg:green) - Scroll Up
					[↓](fg:green) - Scroll Down
					[q](fg:green) - Quit
					[Enter](fg:green) - Open
					[m](fg:green) - Memory Usage
					[f](fg:green) - Disk Information
					[^D (2 times)](fg:green) - Remove file
					[^F](fg:green) - Create file
					[^N](fg:green) - Create folder
					[^R](fg:green) - Rename file
					[^L](fg:green) - Duplicate file
					[^V](fg:green) - Launch VS Code
					[^T](fg:green) - Read file content
					[C](fg:green) - Copy file
					[h](fg:green) - Hide hidden files
					[e](fg:green) - Show file extension
					`
					showHidden = true
				}
			}
		case "<Enter>":
			if !catMode && !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
				selected := getFileName(l.SelectedRow)
				if selected[len(selected)-1] == '/' {
					if selected == "../" {
						splitted := strings.Split(path, "/")
						if len(splitted) > 0 {
							if len(splitted) == 2 {
								path = "/"
							} else {
								path = strings.Join(splitted[:len(splitted)-2], "/")
							}
						} else {
							path = "/"
						}
					} else {
						if path[len(path)-1] == '/' || selected[0] == '/' {
							path = fmt.Sprintf("%v%v", path, selected)
						} else {
							path = fmt.Sprintf("%v/%v", path, selected)
						}
					}
					l.Rows = cfm.ReadFiles(path, showHidden)

					l.SelectedRow = 0
					l.SelectedRowStyle.Fg = ui.ColorBlue
					l.SelectedRowStyle.Modifier = ui.ModifierBold
					p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
				} else {
					var filePath string
					if path[len(path)-1] == '/' || selected[0] == '/' {
						filePath = fmt.Sprintf("%v%v", path, selected)
					} else {
						filePath = fmt.Sprintf("%v/%v", path, selected)
					}
					open.Start(filePath)
				}
			} else if fileCreatingInProgress {
				if len(inputField) >= 3 {
					err := ioutil.WriteFile(fmt.Sprintf("%v/%v", path, inputField), []byte(""), 0755)
					if err == nil {
						l.Rows = cfm.ReadFiles(path, showHidden)
						l.SelectedRow = 0
						inputField = ""
						fileCreatingInProgress = false
						resetColors()
						p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
						p.Text = pText
					} else {
						errorMsg("An unknown error occurred while creating the file.")
					}
				}
			} else if dirCreatingInProgress {
				if len(inputField) >= 3 {
					err := os.Mkdir(fmt.Sprintf("%v/%v", path, inputField), 0755)
					if err == nil {
						l.Rows = cfm.ReadFiles(path, showHidden)
						l.SelectedRow = 0
						inputField = ""
						dirCreatingInProgress = false
						resetColors()
						p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
						p.Text = pText
					} else {
						errorMsg("An unknown error occurred while creating the folder.")
					}
				}
			} else if renameInProgress {
				if len(inputField) >= 3 {
					original := getFileNameByFullName(originalName)
					err := os.Rename(fmt.Sprintf("%v/%v", path, original), fmt.Sprintf("%v/%v", path, inputField))
					if err == nil {
						l.Rows = cfm.ReadFiles(path, showHidden)
						inputField = ""
						originalName = ""
						original = ""
						renameInProgress = false
						resetColors()
						p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
						p.Text = pText
					} else {
						errorMsg("An unknown error occurred while renaming the file.")
					}
				}
			}
		}

		if fileCreatingInProgress {
			if e.ID[0] != '<' {
				inputField = inputField + e.ID
				l.Rows[len(l.Rows)-1] = fmt.Sprintf("[?]: %v", inputField)
			} else if e.ID == "<Backspace>" {
				le := len(inputField)
				if le > 0 {
					inputField = inputField[:le-1]
				}
				l.Rows[len(l.Rows)-1] = fmt.Sprintf("[?]: %v", inputField)
			}
		} else if dirCreatingInProgress {
			if e.ID[0] != '<' {
				inputField = inputField + e.ID
				l.Rows[len(l.Rows)-1] = fmt.Sprintf("[$]: %v", inputField)
			} else if e.ID == "<Backspace>" {
				le := len(inputField)
				if le > 0 {
					inputField = inputField[:le-1]
				}
				l.Rows[len(l.Rows)-1] = fmt.Sprintf("[$]: %v", inputField)
			}
		} else if renameInProgress {
			if e.ID[0] != '<' {
				inputField = inputField + e.ID
				l.Rows[l.SelectedRow] = fmt.Sprintf("[#]: %v", inputField)
			} else if e.ID == "<Backspace>" {
				le := len(inputField)
				if le > 0 {
					inputField = inputField[:le-1]
				}
				l.Rows[l.SelectedRow] = fmt.Sprintf("[#]: %v", inputField)
			}
		}

		if !catMode && !fileCreatingInProgress && !dirCreatingInProgress && !renameInProgress {
			if previousKey == "<C-d>" {
				previousKey = ""
			} else {
				previousKey = e.ID
			}
		}

		ui.Render(l, p, p2, p3)
	}
}

func errorMsg(err string) {
	p3.Title = "Error"
	p3.Text = fmt.Sprintf("[%s](fg:red)", err)
}

func getFileName(n int) string {
	row := l.Rows[n]
	sliced := strings.Split(strings.Replace(row, "](fg:green)", "", 1), " ")
	sliced = sliced[1:]
	result := strings.Join(sliced, " ")

	return result
}

func getFileNameByFullName(s string) string {
	sliced := strings.Split(strings.Replace(s, "](fg:green)", "", 1), " ")
	sliced = sliced[1:]
	result := strings.Join(sliced, " ")

	return result
}

func initItems(pText string, disk cfm.Status) {
	l.Title = fmt.Sprintf("CLI File Manager - %s", VERSION)
	l.Rows = cfm.ReadFiles(path, showHidden)
	l.TextStyle = ui.NewStyle(ui.ColorWhite)
	l.WrapText = false

	l.SetRect(0, 0, cfm.GetCliWidth()/2, int(float64(cfm.GetCliHeight())*0.73))
	//w, h := ui.TerminalDimensions()
	//l.SetRect(0, 0, w/2, int(float64(h)*0.73))

	l.BorderStyle.Fg = ui.ColorBlue
	l.TitleStyle.Modifier = ui.ModifierBold
	l.SelectedRowStyle.Fg = ui.ColorBlue
	l.SelectedRowStyle.Modifier = ui.ModifierBold

	p.Title = "Help Menu"
	p.Text = pText
	p.SetRect(cfm.GetCliWidth()/2, 0, cfm.GetCliWidth(), int(float64(cfm.GetCliHeight())*0.58))
	//p.SetRect(w/2, 0, w, int(float64(h)*0.58))
	p.BorderStyle.Fg = ui.ColorBlue
	p.TitleStyle.Modifier = ui.ModifierBold

	json, err := cfm.ReadJson()
	if json == "memory" || err != nil {
		p3.Title = "Memory Usage"
		p3.Text = cfm.ReadMemStats()
	} else {
		p3.Title = "Disk Information"
		p3.Text = fmt.Sprintf(`[All: ](fg:green) - %.2f GB
		[Used:](fg:green) - %.2f GB
		[Free:](fg:green) - %.2f GB
		`, float64(disk.All)/float64(1024*1024*1024), float64(disk.Used)/float64(1024*1024*1024), float64(disk.Free)/float64(1024*1024*1024))
	}
	p3.Border = true
	p3.SetRect(cfm.GetCliWidth()/2, int(float64(cfm.GetCliHeight())*0.73), cfm.GetCliWidth(), int(float64(cfm.GetCliHeight())*0.58))
	//p3.SetRect(w/2, int(float64(h)*0.73), w, int(float64(h)*0.58))
	p3.BorderStyle.Fg = ui.ColorBlue
	p3.TitleStyle.Modifier = ui.ModifierBold

	p2.Border = true
	p2.Title = "File Information"
	p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(0)))
	p2.SetRect(0, cfm.GetCliHeight(), cfm.GetCliWidth(), int(float64(cfm.GetCliHeight())*0.73))
	//p2.SetRect(0, h, w, int(float64(h)*0.73))
	p2.BorderStyle.Fg = ui.ColorBlue
	p2.WrapText = false
	p2.TitleStyle.Modifier = ui.ModifierBold

	ui.Render(l, p, p2, p3)
}

func backToCfm(pText string, disk cfm.Status, selected int) {
	initItems(pText, disk)
	l.SelectedRow = selected
	catMode = false
}

func cat(path string, filename string) {
	filePath := fmt.Sprintf("%s/%s", path, filename)

	p2.Title = ""
	p2.Text = ""
	p2.Border = false

	p3.Text = ""
	p3.Title = ""
	p3.Border = false

	l.SelectedRow = 0
	l.Title = filename
	p.Text = "[Esc](fg:green) - Back  |  [↑](fg:green) - Scroll Up  |  [↓](fg:green) - Scroll Down"
	p.SetRect(0, 0, cfm.GetCliWidth(), 3)
	l.SetRect(0, 3, cfm.GetCliWidth(), int(float64(cfm.GetCliHeight())))

	p2.SetRect(cfm.GetCliWidth()+1, cfm.GetCliHeight()+1, cfm.GetCliWidth()+1, cfm.GetCliHeight()+1)
	p3.SetRect(cfm.GetCliWidth()+1, cfm.GetCliHeight()+1, cfm.GetCliWidth()+1, cfm.GetCliHeight()+1)

	file, err := os.Open(filePath)
	if err != nil {
		errorMsg("Unable to open this file.")
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	l.Rows = lines

	ui.Render(l, p, p2, p3)
}

func textFieldStyle() {
	l.SelectedRowStyle.Bg = ui.ColorWhite
	l.SelectedRowStyle.Fg = ui.ColorBlack
}

func resetColors() {
	l.SelectedRowStyle.Bg = ui.ColorClear
	l.SelectedRowStyle.Fg = ui.ColorBlue
}

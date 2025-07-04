// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/lrstanley/go-ytdlp"
)

const slowDownload = true

var core *tea.Program

var defaultDownloads = []string{
	"https://cdn.liam.sh/github/go-ytdlp/sample-1.mp4",
	"https://cdn.liam.sh/github/go-ytdlp/sample-2.mp4",
	"https://cdn.liam.sh/github/go-ytdlp/sample-3.mp4",
	"https://cdn.liam.sh/github/go-ytdlp/sample-4.mpg",
}

type model struct {
	urls     []string
	width    int
	height   int
	spinner  spinner.Model
	progress progress.Model
	done     bool

	lastProgress ytdlp.ProgressUpdate
}

var (
	fileStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("93"))
	doneStyle    = lipgloss.NewStyle().Margin(1, 2)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).SetString("✗")
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	etaStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sizeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func newModel() model {
	m := model{
		spinner: spinner.New(),
		progress: progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(40),
		),
	}

	if len(os.Args[1:]) > 0 {
		m.urls = os.Args[1:]
	} else {
		m.urls = defaultDownloads
	}

	for _, uri := range m.urls {
		_, err := url.Parse(uri)
		if err != nil {
			fmt.Printf("%s unvalid URL specified %q: %s\n", errorStyle, uri, err) //nolint:forbidigo
			os.Exit(1)
		}
	}

	m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	return m
}

type MsgToolsVerified struct {
	Resolved []*ytdlp.ResolvedInstall
	Error    error
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		// If yt-dlp/ffmpeg/ffprobe isn't installed yet, download and cache the binaries for further use.
		// Note that the download/installation of ffmpeg/ffprobe is only supported on a handful of platforms,
		// and so it is still recommended to install ffmpeg/ffprobe via other means.
		func() tea.Msg {
			resolved, err := ytdlp.InstallAll(context.TODO())
			if err != nil {
				return MsgToolsVerified{Resolved: resolved, Error: err}
			}
			return MsgToolsVerified{Resolved: resolved}
		},
		m.spinner.Tick,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case MsgToolsVerified:
		if msg.Error != nil {
			return m, tea.Sequence(
				tea.Printf("%s error installing/verifying tools: %s", errorStyle, msg.Error),
				tea.Quit,
			)
		}

		var cmds []tea.Cmd

		for _, r := range msg.Resolved {
			cmds = append(cmds, tea.Printf(
				"%s installed/verified tool %s (version: %s)",
				successStyle,
				r.Executable,
				r.Version,
			))
		}

		return m, tea.Sequence(append(cmds, m.initiateDownload)...)
	case MsgProgress:
		m.lastProgress = msg.Progress
		cmds := []tea.Cmd{m.progress.SetPercent(msg.Progress.Percent() / 100)}

		if m.lastProgress.Status == ytdlp.ProgressStatusFinished {
			cmds = append(cmds, tea.Printf(
				"%s downloaded %s (%s)",
				successStyle,
				fileStyle.Render(*m.lastProgress.Info.URL),
				titleStyle.Render(*m.lastProgress.Info.Filename),
			))
		}
		if m.lastProgress.Status == ytdlp.ProgressStatusError {
			cmds = append(cmds, tea.Printf(
				"%s error downloading: %s",
				errorStyle,
				*m.lastProgress.Info.URL,
			))
		}
		return m, tea.Sequence(cmds...)
	case MsgFinished:
		m.done = true
		if msg.Err != nil {
			return m, tea.Sequence(
				tea.Printf("%s error downloading urls: %s", errorStyle, msg.Err),
				tea.Quit,
			)
		}
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}
	return m, nil
}

type MsgProgress struct {
	Progress ytdlp.ProgressUpdate
}

type MsgFinished struct {
	Result *ytdlp.Result
	Err    error
}

func (m model) initiateDownload() tea.Msg {
	dl := ytdlp.New().
		FormatSort("res,ext:mp4:m4a").
		RecodeVideo("mp4").
		ForceOverwrites().
		ProgressFunc(100*time.Millisecond, func(prog ytdlp.ProgressUpdate) {
			core.Send(MsgProgress{Progress: prog})
		}).
		Output("%(extractor)s - %(title)s.%(ext)s")

	if slowDownload {
		dl = dl.LimitRate("2M")
	}

	result, err := dl.Run(context.TODO(), m.urls...)

	return MsgFinished{Result: result, Err: err}
}

func (m model) View() string {
	// " <spinner> <status> <file> <GAP> <progress> [eta: <eta>] [size: <size>]"

	if m.lastProgress.Status == "" {
		return doneStyle.Render(m.spinner.View() + " fetching url information...\n")
	}

	if m.done {
		return doneStyle.Render(fmt.Sprintf("downloaded %d urls.\n", len(m.urls)))
	}

	spin := m.spinner.View()
	status := string(m.lastProgress.Status)
	prog := m.progress.View()
	eta := m.lastProgress.ETA().Round(time.Second).String()
	eta = "[eta: " + etaStyle.MarginLeft(max(0, 4-len(eta))).Render(eta) + "]"
	size := "[size: " + sizeStyle.Render(humanize.Bytes(uint64(m.lastProgress.TotalBytes))) + "]"

	cellsAvail := max(0, m.width-lipgloss.Width(spin+" "+status+" "+prog+" "+eta+" "+size))

	file := fileStyle.MaxWidth(cellsAvail).Render(m.lastProgress.Filename)

	cellsRemaining := max(0, cellsAvail-lipgloss.Width(file))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + " " + status + " " + file + gap + prog + " " + eta + " " + size
}

func main() {
	core = tea.NewProgram(newModel())
	_, err := core.Run()
	if err != nil {
		fmt.Printf("error running program: %v\n", err) //nolint:forbidigo
		os.Exit(1)
	}
}

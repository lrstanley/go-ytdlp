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

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

var _ tea.Model = model{}

var (
	keyQuit = key.NewBinding(
		key.WithKeys("ctrl+c", "esc", "q"),
		key.WithHelp("q", "quit"),
	)

	fileStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("93"))
	doneStyle    = lipgloss.NewStyle().Margin(1, 2)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).SetString("✗")
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	etaStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sizeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
)

func newProgress() progress.Model {
	return progress.New(
		progress.WithDefaultBlend(),
		progress.WithWidth(40),
	)
}

func newModel() model {
	m := model{
		spinner:  spinner.New(spinner.WithStyle(spinnerStyle)),
		progress: newProgress(),
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
	case tea.KeyPressMsg:
		if key.Matches(msg, keyQuit) {
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
		next := msg.Progress
		if m.shouldResetProgress(next) {
			// Replace the widget. SetPercent(0) only changes the target, so the
			// spring would still render 100% on this frame and tween down.
			m.progress = newProgress()
		}
		m.lastProgress = next

		if next.Status == ytdlp.ProgressStatusFinished {
			return m, tea.Printf(
				"%s downloaded %s (%s)",
				successStyle,
				fileStyle.Render(*next.Info.URL),
				titleStyle.Render(*next.Info.Filename),
			)
		}
		if next.Status == ytdlp.ProgressStatusError {
			return m, tea.Printf(
				"%s error downloading: %s",
				errorStyle,
				*next.Info.URL,
			)
		}
		return m, m.progress.SetPercent(next.Percent() / 100)
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
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) shouldResetProgress(next ytdlp.ProgressUpdate) bool {
	prev := m.lastProgress
	if prev.Status == "" {
		return false
	}
	if next.Status.IsCompletedType() || prev.Status.IsCompletedType() {
		return true
	}
	if next.Status == ytdlp.ProgressStatusStarting {
		return true
	}
	if next.Percent()/100 < m.progress.Percent() || next.Percent() < prev.Percent() {
		return true
	}
	return downloadKey(prev) != downloadKey(next)
}

func downloadKey(p ytdlp.ProgressUpdate) string {
	id, src, dest := "", "", p.Filename
	if p.Info != nil {
		id = p.Info.ID
		if p.Info.URL != nil {
			src = *p.Info.URL
		}
		if p.Info.Filename != nil && *p.Info.Filename != "" {
			dest = *p.Info.Filename
		}
	}
	return id + "\x00" + src + "\x00" + dest + "\x00" + p.Filename
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

func (m model) View() tea.View {
	// " <spinner> <status> <file> <GAP> <progress> [eta: <eta>] [size: <size>]"

	if m.done {
		return tea.NewView(doneStyle.Render(fmt.Sprintf("downloaded %d urls.\n", len(m.urls))))
	}

	if m.lastProgress.Status == "" || m.lastProgress.Status.IsCompletedType() {
		return tea.NewView(m.spinner.View() + " fetching url information...")
	}

	var b strings.Builder

	spin := m.spinner.View()
	status := string(m.lastProgress.Status)
	prog := m.progress.View()
	etaRaw := m.lastProgress.ETA().Round(time.Second).String()
	eta := "[eta: " + etaStyle.MarginLeft(max(0, 4-lipgloss.Width(etaRaw))).Render(etaRaw) + "]"
	size := "[size: " + sizeStyle.Render(humanize.Bytes(uint64(m.lastProgress.TotalBytes))) + "]"

	cellsAvail := max(0, m.width-lipgloss.Width(spin+" "+status+" "+prog+" "+eta+" "+size))

	file := fileStyle.MaxWidth(cellsAvail).Render(m.lastProgress.Filename)

	cellsRemaining := max(0, cellsAvail-lipgloss.Width(file))

	b.WriteString(spin)
	b.WriteByte(' ')
	b.WriteString(status)
	b.WriteByte(' ')
	b.WriteString(file)
	b.WriteString(strings.Repeat(" ", cellsRemaining))
	b.WriteString(prog)
	b.WriteByte(' ')
	b.WriteString(eta)
	b.WriteByte(' ')
	b.WriteString(size)

	return tea.NewView(b.String())
}

func main() {
	core = tea.NewProgram(newModel())
	_, err := core.Run()
	if err != nil {
		fmt.Printf("error running program: %v\n", err) //nolint:forbidigo
		os.Exit(1)
	}
}

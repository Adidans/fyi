package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v4/cpu"
)

const baseWeatherUrl = "http://api.weatherapi.com/v1/"

type keyMap struct {
	Quit key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Quit,
	}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Quit,
		},
	}
}

type weatherResponse struct {
	Location struct {
		Name           string  `json:"name"`
		Region         string  `json:"region"`
		Country        string  `json:"country"`
		Lat            float64 `json:"lat"`
		Lon            float64 `json:"lon"`
		TzID           string  `json:"tz_id"`
		LocaltimeEpoch int     `json:"localtime_epoch"`
		Localtime      string  `json:"localtime"`
	} `json:"location"`
	Current struct {
		LastUpdatedEpoch int     `json:"last_updated_epoch"`
		LastUpdated      string  `json:"last_updated"`
		TempC            float64 `json:"temp_c"`
		TempF            float64 `json:"temp_f"`
		IsDay            int     `json:"is_day"`
		Condition        struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
		WindMph    float64 `json:"wind_mph"`
		WindKph    float64 `json:"wind_kph"`
		WindDegree int     `json:"wind_degree"`
		WindDir    string  `json:"wind_dir"`
		PressureMb float64 `json:"pressure_mb"`
		PressureIn float64 `json:"pressure_in"`
		PrecipMm   float64 `json:"precip_mm"`
		PrecipIn   float64 `json:"precip_in"`
		Humidity   int     `json:"humidity"`
		Cloud      int     `json:"cloud"`
		FeelslikeC float64 `json:"feelslike_c"`
		FeelslikeF float64 `json:"feelslike_f"`
		WindchillC float64 `json:"windchill_c"`
		WindchillF float64 `json:"windchill_f"`
		HeatindexC float64 `json:"heatindex_c"`
		HeatindexF float64 `json:"heatindex_f"`
		DewpointC  float64 `json:"dewpoint_c"`
		DewpointF  float64 `json:"dewpoint_f"`
		VisKm      float64 `json:"vis_km"`
		VisMiles   float64 `json:"vis_miles"`
		Uv         float64 `json:"uv"`
		GustMph    float64 `json:"gust_mph"`
		GustKph    float64 `json:"gust_kph"`
	} `json:"current"`
}

type model struct {
	weatherApiKey string
	weatherData   weatherResponse
	keys          keyMap
	help          help.Model
	currTime      time.Time
	cpuPercent    []float64
	width         int
	height        int
}

type secondMsg time.Time
type minuteMsg time.Time

func tickSystem() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return secondMsg(t)
	})
}

func tickWeather() tea.Cmd {
	return tea.Tick(time.Minute, func(t time.Time) tea.Msg {
		return minuteMsg(t)
	})
}

func initialModel() model {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %s", err)
		os.Exit(1)
	}
	weatherApiKey := os.Getenv("WEATHER_API_KEY")
	var weatherData weatherResponse
	url := fmt.Sprintf("%scurrent.json?key=%s&q=auto:ip", baseWeatherUrl, weatherApiKey)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching weather data: %s", err)
	}
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error fetching weather data: %s", err)
	}
	var data weatherResponse
	err = json.Unmarshal(responseData, &data)
	if err != nil {
		fmt.Printf("Error fetching weather data: %s", err)
	} else {
		weatherData = data
	}
	return model{
		weatherApiKey: weatherApiKey,
		weatherData:   weatherData,
		currTime:      time.Now(),
		cpuPercent:    []float64{},
		keys:          keys,
		help:          help.New(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickSystem(),
		tickWeather(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	case secondMsg:
		m.currTime = time.Now()
		percent, err := cpu.Percent(0, false)
		if err != nil {
			m.cpuPercent = []float64{0}
		}
		m.cpuPercent = percent
		return m, tickSystem()
	case minuteMsg:
		url := fmt.Sprintf("%scurrent.json?key=%s&q=auto:ip", baseWeatherUrl, m.weatherApiKey)
		resp, err := http.Get(url)
		if err != nil {
			m.weatherData = weatherResponse{}
		}
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			m.weatherData = weatherResponse{}
		}
		var data weatherResponse
		err = json.Unmarshal(responseData, &data)
		if err != nil {
			m.weatherData = weatherResponse{}
		} else {
			m.weatherData = data
		}
		return m, tickWeather()
	}
	return m, nil
}

func (m model) View() string {
	if len(m.cpuPercent) == 0 || m.weatherData.Location.Name == "" {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			"Loading...")
	}
	header := "FYI"
	help := m.help.View(m.keys)
	time := fmt.Sprintf("Current Time: %s\nCPU Usage: %.2f%%", m.currTime.Format(time.DateTime), m.cpuPercent[0])
	isDay := m.weatherData.Current.IsDay == 1
	weather := fmt.Sprintf("%s %.1f °C %s", m.weatherData.Location.Name, m.weatherData.Current.TempC, getWeatherIcon(m.weatherData.Current.Condition.Code, isDay))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, header, time, weather, help))
}

func getWeatherIcon(code int, isDay bool) string {
	if isDay {
		switch code {
		case 1000:
			return "☀️"
		case 1003:
			return "🌤️"
		case 1006:
			return "☁️"
		case 1009:
			return "☁️"
		case 1030:
			return "🌫️"
		case 1063, 1066, 1069, 1072:
			return "🌦️"
		case 1087:
			return "⛈️"
		case 1114, 1117:
			return "❄️"
		default:
			return "🌈"
		}
	} else {
		switch code {
		case 1000:
			return "🌙"
		case 1003:
			return "☁️"
		case 1006:
			return "☁️"
		case 1009:
			return "☁️"
		case 1030:
			return "🌫️"
		case 1063, 1066, 1069, 1072:
			return "🌦️"
		case 1087:
			return "⛈️"
		case 1114, 1117:
			return "❄️"
		default:
			return "🌌"
		}
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

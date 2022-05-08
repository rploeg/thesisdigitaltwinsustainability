package main

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"strings"

	"github.com/iot-for-all/iiot-oee/pkg/models"
	"github.com/iot-for-all/iiot-oee/pkg/simulating"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// handle process exit gracefully
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer func() {
		// Close the os signal channel to prevent any leak.
		signal.Stop(sig)
	}()

	// load configuration and initialize logger
	cfg, err := loadConfig()
	if err != nil {
		panic(fmt.Errorf("failed to initialize configuration. %w", err))
	}
	initLogger(cfg)

	// start devices
	for _, plant := range cfg.Plant {
		log.Debug().Str("plant", plant.Name).Msg("Starting up plant")
		for i := 1; i <= plant.BoltMachine.Count; i++ {
			log.Debug().Int("BoltMachine", i).Msg("Starting up bolt machine")
			deviceID := fmt.Sprintf("%s-BoltMachine-%d", plant.Name, i)
			boltMachine := models.BoltMachine{
				PlantName:          plant.Name,
				ProductionLine:     fmt.Sprintf("ProductionLine %d", i),
				ShiftNumber:        0,
				BatchNumber:        0,
				TotalPartsMade:     0,
				DefectivePartsMade: 0,
				MachineHealth:      "Healthy",
				OilLevel:           100,
				Temperature:        100,
				Kwh:				80,
				Format:             plant.BoltMachine.Format,
			}
			device := simulating.NewDevice(ctx, &cfg.Application, deviceID, &boltMachine)

			// start the device simulation of machines
			go device.Start()
		}
	}

	// Wait signal / cancellation
	<-sig

	cancel() // Wait for device to completely shut down.
}

// loadConfig loads the configuration file
func loadConfig() (*config, error) {
	colorReset := "\033[0m"
	//colorRed := "\033[31m"
	colorGreen := "\033[32m"
	//colorYellow := "\033[33m"
	colorBlue := "\033[34m"
	//colorPurple := "\033[35m"
	//colorCyan := "\033[36m"
	//colorWhite := "\033[37m"
	fmt.Printf(string(colorGreen))
	fmt.Printf(`
██╗██╗ ██████╗ ████████╗     ██████╗ ███████╗███████╗  ███████╗
██║██║██╔═══██╗╚══██╔══╝    ██╔═══██╗██╔════╝██╔════╝ ██╔════╝
██║██║██║   ██║   ██║       ██║   ██║█████╗  █████╗   █████╗ 
██║██║██║   ██║   ██║       ██║   ██║██╔══╝  ██╔══╝   ██╔══╝
██║██║╚██████╔╝   ██║       ╚██████╔╝███████╗███████╗ ███████╗
╚═╝╚═╝ ╚═════╝    ╚═╝        ╚═════╝ ╚══════╝╚══════╝ ══════╝
`)
	fmt.Printf(string(colorBlue))
	fmt.Printf(string(colorReset))

	viper.SetConfigName("iiotoee")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./bin")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Print(`Add a configuration file (iiotoee.json) with the file contents below:

{
  "logger": {
    "logLevel": "Debug",
    "logsDir": "./logs"
  },
  "application": {
    "provisioningUrl": "global.azure-devices-provisioning.net",
    "idScope": "CHANGE THIS -- YOUR_APP_IDSCOPE",
    "masterKey": "CHANGE THIS -- DPS Master key",
    "boltMachineModelID": "dtmi:parnellAerospace:BoltMakerV1;1"
  },
  "plant": [
    {
      "name": "Everett",
      "boltMachine":{
        "count": 2,
        "format": "json"
      }
    },
    {
      "name": "Austin",
      "boltMachine":{
        "count": 1,
        "format": "json"
      }
    },
    {
      "name": "Atlanta",
      "boltMachine":{
        "count": 1,
        "format": "json"
      }
    }
  ]
}

\n`)
			return nil, err
		}
	}

	cfg := newConfig()
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	//fmt.Printf("loaded configuration from %s\n", viper.ConfigFileUsed())
	return cfg, nil
}

// initLogger initializes the logger with output format
func initLogger(cfg *config) {
	var writers []io.Writer
	writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	fileLoggingEnabled := false
	if len(cfg.Logger.LogsDir) > 0 {
		fileLoggingEnabled = true
	}
	if fileLoggingEnabled {
		logsDir := cfg.Logger.LogsDir
		if err := os.MkdirAll(logsDir, 0744); err != nil {
			fmt.Printf("can't create log directory, so file logging is disabled, error: %s", err.Error())
		} else {
			fileWriter := &lumberjack.Logger{
				Filename:   path.Join(logsDir, "iiotoee.log"),
				MaxBackups: 3,  // files
				MaxSize:    10, // megabytes
				MaxAge:     30, // days
			}

			writers = append(writers, fileWriter)
			//fmt.Printf("file logging is enabled, logsDir: %s\n", logsDir)
		}
	}
	mw := io.MultiWriter(writers...)

	log.Logger = zerolog.New(mw).With().Timestamp().Logger()
	//log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	switch strings.ToLower(cfg.Logger.LogLevel) {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

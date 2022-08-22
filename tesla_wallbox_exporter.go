// MIT License
//
// Copyright (c) 2021-2022 Josef 'veloc1ty' Stautner
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vlcty/TeslaWallbox"
	"net/http"
	"os"
)

const (
	// Environment variable containing the wallbox IP address
	ENV_TESLA_WALLBOX_IP string = "TESLA_WALLBOX_IP"

	// Enable or disable debug output
	ENV_DEBUG string = "DEBUG"

	// Keep power meter stats when wallbox becomes unreachable
	ENV_KEEP_POWER_METER string = "KEEP_POWER_METER"
)

func getEnvVariableOrDie(lookup string) string {
	value, found := os.LookupEnv(lookup)

	if !found {
		log.Fatalf("Env variable %s not found", lookup)
	}

	return value
}

func main() {
	log.Info("Starting tesla_wallbox_exporter")

	ipAddress := getEnvVariableOrDie(ENV_TESLA_WALLBOX_IP)

	if getEnvVariableOrDie(ENV_DEBUG) == "true" {
		log.SetLevel(log.DebugLevel)
	}

	keepPowerMetersWhenUnreachable := false
	var lastDispensedEnergyValue uint = 0.0
	lastSessionEnergyValue := 0.0

	if value, found := os.LookupEnv(ENV_KEEP_POWER_METER); found && value == "true" {
		keepPowerMetersWhenUnreachable = true

		log.Info("Keeping last power meter stats when wallbox becomes unreachable")
	}

	log.Debugf("Looking for a Tesla walbox under %s", ipAddress)

	http.HandleFunc("/query", func(response http.ResponseWriter, request *http.Request) {
		vitals, vitalsError := teslaWallbox.FetchVitals(ipAddress)
		stats, statsError := teslaWallbox.FetchLifetimeStats(ipAddress)

		if vitalsError != nil {
			log.Error("Vitals error:", vitalsError)
		}

		if statsError != nil {
			log.Error("Stats error:", statsError)
		}

		response.WriteHeader(http.StatusOK)

		if keepPowerMetersWhenUnreachable {
			if vitals.SessionEnergy == 0 {
				vitals.SessionEnergy = lastSessionEnergyValue
			}

			if stats.DispensedEnergy == 0 {
				stats.DispensedEnergy = lastDispensedEnergyValue
			}

			lastSessionEnergyValue = vitals.SessionEnergy
			lastDispensedEnergyValue = stats.DispensedEnergy
		}

		// To whoever reads this: the teslaWallbox module returns initialized structs with default value
		// I actively want to write null/zero/default values to the timeseries database
		fmt.Fprintln(response, prometheusFormatWallboxVitals(vitals))
		fmt.Fprintln(response, prometheusFormatWallboxLifetimeStats(stats))

		log.Debugf("Vitals: %+v", vitals)
		log.Debugf("Stats: %+v", stats)
	})

	if e := http.ListenAndServe(":8420", nil); e != nil {
		log.Fatal(e)
	}
}

func prometheusFormatWallboxLifetimeStats(stats *teslaWallbox.LifetimeStats) string {
	formatString := `
# TYPE contactor_cycles counter
contactor_cycles %d

# TYPE contactor_cycles_loaded counter
contactor_cycles_loaded %d

# TYPE connector_cycles counter
connector_cycles %d

# TYPE thermal_foldbacks counter
thermal_foldbacks %d

# TYPE average_startup_temperature gauge
average_startup_temperature %.1f

# TYPE started_charging_sessions counter
started_charging_sessions %d

# TYPE dispensed_energy counter
dispensed_energy %d

# TYPE total_uptime counter
total_uptime %d

# TYPE total_charging_time counter
total_charging_time %d
`
	return fmt.Sprintf(formatString,
		stats.ContactorCycles,
		stats.ContactorCyclesLoaded,
		stats.ConnectorCycles,
		stats.ThermalFoldbacks,
		stats.AverageStartupTemperature,
		stats.StartedChargingSessions,
		stats.DispensedEnergy,
		stats.TotalUptime,
		stats.ChargingTime)
}

func prometheusFormatWallboxVitals(vitals *teslaWallbox.Vitals) string {
	formatString := `
# TYPE contactor_closed gauge
contactor_closed %d

# TYPE vehicle_connected gauge
vehicle_connected %d

# TYPE session_duration gauge
session_duration %d

# TYPE session_energy gauge
session_energy %.3f

# TYPE grid_voltage gauge
grid_voltage %.3f

# TYPE grid_frequency gauge
grid_frequency %.3f

# TYPE vehicle_current gauge
vehicle_current %.3f

# TYPE phase_a_current gauge
phase_a_current %.3f

# TYPE phase_b_current gauge
phase_b_current %.3f

# TYPE phase_c_current gauge
phase_c_current %.3f

# TYPE neutral_current gauge
neutral_current %.3f

# TYPE phase_a_voltage gauge
phase_a_voltage %.3f

# TYPE phase_b_voltage gauge
phase_b_voltage %.3f

# TYPE phase_c_voltage gauge
phase_c_voltage %.3f

# TYPE relay_coil_voltage gauge
relay_coil_voltage %.3f

# TYPE pcb_temperature gauge
pcb_temperature %.1f

# TYPE handle_temperature gauge
handle_temperature %.1f

# TYPE mcu_temperature gauge
mcu_temperature %.3f

# TYPE uptime gauge
uptime %d

# TYPE proximity_voltage gauge
proximity_voltage %.1f

# TYPE pilot_high_voltage gauge
pilot_high_voltage %.1f

# TYPE pilot_low_voltage gauge
pilot_low_voltage %.1f

# TYPE config_status gauge
config_status %d

# TYPE evse_state gauge
evse_state %d
    `

	return fmt.Sprintf(formatString,
		boolToInt(vitals.ContactorClosed),
		boolToInt(vitals.VehicleConnected),
		vitals.SessionDuration,
		vitals.SessionEnergy,
		vitals.GridVoltage,
		vitals.GridFrequency,
		vitals.VehicleCurrent,
		vitals.PhaseACurrent,
		vitals.PhaseBCurrent,
		vitals.PhaseCCurrent,
		vitals.NeutralCurrent,
		vitals.PhaseAVoltage,
		vitals.PhaseBVoltage,
		vitals.PhaseCVoltage,
		vitals.RelayCoilVoltage,
		vitals.PcbTemperature,
		vitals.HandleTemperature,
		vitals.McuTemperature,
		vitals.Uptime,
		vitals.ProximityVoltage,
		vitals.PilotHighVoltage,
		vitals.PilotLowVoltage,
		vitals.ConfigStatus,
		vitals.EvseState)
}

func boolToInt(v bool) int {
	if v {
		return 1
	}

	return 0
}

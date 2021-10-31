# tesla_wallbox_exporter

Prometheus like exporter/proxy for Tesla Gen 3 wallboxes.

## Usage

The application needs two environment variables:

- `TESLA_WALLBOX_IP`: The IP address of the wallbox. It seems that the webserver currently only listens on IPv4 (shame!)
- `DEBUG`: `true` for extra output, otherwise pass `false`

Both env vars MUST be set!

Then you can run the application. For example:

```
TESLA_WALLBOX_IP="10.10.0.39" DEBUG="true" go run tesla_wallbox_exporter.go
```

You can of course compile it to a binary and deploy it somewhere else.

## Limitations

The known limitations are:

- Nobody knows when or if Tesla revokes access to this "undocumented" API. Or changes it. Don't rely on this piece of software for mission critical tasks.
- Null/Empty/Default values are written when the wallbox can't be reached.
- This project depends on [vlcty/TeslaWallbox](https://github.com/vlcty/TeslaWallbox). Check the README to find out which firmware versions and power grid configurations are compatible.
- The reported grid voltage and grid frequency values seem to be a bit lower compared to the real life measured values. This application does not compensate for that.

## Sample output

You can query the result yourself:

```
curl localhost:8420/query
```

Adapt the URL according to your setup. Sample:

```
# TYPE contactor_closed gauge
contactor_closed 0

# TYPE vehicle_connected gauge
vehicle_connected 1

# TYPE session_duration gauge
session_duration 11012

# TYPE session_energy gauge
session_energy 21687.398

# TYPE grid_voltage gauge
grid_voltage 228.600

# TYPE grid_frequency gauge
grid_frequency 49.920

# TYPE vehicle_current gauge
vehicle_current 0.300

# TYPE phase_a_current gauge
phase_a_current 0.200

# TYPE phase_b_current gauge
phase_b_current 0.300

# TYPE phase_c_current gauge
phase_c_current 0.200

# TYPE neutral_current gauge
neutral_current 0.300

# TYPE phase_a_voltage gauge
phase_a_voltage 0.000

# TYPE phase_b_voltage gauge
phase_b_voltage 0.000

# TYPE phase_c_voltage gauge
phase_c_voltage 0.000

# TYPE relay_coil_voltage gauge
relay_coil_voltage 12.000

# TYPE pcb_temperature gauge
pcb_temperature 20.4

# TYPE handle_temperature gauge
handle_temperature 14.6

# TYPE mcu_temperature gauge
mcu_temperature 26.300

# TYPE uptime gauge
uptime 121159

# TYPE proximity_voltage gauge
proximity_voltage 0.0

# TYPE pilot_high_voltage gauge
pilot_high_voltage 8.7

# TYPE pilot_low_voltage gauge
pilot_low_voltage -11.9

# TYPE config_status gauge
config_status 5

# TYPE evse_state gauge
evse_state 4


# TYPE contactor_cycles counter
contactor_cycles 67

# TYPE contactor_cycles_loaded counter
contactor_cycles_loaded 0

# TYPE connector_cycles counter
connector_cycles 16

# TYPE thermal_foldbacks counter
thermal_foldbacks 0

# TYPE average_startup_temperature gauge
average_startup_temperature 27.9

# TYPE started_charging_sessions counter
started_charging_sessions 67

# TYPE dispensed_energy counter
dispensed_energy 333972

# TYPE total_uptime counter
total_uptime 602392

# TYPE total_charging_time counter
total_charging_time 133400
```

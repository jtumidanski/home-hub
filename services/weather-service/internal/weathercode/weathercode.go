package weathercode

type Info struct {
	Summary string
	Icon    string
}

var codes = map[int]Info{
	0:  {"Clear", "sun"},
	1:  {"Mostly Clear", "sun"},
	2:  {"Partly Cloudy", "cloud-sun"},
	3:  {"Overcast", "cloud"},
	45: {"Fog", "cloud-fog"},
	48: {"Fog", "cloud-fog"},
	51: {"Drizzle", "cloud-drizzle"},
	53: {"Drizzle", "cloud-drizzle"},
	55: {"Drizzle", "cloud-drizzle"},
	56: {"Freezing Drizzle", "cloud-drizzle"},
	57: {"Freezing Drizzle", "cloud-drizzle"},
	61: {"Rain", "cloud-rain"},
	63: {"Rain", "cloud-rain"},
	65: {"Rain", "cloud-rain"},
	66: {"Freezing Rain", "cloud-rain"},
	67: {"Freezing Rain", "cloud-rain"},
	71: {"Snow", "snowflake"},
	73: {"Snow", "snowflake"},
	75: {"Snow", "snowflake"},
	77: {"Snow Grains", "snowflake"},
	80: {"Rain Showers", "cloud-rain"},
	81: {"Rain Showers", "cloud-rain"},
	82: {"Rain Showers", "cloud-rain"},
	85: {"Snow Showers", "snowflake"},
	86: {"Snow Showers", "snowflake"},
	95: {"Thunderstorm", "cloud-lightning"},
	96: {"Thunderstorm with Hail", "cloud-lightning"},
	99: {"Thunderstorm with Hail", "cloud-lightning"},
}

func Lookup(code int) (summary string, icon string) {
	if info, ok := codes[code]; ok {
		return info.Summary, info.Icon
	}
	return "Unknown", "cloud"
}

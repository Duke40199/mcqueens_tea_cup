package idac_domain

import "sort"

type TrackVariant struct {
	Name string
	ID   string
}

// TrackRegistry maps a Track Name to its available Variants
var TrackRegistry = map[string][]TrackVariant{
	"Akina Lake": {
		{"Counter-Clockwise (CCW)", "course-0"},
		{"Clockwise (CW)", "course-2"},
	},
	"Hakone": {
		{"Downhill", "course-52"},
		{"Uphill", "course-54"},
	},
	"Usui": {
		{"Counter-Clockwise (CCW)", "course-36"},
		{"Clockwise (CW)", "course-38"},
	},
	"Myogi": {
		{"Downhill", "course-4"},
		{"Uphill", "course-6"},
	},
	"Akagi": {
		{"Downhill", "course-8"},
		{"Uphill", "course-10"},
	},
	"Akina": {
		{"Downhill", "course-12"},
		{"Uphill", "course-14"},
	},
	"Irohazaka": {
		{"Downhill", "course-16"},
		{"Reverse/Uphill", "course-18"},
	},
	"Tsukuba": {
		{"Outbound", "course-20"},
		{"Inbound", "course-22"},
	},
	"Happogahara": {
		{"Outbound", "course-24"},
		{"Inbound", "course-26"},
	},
	"Nagao": {
		{"Downhill", "course-28"},
		{"Uphill", "course-30"},
	},
	"Tsubaki Line": {
		{"Downhill", "course-32"},
		{"Uphill", "course-34"},
	},
	"Sadamine": {
		{"Downhill", "course-40"},
		{"Uphill", "course-42"},
	},
	"Tsuchisaka": {
		{"Outbound", "course-44"},
		{"Inbound", "course-46"},
	},
	"Akina Snow": {
		{"Downhill", "course-48"},
		{"Uphill", "course-50"},
	},
	"Momiji Line": {
		{"Downhill", "course-56"},
		{"Uphill", "course-58"},
	},
	"Nanamagari": {
		{"Downhill", "course-60"},
		{"Uphill", "course-62"},
	},
	"Gunsai": {
		{"Outbound", "course-64"},
		{"Inbound", "course-66"},
	},
	"Odawara": {
		{"Outbound", "course-68"},
		{"Inbound", "course-70"},
	},
	"Tsukuba Snow": {
		{"Outbound", "course-72"},
		{"Inbound", "course-74"},
	},
	"Yabitsu": {
		{"Downhill", "course-76"},
		{"Uphill", "course-78"},
	},
	"Tsuchisaka Snow": {
		{"Outbound", "course-80"},
		{"Inbound", "course-82"},
	},
	"Manazuru": {
		{"Outbound", "course-84"},
		{"Inbound", "course-86"},
	},
	"Usui Snow": {
		{"Counter-Clockwise (CCW)", "course-88"},
		{"Clockwise (CW)", "course-90"},
	},
	"Akina Rain": {
		{"Downhill", "course-92"},
		{"Uphill", "course-94"},
	},
}

// Helper to get keys for the Dropdown
func GetTrackNames() []string {
	keys := make([]string, 0, len(TrackRegistry))
	for k := range TrackRegistry {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Ensure consistent order
	return keys
}

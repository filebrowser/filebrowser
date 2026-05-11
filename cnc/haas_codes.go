package cnc

// Haas alarm / setting / parameter code lookups. Built from the Haas
// operator's manual (Next Generation Control + classic) so we can turn a
// bare number like "Setting 414" into a one-line explanation the
// operator can act on without flipping to the manual.
//
// Coverage is curated, not exhaustive — we include the codes operators
// reach the dashboard with (RS-232 / DNC / probing / tool-table /
// communication failures). Anything we don't recognize falls through to
// LookupCode returning ok=false; the UI surfaces the bare number rather
// than guessing.
//
// Sources:
//   - Haas Operator's Manual NGC (Settings, Alarms appendix)
//   - Haas Service Manual (Alarm codes / Diagnostic codes)
//   - Forum verifications for the codes we've actually seen in shop logs.

import (
	"sort"
	"strings"
)

// CodeKind tags whether a number is an alarm, a setting, or a parameter.
// The same integer (e.g. 414) can mean wildly different things in each
// space, so callers always carry the kind alongside.
type CodeKind string

const (
	CodeKindAlarm     CodeKind = "alarm"
	CodeKindSetting   CodeKind = "setting"
	CodeKindParameter CodeKind = "parameter"
)

// CodeEntry is one row in the curated lookup.
type CodeEntry struct {
	Kind     CodeKind `json:"kind"`
	Number   int      `json:"number"`
	Title    string   `json:"title"`
	Category string   `json:"category,omitempty"`
	Summary  string   `json:"summary"`
	// Hint is an actionable next step. Empty string is allowed.
	Hint string `json:"hint,omitempty"`
}

// Settings — focus on RS-232 / DNC / probing / network / tool changer.
var haasSettings = map[int]CodeEntry{
	1: {
		Kind: CodeKindSetting, Number: 1, Title: "Auto Power Off Timer",
		Category: "machine", Summary: "Minutes idle before the control auto-shuts down.",
	},
	7: {
		Kind: CodeKindSetting, Number: 7, Title: "Parameter Lock",
		Category: "machine", Summary: "Locks parameter pages. Set OFF to edit parameters.",
	},
	8: {
		Kind: CodeKindSetting, Number: 8, Title: "Program Memory Lock",
		Category: "programs", Summary: "Locks edits to programs stored in memory.",
	},
	11: {
		Kind: CodeKindSetting, Number: 11, Title: "Baud Rate Selection",
		Category: "communication", Summary: "RS-232 baud. 9600 (default) or 38400 are the supported choices on most Haas controls.",
		Hint: "Match the Pi side to this value. Mismatched baud rates manifest as garbled file content on receive.",
	},
	12: {
		Kind: CodeKindSetting, Number: 12, Title: "Parity Selection",
		Category: "communication", Summary: "Even / Odd / None. Default None on most installs.",
		Hint: "Pi side must match. Pi default in this project is 8N1.",
	},
	13: {
		Kind: CodeKindSetting, Number: 13, Title: "Stop Bits",
		Category: "communication", Summary: "1 or 2. Default 1.",
	},
	14: {
		Kind: CodeKindSetting, Number: 14, Title: "Synchronization (Handshake)",
		Category: "communication", Summary: "XON/XOFF or RTS/CTS. XON/XOFF is the Haas-typical default for DNC drip-feed.",
		Hint: "Drip-feed / DNC needs XON/XOFF; the Pi bridge honors it. Hardware (RTS/CTS) requires a fully-wired serial cable.",
	},
	23: {
		Kind: CodeKindSetting, Number: 23, Title: "9xxx Programs Edit Lock",
		Category: "programs", Summary: "Locks 9000-series macro programs (vendor utilities).",
	},
	28: {
		Kind: CodeKindSetting, Number: 28, Title: "Can Cycle Act w/o XYZ",
		Category: "machine", Summary: "Allow canned cycle to execute with no XYZ axis move.",
	},
	36: {
		Kind: CodeKindSetting, Number: 36, Title: "Program Restart",
		Category: "machine", Summary: "Scan program from current cursor back to find modal state when restarting mid-program.",
	},
	37: {
		Kind: CodeKindSetting, Number: 37, Title: "RS-232 Data Bits",
		Category: "communication", Summary: "7 or 8 data bits. Pi side runs 8.",
	},
	39: {
		Kind: CodeKindSetting, Number: 39, Title: "Beep at M00, M01, M02, M30",
		Category: "ui", Summary: "Audible cue at program stop / end.",
	},
	52: {
		Kind: CodeKindSetting, Number: 52, Title: "G83 Retract Above R Plane",
		Category: "drilling", Summary: "Extra retract distance on peck-drill cycles.",
	},
	57: {
		Kind: CodeKindSetting, Number: 57, Title: "Exact Stop Canned X-Y",
		Category: "machine", Summary: "Force exact-stop between blocks inside canned cycles.",
	},
	77: {
		Kind: CodeKindSetting, Number: 77, Title: "Scale Integer F",
		Category: "machine", Summary: "F-value scaling: 1.0 vs 0.0001 in/min.",
	},
	130: {
		Kind: CodeKindSetting, Number: 130, Title: "Tap Retract Speed",
		Category: "tapping", Summary: "Multiplier applied to tap-out feedrate.",
	},
	142: {
		Kind: CodeKindSetting, Number: 142, Title: "Offset Change Tolerance",
		Category: "offsets", Summary: "Warn when an operator-entered tool / work offset jumps by more than this amount.",
	},
	143: {
		Kind: CodeKindSetting, Number: 143, Title: "Machine Data Collect",
		Category: "communication", Summary: "Master switch for Q-code / macro data collection over RS-232.",
		Hint: "MUST be ON. With this OFF, Q201 / Q500 / Q600 etc. are silently no-ops — bridge timeouts / blank metrics typically trace back here.",
	},
	158: {
		Kind: CodeKindSetting, Number: 158, Title: "Thermal Compensation %",
		Category: "machine", Summary: "Scales thermal-comp axis correction.",
	},
	185: {
		Kind: CodeKindSetting, Number: 185, Title: "Default G51 Scale Factor",
		Category: "machine", Summary: "Default scale when G51 fires without P.",
	},
	187: {
		Kind: CodeKindSetting, Number: 187, Title: "Serial Echo",
		Category: "communication", Summary: "When ON, the Haas echoes Q-code requests before returning the framed response.",
		Hint: "Bridge parser strips the echo either way; leave as-is unless troubleshooting raw frames.",
	},
	200: {
		Kind: CodeKindSetting, Number: 200, Title: "Jog Lock - Z Axis",
		Category: "ui", Summary: "Disables Z jog from the wheel until cleared.",
	},
	216: {
		Kind: CodeKindSetting, Number: 216, Title: "Servo and Hydraulic Shutoff",
		Category: "machine", Summary: "Minutes idle before servos power down.",
	},
	232: {
		Kind: CodeKindSetting, Number: 232, Title: "Default G51 Center Z",
		Category: "machine", Summary: "Default Z center when G51 omits Z.",
	},
	236: {
		Kind: CodeKindSetting, Number: 236, Title: "Floppy Disk Compatibility",
		Category: "communication", Summary: "Legacy. Leave OFF on machines using RS-232 or Ethernet drip.",
	},
	250: {
		Kind: CodeKindSetting, Number: 250, Title: "Mirror Image X Axis",
		Category: "machine", Summary: "Mirror about X.",
	},
	251: {
		Kind: CodeKindSetting, Number: 251, Title: "Mirror Image Y Axis",
		Category: "machine", Summary: "Mirror about Y.",
	},
	252: {
		Kind: CodeKindSetting, Number: 252, Title: "Mirror Image Z Axis",
		Category: "machine", Summary: "Mirror about Z.",
	},
	276: {
		Kind: CodeKindSetting, Number: 276, Title: "Workholding Clamp Input",
		Category: "machine", Summary: "Input number used to confirm clamp state before cycle start.",
	},
	316: {
		Kind: CodeKindSetting, Number: 316, Title: "Tool Carousel Direction",
		Category: "tool-changer", Summary: "Shortest-path vs unidirectional tool-carousel rotation.",
	},
	330: {
		Kind: CodeKindSetting, Number: 330, Title: "Multi Boot Selection Time-out",
		Category: "ui", Summary: "Seconds the operator has to pick a boot config at startup.",
	},
	340: {
		Kind: CodeKindSetting, Number: 340, Title: "Front Door Hold Mode",
		Category: "machine", Summary: "Feed-hold response when the front door opens mid-cycle.",
	},
	383: {
		Kind: CodeKindSetting, Number: 383, Title: "Disable Multi-page Programs",
		Category: "programs", Summary: "Forces single-program-per-file view in the editor.",
	},
	408: {
		Kind: CodeKindSetting, Number: 408, Title: "Disable Software Limits",
		Category: "machine", Summary: "Service-only. Should be OFF in normal operation.",
	},
	409: {
		Kind: CodeKindSetting, Number: 409, Title: "Default Coolant Pressure",
		Category: "coolant", Summary: "Default P-value for M88 (TSC) when omitted.",
	},
	414: {
		Kind: CodeKindSetting, Number: 414, Title: "Probe Calibration Cycles",
		Category: "probing", Summary: "Number of repeat strokes used when calibrating a spindle or work probe (Setting 414).",
		Hint: "If a tool-length probe macro raises an alarm pointing to Setting 414, the probe-cycle count parameter is either zero or out of range. Set to 1 (or the probe vendor's value) before re-running the macro.",
	},
	445: {
		Kind: CodeKindSetting, Number: 445, Title: "Mist Collector Manual ON Time",
		Category: "machine", Summary: "Minutes the mist-collector relay stays energized when toggled manually.",
	},
	900: {
		Kind: CodeKindSetting, Number: 900, Title: "CNC Network Name",
		Category: "network", Summary: "Hostname used by NetShare / NGC networking.",
	},
	901: {
		Kind: CodeKindSetting, Number: 901, Title: "Obtain Address Automatically",
		Category: "network", Summary: "DHCP vs static IP on the Haas Ethernet port.",
	},
	902: {
		Kind: CodeKindSetting, Number: 902, Title: "IP Address",
		Category: "network", Summary: "Static IP when 901 is OFF.",
	},
	903: {
		Kind: CodeKindSetting, Number: 903, Title: "Subnet Mask",
		Category: "network", Summary: "Network mask used with the static IP.",
	},
	904: {
		Kind: CodeKindSetting, Number: 904, Title: "Gateway",
		Category: "network", Summary: "Default route.",
	},
	907: {
		Kind: CodeKindSetting, Number: 907, Title: "Domain / Workgroup Name",
		Category: "network", Summary: "SMB workgroup the Haas advertises into.",
	},
}

// Alarms — communication, probing, tool changer, the classic safety stops.
var haasAlarms = map[int]CodeEntry{
	102: {
		Kind: CodeKindAlarm, Number: 102, Title: "Servos Off",
		Category: "machine", Summary: "Servos disabled — usually after E-stop release without RESET.",
		Hint: "Press RESET, then POWER UP/RESTART.",
	},
	120: {
		Kind: CodeKindAlarm, Number: 120, Title: "Spindle Drive Fault",
		Category: "spindle", Summary: "Spindle inverter raised a fault.",
		Hint: "Cycle power. Recurring faults need a Haas service ticket — pull the drive log.",
	},
	121: {
		Kind: CodeKindAlarm, Number: 121, Title: "Low Air Pressure",
		Category: "pneumatics", Summary: "Compressed air below the regulator threshold.",
		Hint: "Verify shop air > 85 psi; check the inline regulator and oiler.",
	},
	125: {
		Kind: CodeKindAlarm, Number: 125, Title: "Low Coolant Level",
		Category: "coolant", Summary: "Coolant level switch tripped.",
	},
	133: {
		Kind: CodeKindAlarm, Number: 133, Title: "Spindle Not Locked",
		Category: "tool-changer", Summary: "Tool clamp didn't confirm — tool change blocked.",
		Hint: "Cycle the drawbar (POWER UP / unclamp+clamp). If it persists, drawbar wear is the usual culprit.",
	},
	139: {
		Kind: CodeKindAlarm, Number: 139, Title: "Tool # Disagreement",
		Category: "tool-changer", Summary: "Tool changer carousel and spindle disagree on which tool is in cut.",
		Hint: "Manually verify the spindle tool number matches the carousel position, then run TOOL CHANGER RECOVER.",
	},
	150: {
		Kind: CodeKindAlarm, Number: 150, Title: "Receiving Error From Tape Reader",
		Category: "communication", Summary: "RS-232 receive failed mid-frame.",
		Hint: "Check baud / parity / stop bits match between Haas (Settings 11/12/13/37) and the Pi. Confirm Setting 14 = XON/XOFF.",
	},
	151: {
		Kind: CodeKindAlarm, Number: 151, Title: "Send Buffer Overflow",
		Category: "communication", Summary: "Haas couldn't keep up with incoming DNC data.",
		Hint: "Flow control (XON/XOFF) misconfigured or cable wired without RX/TX/GND clean.",
	},
	152: {
		Kind: CodeKindAlarm, Number: 152, Title: "Bad Number Format",
		Category: "communication", Summary: "Haas saw a non-numeric token where it expected a number during RS-232 receive.",
		Hint: "Run dos2unix on the file; check for stray Windows BOM or 8-bit characters.",
	},
	153: {
		Kind: CodeKindAlarm, Number: 153, Title: "Programmed Stop",
		Category: "programs", Summary: "M00 reached.",
	},
	154: {
		Kind: CodeKindAlarm, Number: 154, Title: "Unrecognized Code",
		Category: "programs", Summary: "Block contains an unknown G/M/word.",
		Hint: "Common after vendor-specific post output — search the offending block for non-Haas G-codes.",
	},
	164: {
		Kind: CodeKindAlarm, Number: 164, Title: "Block End Of Memory",
		Category: "programs", Summary: "Program ended without M30/M99/%; receive buffer flushed.",
	},
	2010: {
		Kind: CodeKindAlarm, Number: 2010, Title: "Probe Cycle Error",
		Category: "probing", Summary: "Probe macro failed before reaching a touch.",
		Hint: "Re-seat probe, check Setting 414 calibration cycles, verify probe input on diagnostics page.",
	},
	2027: {
		Kind: CodeKindAlarm, Number: 2027, Title: "Probe Tool Length Macro Error",
		Category: "probing", Summary: "Length probe didn't trigger within the expected travel window.",
		Hint: "Confirm the spindle has the probe loaded; verify Settings 414 (cycles) and 60 (T probe Z).",
	},
}

// Parameters — much narrower set; we mostly use these to translate
// macro variable references back to human terms (#5021 / #5041 etc).
// The aggregator already speaks these natively but the codes panel is
// a useful operator-side cheatsheet.
var haasParameters = map[int]CodeEntry{
	2201: {
		Kind: CodeKindParameter, Number: 2201, Title: "Spindle Encoder Counts/Rev",
		Category: "spindle", Summary: "Service parameter — encoder resolution.",
	},
}

// LookupCode resolves (kind, number) → entry. Returns ok=false on miss
// rather than synthesizing a placeholder; callers render the bare
// number when ok is false.
func LookupCode(kind CodeKind, number int) (CodeEntry, bool) {
	switch kind {
	case CodeKindSetting:
		e, ok := haasSettings[number]
		return e, ok
	case CodeKindAlarm:
		e, ok := haasAlarms[number]
		return e, ok
	case CodeKindParameter:
		e, ok := haasParameters[number]
		return e, ok
	}
	return CodeEntry{}, false
}

// SearchCodes returns up to `limit` entries whose Title / Summary
// contain `q` (case-insensitive). kindFilter may be empty (all kinds)
// or one of the three known kinds. Results sort by kind, then number.
func SearchCodes(kindFilter CodeKind, q string, limit int) []CodeEntry {
	if limit <= 0 {
		limit = 50
	}
	q = strings.ToLower(strings.TrimSpace(q))
	out := make([]CodeEntry, 0, limit)
	consider := func(e CodeEntry) {
		if kindFilter != "" && e.Kind != kindFilter {
			return
		}
		if q != "" {
			hay := strings.ToLower(e.Title + " " + e.Summary + " " + e.Hint + " " + e.Category)
			if !strings.Contains(hay, q) {
				return
			}
		}
		out = append(out, e)
	}
	for _, e := range haasSettings {
		consider(e)
	}
	for _, e := range haasAlarms {
		consider(e)
	}
	for _, e := range haasParameters {
		consider(e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Number < out[j].Number
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

// NormalizeKind maps a wire-string kind to a known CodeKind, falling
// back to setting (the most operator-relevant default).
func NormalizeKind(s string) CodeKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(CodeKindAlarm):
		return CodeKindAlarm
	case string(CodeKindParameter):
		return CodeKindParameter
	default:
		return CodeKindSetting
	}
}

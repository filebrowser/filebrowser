package settings

// RuntimeDefaults defines default values for runtime parameters
var RuntimeDefaults = map[string]string{
	"root":    ".",
	"baseurl": "",
	"address": "127.0.0.1",
	"port":    "8080",
	"cert":    "",
	"key":     "",
	"log":     "stdout",
}

// RuntimeCfg contains parameters to be used at runtime
var RuntimeCfg = map[string]string{}

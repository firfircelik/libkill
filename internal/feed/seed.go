package feed

import (
	_ "embed"
)

//go:embed seed.json
var seedData []byte

//go:embed seed_malware.json
var seedMalwareData []byte

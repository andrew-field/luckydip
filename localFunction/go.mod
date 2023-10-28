module github.com/andrew-field/pickmypostcode/localFunction

go 1.21.1

replace github.com/andrew-field/pickmypostcode/cloudFunction => ../cloudFunction

require github.com/andrew-field/pickmypostcode/cloudFunction v0.0.0-00010101000000-000000000000

require (
	github.com/go-rod/rod v0.114.4 // indirect
	github.com/go-rod/stealth v0.4.9 // indirect
	github.com/ysmood/fetchup v0.2.3 // indirect
	github.com/ysmood/goob v0.4.0 // indirect
	github.com/ysmood/got v0.34.1 // indirect
	github.com/ysmood/gson v0.7.3 // indirect
	github.com/ysmood/leakless v0.8.0 // indirect
)

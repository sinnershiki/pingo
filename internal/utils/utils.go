package utils

import "github.com/sinnershiki/pingo/pkg/ping"

func Remove(slice []ping.PingResult, i int) []ping.PingResult {
	return slice[:i+copy(slice[i:], slice[i+1:])]
}

package service

import "omepic/backend/internal/iputil"

func ipHash(ipAddress string) string {
	return iputil.Hash(ipAddress)
}

func maskIPAddress(ipAddress string) string {
	return iputil.Mask(ipAddress)
}

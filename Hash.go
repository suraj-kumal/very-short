package main

const mixMask = (1 << 32) - 1

func obfuscate(id int, mixMultiplier int) int {
	return (id * mixMultiplier) & mixMask
}

func EncodeBase62(id int, mixMultiplier int) string {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	obfuscated := obfuscate(id, mixMultiplier)
	if obfuscated == 0 {
		return string(chars[0])
	}
	var result []byte
	for obfuscated > 0 {
		result = append(result, chars[obfuscated%62])
		obfuscated /= 62
	}
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

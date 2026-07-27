package main

func EncodeBase62(id int) string{
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	if id  == 0{
		return string(chars[0])
	}

	var result []byte

	for id > 0 {
        result = append([]byte{chars[id%62]}, result...)
        id /= 62
    }
    return string(result)
}

package model

import (
    "math/rand"
    "strings"
    "time"
)

var src = rand.NewSource(time.Now().UnixNano())
const (
    uidLength = 16 // DONT CHANGE THIS!!!!!!!!!!!!!!
    letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    letterIdxBits = 6                    // 6 bits to represent a letter index
    letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
    letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func generateUID() string {
    sb := strings.Builder{}
    sb.Grow(uidLength)
    // A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
    for i, cache, remain := uidLength-1, src.Int63(), letterIdxMax; i >= 0; {
        if remain == 0 {
            cache, remain = src.Int63(), letterIdxMax
        }
        if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
            sb.WriteByte(letterBytes[idx])
            i--
        }
        cache >>= letterIdxBits
        remain--
    }

    return sb.String()
}


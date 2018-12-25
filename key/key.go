package key

import (
	"crypto/sha1"
	"fmt"
	"time"
)

const (
	ServiceTypeHttp = 0
	ServiceTypeTcp  = 1
	ServiceTypeIcmp = 2
)

func MsFromDuration(d time.Duration) string {
	return fmt.Sprintf("%.2f", float64(d.Nanoseconds())/1000000.0)
}

// generate unique id for each request from check id and timestamp
// result is md5 hash
func GenerateReqId(id int) string {

	// initialize SHA encoder
	h := sha1.New()
	// convert check id into string value
	idString := fmt.Sprintf("%d", id)
	// get current unix timestamp and conver to string
	timestampString := fmt.Sprintf("%d", time.Now().Unix())
	// write string into SHA encoder
	h.Write([]byte(idString + timestampString))

	// output md5 sum of the hashed string
	md5Sum := fmt.Sprintf("%x", h.Sum(nil))

	return md5Sum
}

// 唯一ID生成器
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/12
package grapeConn

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var nseeduuid uint64 = 0

// uuId生成器
func CreateUUID(seed int) string {
	nseeduuid = atomic.AddUint64(&nseeduuid, 1)

	vmd5 := md5.New()
	vmd5.Write([]byte(strconv.FormatInt(int64(seed), 10)))
	vmd5.Write([]byte("_"))
	vmd5.Write([]byte(strconv.FormatUint(nseeduuid, 36)))
	vmd5.Write([]byte("_"))
	vmd5.Write([]byte(time.Now().Format("2006-01-02 15:04:05")))

	return strings.ToUpper(hex.EncodeToString(vmd5.Sum(nil)))
}

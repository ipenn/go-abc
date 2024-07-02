package abc

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/text/encoding/charmap"
	//gzip "github.com/klauspost/pgzip"
)

func PtoString(res interface{}, key string) string {
	r := res.(map[string]interface{})
	return *r[key].(*string)
}

//func PtoSqlNilStr(res interface{}, key string) sql.NullString {
//
//	r := res.(map[string]interface{})
//	return *r[key].(*sql.NullString)
//}

func HandleRawSQL(RawSQL *sql.Rows) []interface{} {
	columns, err := RawSQL.Columns()
	if err != nil {
		panic("HandleRawSQL1 " + err.Error())
	}
	length := len(columns)
	result := make([]interface{}, 0)
	for RawSQL.Next() {
		current := makeReceiver(length)
		err := RawSQL.Scan(current...)
		if err != nil {
			panic("HandleRawSQL2 " + err.Error())
		}

		value := make(map[string]interface{}, 0)
		for i := 0; i < length; i++ {
			key := columns[i]
			v := current[i].(*sql.NullString)
			val := &v.String
			value[key] = val
		}
		result = append(result, value)
	}
	return result
}

func makeReceiver(length int) []interface{} {
	make := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		var item sql.NullString
		make = append(make, &item)
	}
	return make
}

func SqlOperator(sql string, args ...interface{}) (interface{}, error) {
	rows, err := db.Debug().Raw(sql, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	d := HandleRawSQL(rows)
	if len(d) == 0 {
		return nil, nil
	}
	return d[0], nil
}

func SqlOperators(sql string, args ...interface{}) ([]interface{}, error) {
	rows, err := db.Debug().Raw(sql, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	d := HandleRawSQL(rows)
	if len(d) == 0 {
		return nil, nil
	}
	return d, nil
}

func TxSqlOperator(tx *gorm.DB, sql string, args ...interface{}) (interface{}, error) {
	rows, err := tx.Debug().Raw(sql, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	d := HandleRawSQL(rows)
	if len(d) == 0 {
		return nil, nil
	}
	return d[0], nil
}

func TxSqlOperators(tx *gorm.DB, sql string, args ...interface{}) ([]interface{}, error) {
	rows, err := tx.Debug().Raw(sql, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	d := HandleRawSQL(rows)
	if len(d) == 0 {
		return nil, nil
	}
	return d, nil
}

func DoRequest(method, urls string, payload io.Reader, header map[string][]string) []byte {
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic recovered", r)
		}
	}()
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(method, urls, payload)
	if err != nil {
		panic("DoRequest 初始化网络失败 " + err.Error())
	}

	//req.Header.Set("Content-Type", "application/json")
	req.Header = header

	res, err := client.Do(req)
	if err != nil {
		panic("DoRequest network connection failed " + err.Error())
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic("DoRequest EOF " + err.Error())
	}

	return body
}

func ToInt(t interface{}) int {
	switch t := t.(type) {
	case int:
		return t
	case string:
		tt, _ := strconv.Atoi(t)
		return tt
	case float32:
		return int(t)
	case float64:
		return int(t)
	case interface{}:
		return t.(int)
	}

	return 0
}

func ToInt64(t interface{}) int64 {
	switch t := t.(type) {
	case int:
		return int64(t)
	case string:
		tt, _ := strconv.ParseInt(t, 10, 64)
		return tt
	case float32:
		return int64(t)
	case float64:
		return int64(t)
	case interface{}:
		return t.(int64)
	}

	return 0
}

func ToString(t interface{}) string {
	switch t := t.(type) {
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case string:
		return t
	case float32:
		return strconv.FormatFloat(float64(t), 'f', 4, 64)
	case float64:
		return strconv.FormatFloat(t, 'f', 4, 64)
	case interface{}:
		return t.(string)
	}

	return ""
}

func ToFloat64(t interface{}) float64 {
	switch t := t.(type) {
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		tt, _ := strconv.ParseFloat(t, 64)
		return tt
	case float64:
		return t
	case interface{}:
		return t.(float64)
	}

	return 0.00
}

func Md5(pwd string) string {
	h := md5.New()
	h.Write([]byte(pwd))
	return hex.EncodeToString(h.Sum(nil))
}

func ToJSON(s interface{}) []byte {
	r, err := json.Marshal(s)
	if err != nil {
		panic("ToJSON " + err.Error())
	}
	return r
}

func ParseJSON(datas []byte, s interface{}) interface{} {
	json.Unmarshal(datas, s)
	return s
}

func GZipCompress(data []byte) []byte {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		gz.Close()
		panic("GZipCompress1 " + err.Error())
	}
	//if err := gz.Flush(); err != nil {
	//	gz.Close()
	//	panic("abc GZipCompress2" + err.Error())
	//}

	gz.Close()
	return b.Bytes()
}

func GZipUncompress(raw []byte) string {
	var out bytes.Buffer
	r, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		panic("GZipUncompress " + err.Error())
	}
	defer r.Close()
	io.Copy(&out, r)
	return out.String()
}

func Utf8To_ISO8859_1(iso8859_1 []byte) []byte {
	var buf bytes.Buffer
	w := charmap.ISO8859_1.NewEncoder().Writer(&buf)
	r := bytes.NewReader(iso8859_1)
	//if err != nil {
	//	log.Println("abc Uft8To_ISO8859_1", err)
	//}
	io.Copy(w, r)

	return buf.Bytes()
}

func ISO8859_1To_Utf8(input []byte) string {
	var buf bytes.Buffer
	r := charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(input))
	//if err != nil {
	//	log.Println("abc ISO8859_1To_Utf8", err)
	//}
	io.Copy(&buf, r)
	return buf.String()
}

func RandStr(n int) string {
	rand.Seed(time.Now().UnixNano())
	letterRunes := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func PicCompress(pic os.FileInfo, picName string) (string, error) {
	f, _ := os.Open(picName)
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		return format, err
	}
	switch format {
	case "jpeg":
		nf, _ := os.OpenFile(picName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		defer nf.Close()
		jpeg.Encode(nf, img, &jpeg.Options{Quality: 30})
	case "png":
		newImg := image.NewRGBA(img.Bounds())
		draw.Draw(newImg, newImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(newImg, newImg.Bounds(), img, img.Bounds().Min, draw.Over)
		nf, _ := os.OpenFile(picName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		defer nf.Close()
		jpeg.Encode(nf, newImg, &jpeg.Options{Quality: 30})
	case "gif":
		if pic.Size() < 204800 {
			return format, nil
		}
		f.Seek(0, 0)
		gifImg, _ := gif.DecodeAll(f)
		outGif := &gif.GIF{}
		for _, g := range gifImg.Image {
			overImage := image.NewRGBA(g.Rect)
			draw.Draw(overImage, overImage.Bounds(), g, image.Point{}, draw.Src)
			draw.Draw(overImage, overImage.Bounds(), g, image.Point{}, draw.Over)
			palettedImage := image.NewPaletted(g.Rect, palette.WebSafe)
			draw.Draw(palettedImage, palettedImage.Rect, overImage, image.Point{}, draw.Src)
			draw.Draw(palettedImage, palettedImage.Rect, overImage, image.Point{}, draw.Over)
			outGif.Image = append(outGif.Image, palettedImage)
			outGif.Delay = append(outGif.Delay, 0)
		}
		nf, _ := os.OpenFile(picName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		defer nf.Close()
		gif.EncodeAll(nf, outGif)
	default:
	}

	return format, nil
}

func RandonNumber(num int) string {
	letters := []byte("0123456789")
	randDigits := make([]byte, num)
	for i := 0; i < num; i++ {
		randDigits[i] = letters[rand.Intn(len(letters))]
	}
	return string(randDigits)
}

func FetchNetFile(dir, url string) string {
	times := strconv.FormatInt(time.Now().Unix(), 10)
	random := ToString(RandonNumber(6))

	res, _ := http.Get(url)
	defer res.Body.Close()

	filename := times + random

	file, err := os.Create(dir + "/" + filename)
	if err != nil {
		panic("FetchNetFile" + err.Error())
	}
	defer file.Close()

	io.Copy(file, res.Body)

	return filename
}

func FormatNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func UnixTimeToStr(t int64) string {
	return time.Unix(t, 0).Format("2006-01-02 15:04:05")
}

func GetTimer(date string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", date, time.Local)
	return t
}

func GetTimerDate(date string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02", date, time.Local)
	return t
}

func StringToUnix(strTime string) int64 {
	var t int64
	if strTime != "" {
		loc, _ := time.LoadLocation("Local") //获取当地时区
		location, _ := time.ParseInLocation("2006-01-02 15:04:05", strTime, loc)
		t = location.Unix()
	}
	return t
}

func ToAddDay(number int) string {
	nowTime := time.Now()
	var getTime time.Time
	getTime = nowTime.AddDate(0, 0, number)
	return getTime.Format("2006-01-02 15:04:05")
}

func HmacSha256(message string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func AesCBCEncrypt(key string, iv, plaintext []byte) ([]byte, []byte) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic("AesCBCEncrypt " + err.Error())
	}
	plaintextPadding := PKCS5Padding(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(plaintextPadding))
	if len(iv) == 0 {
		iv = []byte(RandStr(16))
	}
	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(ciphertext, plaintextPadding)

	return ciphertext, iv
}

func AesCBCDecrypt(key string, iv, ciphertext []byte) []byte {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic("AesCBCDecrypt " + err.Error())
	}
	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(ciphertext, ciphertext)
	return PKCS7Unpad(ciphertext, aes.BlockSize)
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7Padding(b []byte, blockSize int) ([]byte, error) {
	n := blockSize - (len(b) % blockSize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

func PKCS7Unpad(b []byte, blockSize int) []byte {
	if len(b) == 0 {
		panic("PKCS7Unpad invalid PKCS data")
	}
	if len(b)%blockSize != 0 {
		panic("PKCS7Unpad invalid padding")
	}
	c := b[len(b)-1]
	n := int(c)
	if n == 0 || n > len(b) {
		panic("PKCS7Unpad invalid PKCS length")
	}

	for i := 0; i < n; i++ {
		if b[len(b)-n+i] != c {
			panic("PKCS7Unpad invalid PKCS element")
		}
	}

	return b[:len(b)-n]
}

func ReceiveFrom(body, data interface{}) {
	switch body := body.(type) {
	case string:
		ParseJSON([]byte(body), data)
	default:
		mapstructure.Decode(body, data)
	}
}

func HandleHTMLEscape(t interface{}) interface{} {
	data := JSONMarshal(t)
	return ParseJSON(data, t)
}

func JSONMarshal(t interface{}) []byte {
	var out bytes.Buffer
	n := json.NewEncoder(&out)
	n.SetEscapeHTML(false)
	err := n.Encode(t)
	if err != nil {
		panic("JSONMarshal" + err.Error())
	}
	return out.Bytes()
}

func ForLimiter24hour() int64 {
	return time.Date(time.Now().Year(),
		time.Now().Month(),
		time.Now().Day()+1,
		0, 0, 0, 0,
		time.Local).Unix()
}

func ForLimiterSecond() int64 {
	return time.Now().Add(3 * time.Second).Unix()
}

func IsDatetimeIllegal(format, input string) bool {
	_, err := time.ParseInLocation(format, input, time.Local)
	return err != nil
}

func SwitchLanguage(language string) (lang string) {
	switch language {
	case "CN":
		lang = "zh"
	case "TC":
		lang = "hk"
	default:
		lang = "en"
	}
	return lang
}

func Reverse[T StrdNumd](list []T) map[T]bool {
	r := make(map[T]bool, len(list))
	for _, item := range list {
		r[item] = true
	}
	return r
}

func GetDaysBetweenDate(format, start, end string) (int64, error) {
	// 将字符串转化为Time格式
	date1, err := time.ParseInLocation(format, start, time.Local)
	if err != nil {
		return 0, err
	}
	// 将字符串转化为Time格式
	date2, err := time.ParseInLocation(format, end, time.Local)
	if err != nil {
		return 0, err
	}
	//计算相差天数
	return ToInt64(date2.Sub(date1).Hours() / 24), nil
}

func ForLimiter1Minute() int64 {
	//return time.Date(time.Now().Year(),
	//	time.Now().Month(),
	//	time.Now().Day(),
	//	time.Now().Hour(), time.Now().Second()+1, time.Now().Minute(), time.Now().Nanosecond(),
	//	time.Local).Unix()
	return time.Now().Unix() + 60
}

func RemoveDuplicatesAndEmpty(str []string) []string {
	result := []string{}
	tempMap := map[string]byte{} // 存放不重复字符串
	for _, e := range str {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

func RemoveFile(path string) {
	if err := os.Remove(path); err != nil {
		fmt.Println("删除文件时出错:", err)
		return
	}
}

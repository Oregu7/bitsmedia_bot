package helpers

import (
	"crypto/md5"
	"encoding/hex"
	"regexp"

	"github.com/syndtr/goleveldb/leveldb"
)

// Article статья
type Article struct {
	title, image, text string
	tags               []string
}

// GetMD5Hash возвращает хеш строки
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// GetUpdates возвращает новые статьи
func GetUpdates(db *leveldb.DB, items []string) []string {
	updates := make([]string, 0)
	newPatternStr := ""
	pattern, _ := db.Get([]byte("bitsmedia"), nil)
	for i, item := range items {
		// ищем обновления
		linkHash := GetMD5Hash(item)
		mtch, _ := regexp.MatchString(string(pattern), linkHash)
		if len(pattern) == 0 || !mtch {
			updates = append(updates, item)
		}
		// формируем новый паттерн с текущими данными
		newPatternStr += linkHash
		if i < len(items)-1 {
			newPatternStr += "|"
		}
	}
	// сохраняем новый паттерн
	db.Put([]byte("bitsmedia"), []byte(newPatternStr), nil)
	return updates
}

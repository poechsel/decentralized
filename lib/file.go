package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
)

func UidToHash(uid string) []byte {
	//o, _ := base64.URLEncoding.DecodeString(uid)
	o, _ := hex.DecodeString(uid)
	return o
}

func HashToUid(hash []byte) string {
	return hex.EncodeToString(hash)
}

func UidIsValidHash(uid string) bool {
	out, _ := regexp.MatchString("[a-f0-9A-F]{64}", uid)
	return out
}

var SHAREDFOLDER string = "_SharedFiles/"
var TEMPFOLDER string = "_tmp/"
var FILECHUNKSIZE int = 8 * 1024
var DOWNLOADFOLDER string = "_Downloads/"

func InitializeTempDir(server_name string) {
	TEMPFOLDER = "_tmp_" + server_name + "/"
	os.MkdirAll(TEMPFOLDER, os.ModePerm)
}

/* TODO: launch several goroutines reading separate chunk of the file
to go faster */
func SplitFile(file_name string) ([]byte, int64) {
	file, err := os.Open(SHAREDFOLDER + file_name)
	if err != nil {
		fmt.Println(err)
		return []byte{}, 0
	}
	defer file.Close()

	buffer := make([]byte, FILECHUNKSIZE)

	metafile := []byte{}

	filesize := int64(0)

	for {
		bytesread, err := file.Read(buffer)

		filesize += int64(bytesread)

		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}

		hash := sha256.Sum256(buffer[:bytesread])
		metafile = append(metafile, hash[:]...)
		uid := HashToUid(hash[:])

		if _, err := os.Stat(TEMPFOLDER + uid); os.IsNotExist(err) {
			chunk_file, err := os.Create(TEMPFOLDER + uid)
			if err != nil {
				fmt.Println(err)
			} else {
				chunk_file.Write(buffer[:bytesread])
			}
			chunk_file.Close()
		}
	}
	return metafile, filesize
}

func ReconstructFile(out_file string, metafile []byte) {
	file, err := os.Create(DOWNLOADFOLDER + out_file)
	defer file.Close()

	file.Truncate(0)

	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < len(metafile); i += 32 {
		hash := metafile[i : i+32]
		uid := HashToUid(hash)

		chunk_file, err := os.Open(TEMPFOLDER + uid)
		if err != nil {
			fmt.Println(err)
		} else {
			chunk_buffer := make([]byte, FILECHUNKSIZE)
			bytesread, _ := chunk_file.Read(chunk_buffer)
			file.Write(chunk_buffer[:bytesread])
		}
		chunk_file.Close()
	}
}
func WriteFile(name string, data []byte) {
	file, err := os.Create(name)
	defer file.Close()
	if err != nil {
		fmt.Println(err)
	} else {
		file.Write(data)
	}
}

func GetMetaHash(metafile []byte) string {
	hash := sha256.Sum256(metafile)
	return HashToUid(hash[:])
}

func WriteMetaFile(metafile []byte) {
	uid := GetMetaHash(metafile)
	WriteFile(TEMPFOLDER+uid+".meta", metafile)
}

func WriteChunkFile(chunk []byte) {
	hash := sha256.Sum256(chunk)
	uid := HashToUid(hash[:])
	WriteFile(TEMPFOLDER+uid, chunk)
}

func ReadAllFile(filename string) []byte {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	} else {
		return content
	}
}

const (
	MetaFileId  = iota
	ChunkFileId = iota
	NoFileId    = iota
)

func ReadFileForHash(hash []byte) (int, []byte) {
	uid := HashToUid(hash)
	if _, err := os.Stat(TEMPFOLDER + uid + ".meta"); !os.IsNotExist(err) {
		return MetaFileId, ReadAllFile(TEMPFOLDER + uid + ".meta")
	} else if _, err := os.Stat(TEMPFOLDER + uid); !os.IsNotExist(err) {
		return ChunkFileId, ReadAllFile(TEMPFOLDER + uid)
	}
	return NoFileId, []byte{}
}

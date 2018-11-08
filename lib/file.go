package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func UidToHash(uid string) []byte {
	//o, _ := base64.URLEncoding.DecodeString(uid)
	o, _ := hex.DecodeString(uid)
	return o
}

func HashToUid(hash []byte) string {
	return hex.EncodeToString(hash)
}

var SHAREDFOLDER string = "_SharedFiles/"
var TEMPFOLDER = "_tmp/"
var FILECHUNKSIZE int = 8 * 1024
var DOWNLOADFOLDER string = "_Downloads/"

/* TODO: launch several goroutines reading separate chunk of the file
to go faster */
func SplitFile(file_name string) []byte {
	file, err := os.Open(SHAREDFOLDER + file_name)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	defer file.Close()

	buffer := make([]byte, FILECHUNKSIZE)

	metafile := []byte{}

	for {
		bytesread, err := file.Read(buffer)

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
		} else {
			fmt.Println("file exists")
		}
	}
	return metafile
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

func WriteMetaFile(metafile []byte) {
	hash := sha256.Sum256(metafile)
	uid := HashToUid(hash[:])
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
	uid := HashToUid(hash[:])
	if _, err := os.Stat(TEMPFOLDER + uid + ".meta"); !os.IsNotExist(err) {
		return MetaFileId, ReadAllFile(TEMPFOLDER + uid + ".meta")
	} else if _, err := os.Stat(TEMPFOLDER + uid); !os.IsNotExist(err) {
		return ChunkFileId, ReadAllFile(TEMPFOLDER + uid)
	}
	return NoFileId, []byte{}
}

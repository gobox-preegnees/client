package fileinfo

import (
	"crypto/sha256"
	"fmt"
	"sync"

	cacheDTO "github.com/gobox-preegnees/gobox-client/internal/adapter/storage/cache"

	entity "github.com/gobox-preegnees/gobox-client/internal/domain/entity"
)

type fileId struct {
	fileName string
	ModTime  int64
}
type file struct {
	fileId
	HashSum  string
	SizeFile int64
}

type cache struct {
	ok   int
	fail int

	fInfoMutext sync.Mutex
	fInfo       map[fileId]file
}

func NewCache() *cache {

	return &cache{
		ok:          0,
		fail:        0,
		fInfoMutext: sync.Mutex{},
		fInfo:       make(map[fileId]file),
	}
}

func (c *cache) Get(in cacheDTO.GetFileDTO) entity.File {

	c.fInfoMutext.Lock()
	defer c.fInfoMutext.Unlock()

	info, ok := c.fInfo[fileId{
		fileName: in.FileName,
		ModTime:  in.ModTime,
	}]
	if !ok {
		c.fail++
	} else {
		c.ok++
	}

	return entity.File{
		FileName: info.fileName,
		HashSum:  info.HashSum,
		SizeFile: info.SizeFile,
		ModTime:  in.ModTime,
	}
}

func (c *cache) Put(in cacheDTO.PutFileDTO) {

	c.fInfoMutext.Lock()
	defer c.fInfoMutext.Unlock()

	c.fInfo[fileId{
		fileName: in.FileName,
		ModTime:  in.ModTime,
	}] = file{
		fileId: fileId{
			fileName: in.FileName,
			ModTime:  in.ModTime,
		},
		HashSum:  in.HashSum,
		SizeFile: in.SizeFile,
	}
}

func (c *cache) GetAll() (string, []entity.File) {

	c.fInfoMutext.Lock()
	defer c.fInfoMutext.Unlock()

	files := make([]entity.File, 0, len(c.fInfo))
	h := sha256.New()

	for _, v := range c.fInfo {
		file := entity.File{
			FileName: v.fileName,
			HashSum:  v.HashSum,
			SizeFile: v.SizeFile,
			ModTime:  v.ModTime,
		}
		h.Write([]byte(fmt.Sprintf("%v", file)))
		files = append(files, file)
	}
	return string(h.Sum(nil)), files
}

func (c *cache) Stat() (int, int) {

	return c.ok, c.fail
}

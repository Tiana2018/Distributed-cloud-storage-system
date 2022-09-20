package meta

import mydb "Distributed-cloud-storage-system/db"

// FileMeta: 文件元信息结构
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

//UpdateFileMeta 新增/更新文件元信息
func UpdateFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// UpdateFileMetaDB 新增/更新文件元信息到mysql中
func UpdateFileMetaDB(fmeta FileMeta) bool {
	return mydb.OnFileUploadFinished(fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

// GetFileMeta 通过sha1值获取文件的元信息对象
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// GetFileMetaDB 从MySQL获取文件信息
func GetFileMetaDB(fileSha1 string) (*FileMeta, error) {
	tfile, err := mydb.GetFileMeta(fileSha1)
	if tfile == nil || err != nil { //漏判断了tfile == nil
		return nil, err //这里应该是nil不然查找失败判别不了
	}
	fmeta := FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return &fmeta, nil
}
func GetLastFileMetas(count int) []FileMeta {
	fMetaArray := make([]FileMeta, len(fileMetas))
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}
	//sort.Sort(fMetaArray)
	return fMetaArray[0:count]
}

// RemoveFileMeta means remove file
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}

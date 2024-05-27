package file_util

import "macOS-auto-backup-to-Discord-be/configs"

func Chunkify(file []byte) [][]byte {
	var newChunk [][]byte
	currentChunk := file
	for len(currentChunk) > 0 {
		if len(currentChunk) <= configs.FileSizeLimit {
			newChunk = append(newChunk, currentChunk)
			break
		}
		newChunk = append(newChunk, currentChunk[:configs.FileSizeLimit])
		currentChunk = currentChunk[configs.FileSizeLimit:]
	}
	return newChunk
}

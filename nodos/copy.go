package nodos

func Copy(src, dst string, isFailIfExists bool) error {
	return copyFile(src, dst, isFailIfExists)
}

func Move(src, dst string) error {
	return moveFile(src, dst)
}

func ReadShortcut(path string) (string, string, error) {
	return readShortcut(path)
}

//go:build windows

package directchat

func createFileLoaderPipe(string) (bool, error) {
	return false, nil
}

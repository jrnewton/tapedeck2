package tapedeck

import (
	"bufio"
	"os"
)

func TouchFile(path string) (*os.File, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func DeleteFile(path string) error {
	return os.Remove(path)
}

func ReadLines(path string, handleLine func(string) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		err := handleLine(scanner.Text())
		if err != nil {
			return err
		}
	}

	return nil
}

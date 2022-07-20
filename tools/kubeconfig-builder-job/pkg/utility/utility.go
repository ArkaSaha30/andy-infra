package utility

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// GetEnv returns the environment variable value from the runtime for the provided key.
//
// If the requested key doesn't exist in the environment, returns the default value specified as the param.
//
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

// WriteFileToLocal writes the specified blob content to a local file.
//
// Specify the filename and content to be written. If the file doesnt exist, it will create, if the file exists, it will overwrite the content.
//
func WriteFileToLocal(filename string, data []byte) error {
	f, err := os.Create(filename)

	if err != nil {
		log.WithError(err).Errorf("Error creating/accessing file %s", filename)
		return err
	}

	defer f.Close()

	_, err = f.WriteString(string(data))

	if err != nil {
		log.WithError(err).Errorf("Error Writing data to file %s", filename)
		return err
	}

	return nil
}

// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.

package dockerfile

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/afero"
)

const exposeRegexPattern = `(\d+)(\/(\w+))?` // port and optional protocol, at least 1 time on a line

var (
	errCouldntParseDockerfilePort = errors.New("parse port from EXPOSE")
)

type Dockerfile interface {
	GetExposedPorts() []uint16
}

type PortConfig struct {
	Port      uint16
	Protocol  string
	RawString string
}

type Config struct {
	ExposedPorts []PortConfig
	parsed       bool
	path         string

	fs afero.Fs
}

func NewConfig(fs afero.Fs, path string) *Config {
	return &Config{
		ExposedPorts: []PortConfig{},
		fs:           fs,
		path:         path,
		parsed:       false,
	}
}

func (df *Config) GetExposedPorts() []uint16 {
	if !df.parsed {
		df.parse()
	}

	ports := []uint16{}
	for _, port := range df.ExposedPorts {
		ports = append(ports, port.Port)
	}
	return ports
}

// ParseDockerfile takes a Dockerfile and struct of methods and returns a json representation
// of all lines matching any method passed in
func (df *Config) parse() error {
	file, err := df.fs.Open(df.path)
	if err != nil {
		return fmt.Errorf("read dockerfile: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	methods := getLineParseMethods()
	parsedDockerfile := parseFromScanner(scanner, methods)

	df.ExposedPorts = parsedDockerfile.ExposedPorts
	df.parsed = true
	return nil
}

func parseFromScanner(scanner *bufio.Scanner, methods lineParseMethods) Config {
	var line = ""
	var df Config
	df.ExposedPorts = []PortConfig{}
	var currentPorts []PortConfig
	for scanner.Scan() {
		line = scanner.Text()
		prefix := strings.SplitN(line, " ", 2)[0]
		switch prefix {
		case "EXPOSE":
			currentPorts = methods.EXPOSE(line)
			df.ExposedPorts = append(df.ExposedPorts, currentPorts...)
		}
	}

	return df
}

type lineParseMethods struct {
	EXPOSE func(string) []PortConfig
}

func getLineParseMethods() lineParseMethods {
	methods := lineParseMethods{
		EXPOSE: parseExpose,
	}
	return methods
}

func parseExpose(line string) []PortConfig {
	// group 0: whole match
	// group 1: port
	// group 2: /protocol
	// group 3: protocol
	// matches strings of form <digits>(/<string>)?
	// for any number of digits and optional protocol string
	// separated by forward slash
	re := regexp.MustCompile(exposeRegexPattern)

	matches := re.FindAllStringSubmatch(line, -1)
	// check that there are matches, if not return port with only raw data
	// there will only ever be length 0 or 4 arrays
	// TODO implement arg parser regex
	if len(matches) == 0 {
		return []PortConfig{
			{
				RawString: line,
			},
		}
	}
	var port PortConfig
	ports := []PortConfig{}
	for _, match := range matches {
		port = PortConfig{
			RawString: match[0],
		}
		// set protocol if found
		port.Protocol = match[3]

		// convert the matched port to int and validate
		// We don't use the validate func in the cli package to avoid a circular dependency
		extractedPort, err := strconv.Atoi(match[1])
		if err == nil && extractedPort >= 1 && extractedPort <= 65535 {
			port.Port = uint16(extractedPort)
		}
		ports = append(ports, port)
	}
	return ports
}

/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
)

const (
	GiB              = int64(1024 * 1024 * 1024)
	DefaultBlockSize = 4096
)

var (
	isAlphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString
	// MAC Address Regex Source: https://stackoverflow.com/a/4260512
	isMACAddressRegex = regexp.MustCompile(`([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})`)
)

// RoundUpBytes rounds up the volume size in bytes up to multiplications of GiB.
func RoundUpBytes(volumeSizeBytes int64) int64 {
	return roundUpSize(volumeSizeBytes, GiB) * GiB
}

// RoundUpGiB rounds up the volume size in bytes upto multiplications of GiB
// in the unit of GiB.
func RoundUpGiB(volumeSizeBytes int64) (int32, error) {
	result := roundUpSize(volumeSizeBytes, GiB)
	if result > int64(math.MaxInt32) {
		return 0, fmt.Errorf("rounded up size exceeds maximum value of int32: %d", result)
	}
	//nolint:gosec // Integer overflow handled
	return int32(result), nil
}

// BytesToGiB converts Bytes to GiB.
func BytesToGiB(volumeSizeBytes int64) int32 {
	result := volumeSizeBytes / GiB
	if result > int64(math.MaxInt32) {
		// Handle overflow
		return math.MaxInt32
	}
	//nolint:gosec // Integer overflow handled
	return int32(result)
}

// GiBToBytes converts GiB to Bytes.
func GiBToBytes(volumeSizeGiB int32) int64 {
	return int64(volumeSizeGiB) * GiB
}

func ParseEndpoint(endpoint string, hostprocess bool) (string, string, error) {
	if runtime.GOOS == "windows" && hostprocess {
		parts := strings.SplitN(endpoint, "://", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid endpoint format: %s", endpoint)
		}
		scheme := strings.ToLower(parts[0])
		addr := parts[1]

		// Remove the socket file if it already exists
		if scheme == "unix" {
			if _, err := os.Stat(addr); err == nil {
				if err := os.Remove(addr); err != nil {
					return "", "", fmt.Errorf("failed to remove existing socket file: %w", err)
				}
			}
		}
		return scheme, addr, nil
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", fmt.Errorf("could not parse endpoint: %w", err)
	}

	addr := filepath.Join(u.Host, filepath.FromSlash(u.Path))

	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "tcp":
	case "unix":
		addr = filepath.Join("/", addr)
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			return "", "", fmt.Errorf("could not remove unix domain socket %q: %w", addr, err)
		}
	default:
		return "", "", fmt.Errorf("unsupported protocol: %s", scheme)
	}

	return scheme, addr, nil
}

func roundUpSize(volumeSizeBytes int64, allocationUnitBytes int64) int64 {
	if allocationUnitBytes == 0 {
		return 0 // Avoid division by zero
	}
	return (volumeSizeBytes + allocationUnitBytes - 1) / allocationUnitBytes
}

// GetAccessModes returns a slice containing all of the access modes defined
// in the passed in VolumeCapabilities.
func GetAccessModes(caps []*csi.VolumeCapability) *[]string {
	modes := []string{}
	for _, c := range caps {
		modes = append(modes, c.GetAccessMode().GetMode().String())
	}
	return &modes
}

func IsSBE(region string) bool {
	return region == "snow"
}

// StringIsAlphanumeric returns true if a given string contains only English letters or numbers.
func StringIsAlphanumeric(s string) bool {
	return isAlphanumericRegex(s)
}

// CountMACAddresses returns the amount of MAC addresses within a string.
func CountMACAddresses(s string) int {
	matches := isMACAddressRegex.FindAllStringIndex(s, -1)
	return len(matches)
}

// NormalizeWindowsPath normalizes a Windows path.
func NormalizeWindowsPath(path string) string {
	normalizedPath := strings.ReplaceAll(path, "/", "\\")
	if strings.HasPrefix(normalizedPath, "\\") {
		normalizedPath = "c:" + normalizedPath
	}
	return normalizedPath
}

// SanitizeRequest takes a request object and returns a copy of the request with
// the "Secrets" field cleared.
func SanitizeRequest(req interface{}) interface{} {
	v := reflect.ValueOf(&req).Elem()
	e := reflect.New(v.Elem().Type()).Elem()

	e.Set(v.Elem())

	f := reflect.Indirect(e).FieldByName("Secrets")

	if f.IsValid() && f.CanSet() && f.Kind() == reflect.Map {
		f.Set(reflect.MakeMap(f.Type()))
		v.Set(e)
	}
	return req
}

// ValueOrDefault returns the value of a pointer if it is not nil,
// otherwise it returns the default value
func ValueOrDefault[T any](v *T, d T) T {
	if v == nil {
		return d
	}
	return *v
}

func ItoaOrDefault(v *int, d string) string {
	if v == nil {
		return d
	}

	return strconv.Itoa(*v)
}

// JoinWinPath joins the elements of path with "\\" and returns the result
func JoinWinPath(elem ...string) string {
	for i, e := range elem {
		if e != "" {
			return filepath.Clean(strings.Join(elem[i:], string("\\")))
		}
	}

	return ""
}

// SerializeData is helper function to serialize data
func SerializeData[T any](msg *T) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeData is helper function to deserialize data
func DeserializeData[T any](data []byte) (*T, error) {
	buf := bytes.NewReader(data)
	msg := new(T)
	err := binary.Read(buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// GetFileFirstLine returns the first line of a file
func GetFileFirstLine(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read the first line
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text(), nil
	}

	return "", scanner.Err()
}

// IsTrue returns true if the value is "true"
func IsTrue(value string) bool {
	return value == "true"
}
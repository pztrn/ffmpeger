package config

import (
	// stdlib
	"flag"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"testing"

	// other
	"github.com/stretchr/testify/require"
)

const (
	testConfig = `nats:
  connection_string: "nats://127.0.0.1:14222"`
	testConfigBad = `nats
connection_string: "nats://127.0.0.1:14222"`
	testConfigPath = "/tmp/ffmpeger-test-config"
)

func TestConfigPackageInitialization(t *testing.T) {
	Initialize()
	require.Empty(t, configPathRaw)
}

func TestConfigFileLoad(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("ffmpeger-test-config-load", flag.ExitOnError)
	Initialize()

	err := ioutil.WriteFile(testConfigPath, []byte(testConfig), os.ModePerm)
	if err != nil {
		t.Fatal("Failed to write test config file:", err.Error())
	}

	configPathRaw = testConfigPath
	err1 := Load()
	require.Nil(t, err1)
	require.NotEmpty(t, Cfg.NATS.ConnectionString)
	require.Equal(t, "nats://127.0.0.1:14222", Cfg.NATS.ConnectionString)
}

func TestConfigFileLoadWithoutFilePath(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("ffmpeger-test-config-load", flag.ExitOnError)
	Initialize()
	err := Load()
	require.NotNil(t, err)
	require.Empty(t, Cfg.NATS.ConnectionString)
}

func TestConfigFileLoadWithoutConfigItself(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("ffmpeger-test-config-load", flag.ExitOnError)
	Initialize()
	configPathRaw = "/tmp/nonexistant"
	err := Load()
	require.NotNil(t, err)
	require.Empty(t, Cfg.NATS.ConnectionString)
}

func TestConfigFileLoadBadConfigFile(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("ffmpeger-test-config-load", flag.ExitOnError)
	Initialize()

	err := ioutil.WriteFile(testConfigPath, []byte(testConfigBad), os.ModePerm)
	if err != nil {
		t.Fatal("Failed to write test config file:", err.Error())
	}

	configPathRaw = testConfigPath
	err1 := Load()
	require.NotNil(t, err1)
	require.Empty(t, Cfg.NATS.ConnectionString)
}

func TestConfigFileLoadWithTilde(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("ffmpeger-test-config-load", flag.ExitOnError)
	Initialize()

	// Sorry, this test isn't supposed to be run on Windows. Patches
	// welcome, until don't even try to whine :).
	tildeConfigPath := "~/.cache/ffmpeger-test-config"
	u, err := user.Current()
	if err != nil {
		t.Fatal("Failed to get current user:", err.Error())
	}

	tildeConfigPathToWrite := strings.Replace(tildeConfigPath, "~", u.HomeDir, 1)
	err1 := ioutil.WriteFile(tildeConfigPathToWrite, []byte(testConfigBad), os.ModePerm)
	if err1 != nil {
		t.Fatal("Failed to write test config file:", err1.Error())
	}

	configPathRaw = tildeConfigPath
	err2 := Load()
	require.NotNil(t, err2)
	require.Empty(t, Cfg.NATS.ConnectionString)
}

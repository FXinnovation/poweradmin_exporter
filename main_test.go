package main

import "testing"

func TestMain_LoadConfig(t *testing.T) {

	config, err := loadConfig("config_example")
	if err != nil {
		t.Errorf("There were errors during the parsing %v", err)
	}
	if len(config.Groups) != 2 {
		t.Errorf("Groups filter has not been parsed correctly")
	}
	if config.Groups[0].GroupPath != "Servers/Devices^Live^Dev" {
		t.Errorf("Groups path not parsed correctly")
	}
	if config.Groups[0].Servers[0] != "Server1" {
		t.Errorf("Servers not parsed correctly first of first group should be 'Server1' got '%s'", config.Groups[0].Servers[0])
	}

}

func TestMain_LoadConfig_ShouldRaiseError(t *testing.T) {
	_, err := loadConfig("nosuchdirectory")
	if err == nil {
		t.Errorf("A nonexistent folder didn't raise an error")
	}
}

package main

// Configuration struct, this will be read from json config file
type Configuration struct {
	SlackKey  string `json:"slack-key"`
	AWSRegion string `json:"aws-region"`
	AnnounceChannel string `json:"announce-channel"`
}

type BuildStatuses map[string]string
type EnvStatuses map[string]string
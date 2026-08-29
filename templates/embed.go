package templates

import "embed"

//go:embed nginx/* sql/* php/* systemd/* public/*
var FS embed.FS

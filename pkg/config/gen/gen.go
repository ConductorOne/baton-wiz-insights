package main

import (
	"github.com/conductorone/baton-sdk/pkg/config"
	cfg "github.com/conductorone/baton-wiz-insights/pkg/config"
)

func main() {
	config.Generate("wiz-insights", cfg.Config)
}

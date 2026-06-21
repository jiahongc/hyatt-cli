// Copyright 2026 Jiahong Chen and contributors. Licensed under Apache-2.0. See LICENSE.
// Maintained in the World of Hyatt CLI repository.

package types

type CalendarPage struct {
	Title  string `json:"title"`
	Text   string `json:"text"`
	Nights int    `json:"nights"`
}

type Hotel struct {
	SpiritCode string `json:"spiritCode"`
	Name       string `json:"name"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	Category   string `json:"category"`
}

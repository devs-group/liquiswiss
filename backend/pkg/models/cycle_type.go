package models

type CycleType string

const (
	CycleMonthly    CycleType = "monthly"
	CycleQuarterly  CycleType = "quarterly"
	CycleBiannually CycleType = "biannually"
	CycleYearly     CycleType = "yearly"
)

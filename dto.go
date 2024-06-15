package main

type publishDto struct {
	Topic string `json:"topic" binding:"required"`
	Data  string `json:"data" binding:"required"`
}

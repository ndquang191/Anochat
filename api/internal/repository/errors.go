package repository

import "errors"

// ErrNotFound is returned when a record is not found.
var ErrNotFound = errors.New("record not found")

// ErrNicknameChangeCooldown is returned when an atomic nickname update loses
// the 30-day cooldown race to another request.
var ErrNicknameChangeCooldown = errors.New("nickname change cooldown is active")

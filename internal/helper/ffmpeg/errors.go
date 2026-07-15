package ffmpeg

import "errors"

// Sentinel errors for the ffmpeg package.
var (
	ErrNotFound   = errors.New("ffmpeg: binary not found in PATH")
	ErrConvert    = errors.New("ffmpeg: conversion failed")
	ErrOutputPath = errors.New("ffmpeg: invalid output path")
)

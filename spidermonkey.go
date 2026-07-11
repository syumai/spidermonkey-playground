package main

import (
	"time"

	"github.com/goccy/go-spidermonkey"
)

func Eval(code string) (string, error) {
	i, err := spidermonkey.NewInterpreter(spidermonkey.Config{})
	if err != nil {
		return "", err
	}
	defer i.Close()

	ip, err := i.PrepareInterrupt()
	if err != nil {
		return "", err
	}
	go func() {
		time.Sleep(time.Second)
		ip.Fire()
	}()

	result, err := i.Eval(code)
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}

package main

type Task interface {
	Run(args []string, lastResult []string) []string
}
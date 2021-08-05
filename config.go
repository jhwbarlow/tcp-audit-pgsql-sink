package main

type configGetter interface {
	host() string
	port() uint16
}

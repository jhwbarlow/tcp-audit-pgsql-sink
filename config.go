package main

type configGetter interface {
	config() (string, error)
}

type fixedStringConfigGetter struct{}

func (cg *fixedStringConfigGetter) config() (string, error) {
	return "postgresql://postgres:mysecretpassword@127.0.0.1/postgres", nil
}

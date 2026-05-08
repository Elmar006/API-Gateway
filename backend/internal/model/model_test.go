package model

import "testing"

func TestDB_DSN(t *testing.T) {
	d := DB{Host: "h", Port: "5432", User: "u", Password: "p", DBName: "db", SSLMode: "disable"}
	want := "host=h port=5432 user=u password=p dbname=db sslmode=disable"
	if got := d.DSN(); got != want {
		t.Fatalf("DSN=%q want %q", got, want)
	}
}

package rodbsearcher

import (
	internal "github.com/invxp/rodbsearcher/internal/http"
	"github.com/invxp/rodbsearcher/internal/util/convert"
	"net/http"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	srv, err := New(
		WithMySQLConfig(map[string]string{"timeout": "5s"}))

	if err != nil {
		panic(err)
	}

	go func() {
		panic(srv.Serv())
	}()

	time.Sleep(time.Second)

	m.Run()
}

func TestHTTPGet(t *testing.T) {
	resp, err := internal.Request(http.MethodGet, "http://localhost/test?key=k&value=v", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(convert.ByteToString(resp.Data.Payload))

	resp, err = internal.Request(http.MethodGet, "http://localhost/test?key=k", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(convert.ByteToString(resp.Data.Payload))

	resp, err = internal.Request(http.MethodGet, "http://localhost/test?value=v", nil, nil)
	if err == nil {
		t.Fatal("bind query key was not found")
	}
	t.Log(err)
}

func TestHTTPPost(t *testing.T) {
	resp, err := internal.Request(http.MethodPost, "http://localhost/api/cron", internal.RequestPOST{Key: "*/1 * * * * *", Value: convert.StringToByte("...CRON...")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(convert.ByteToString(resp.Data.Payload))

	resp, err = internal.Request(http.MethodPost, "http://localhost/api/cron", internal.RequestPOST{Key: "*/1 * * * * *"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(convert.ByteToString(resp.Data.Payload))

	resp, err = internal.Request(http.MethodPost, "http://localhost/api/cron", internal.RequestPOST{}, nil)
	if err == nil {
		t.Fatal("bind post key was not found")
	}
	t.Log(err)

	time.Sleep(5 * time.Second)
}

func TestHTTPGetCustomHeaderForAuth(t *testing.T) {
	_, err := internal.Request(http.MethodGet, "http://localhost/test?key=k&value=v", nil, map[string]string{"Auth": "FALSE"})
	if err == nil {
		t.Fatal("auth success")
	}
	t.Log(err)
}

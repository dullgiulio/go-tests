package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
)

type part struct {
	data   []byte
	header textproto.MIMEHeader
}

func parse(r io.Reader) ([]*part, error) {
	var ps []*part
	var buf bytes.Buffer
	tp := textproto.NewReader(bufio.NewReader(r))
	hdrs, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	_, params, err := mime.ParseMediaType(hdrs.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	mr := multipart.NewReader(tp.R, params["boundary"])
	// Parse the parts
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(&buf, p); err != nil {
			return nil, err
		}
        data := make([]byte, buf.Len())
        copy(data, buf.Bytes())
        ps = append(ps, &part{data, p.Header})
		buf.Reset()
	}
	return ps, nil
}

func main() {
	raw := []byte(`MIME-Version: 1.0
Subject: Test Subject
From: Jordan Wright <test@gmail.com>
To: Jordan Wright <test@gmail.com>
Content-Type: multipart/alternative; boundary=001a114fb3fc42fd6b051f834280

--001a114fb3fc42fd6b051f834280
Content-Type: text/plain; charset=UTF-8

This is a test email with HTML Formatting. It also has very long lines so
that the content must be wrapped if using quoted-printable decoding.

--001a114fb3fc42fd6b051f834280
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<div dir=3D"ltr">This is a test email with <b>HTML Formatting.</b>=C2=A0It =
also has very long lines so that the content must be wrapped if using quote=
d-printable decoding.</div>

--001a114fb3fc42fd6b051f834280--`)
	ps, err := parse(bytes.NewReader(raw))
	if err != nil {
		log.Fatal(err)
	}
	for i := range ps {
		for k := range ps[i].header {
			fmt.Printf("%s: %s\n", k, strings.Join(ps[i].header[k], ", "))
		}
		fmt.Printf("\n%s\n\n", ps[i].data)
	}
}

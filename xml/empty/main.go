package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

type propfindStruct struct {
	Multistatus string               `xml:"d:multistatus"`
	Response    []profResponseStruct `xml:"response"`
}

type profResponseStruct struct {
	//XMLName      xml.Name `xml:"response"`
	Href         string `xml:"href"`
	Filename     string `xml:"propstat>prop>filename"`
	Displayaname string `xml:"propstat>prop>displayname"`
	Resourcetype struct {
		Collection      string `xml:"collection"`
		CollectionInner string `xml:",innerxml" json:"-"`
	} `xml:"propstat>prop>resourcetype" json:",omitempty"`
	AbsoluteUri  string `xml:"propstat>prop>absoluteUri"`
	Lastmodified string `xml:"propstat>prop>lastmodified"`
	Creationdate string `xml:"propstat>prop>creationdate"`
	Status       string `xml:"propstat>status"`
}

func main() {
	bob := propfindStruct{}
	err := xml.Unmarshal([]byte(webdavresp), &bob)
	if err != nil {
		fmt.Println(err)
	}
	for i := range bob.Response {
		if bob.Response[i].Resourcetype.Collection == "" {
			if bob.Response[i].Resourcetype.CollectionInner != "" {
				// Present and empty: <d:collection />
				fmt.Printf("Node Resourcetype.Collection.%d has empty tag\n", i)
			} else {
				// Not present
				fmt.Printf("Node Resourcetype.Collection.%d has no tag\n", i)
			}
		}
	}
	js, err := json.Marshal(bob)
	if err != nil {
		return
	}
	fmt.Println(string(js))
}

var webdavresp = `<?xml version="1.0" encoding="utf-8" ?>
<d:multistatus xmlns:e="ESPDAV:" xmlns:d="DAV:">
        <d:response>
                <d:href>/apps</d:href>
                <d:propstat>
                        <d:prop>
                                <d:filename>apps</d:filename>
                                <d:displayname>apps</d:displayname>
                                <d:resourcetype><d:collection /></d:resourcetype>
                                <e:absoluteUri>/apps</e:absoluteUri>
                                <d:lastmodified>2015-04-14 13:31:53Z</d:lastmodified>
                                <d:creationdate>2014-12-03 17:08:38Z</d:creationdate>
                        </d:prop>
                        <d:status>HTTP/1.1 200 OK</d:status>
                </d:propstat>
        </d:response>
        <d:response>
                <d:href>/cafs_raw</d:href>
                <d:propstat>
                        <d:prop>
                                <d:filename>raw</d:filename>
                                <d:displayname>raw</d:displayname>
                                <d:resourcetype></d:resourcetype>
                                <e:absoluteUri>/raw</e:absoluteUri>
                                <d:lastmodified>1970-01-01 00:00:00Z</d:lastmodified>
                                <d:creationdate>1970-01-01 00:00:00Z</d:creationdate>
                        </d:prop>
                        <d:status>HTTP/1.1 200 OK</d:status>
                </d:propstat>
        </d:response>
</d:multistatus>`

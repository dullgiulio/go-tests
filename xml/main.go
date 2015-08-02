package main

import (
    "encoding/xml"
    "log"
    "fmt"
)

var xmlstring string = `<request signature="1c7f34333d4d-6904c8647b5f-4ec7aee89801-02bc">
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
    <xs:element name="methodCall">
        <xs:complexType>
            <xs:sequence>
                <xs:element name="SessionState" type="xs:string" minOccurs="0"/>
                <xs:element name="trace" type="xs:boolean" minOccurs="0"/>
                <xs:element name="params">
                    <xs:complexType>
                        <xs:sequence>
                            <xs:element name="actorID" type="xs:string" minOccurs="1" maxOccurs="1" >
                                <xs:annotation>
                                    <xs:documentation>
                                        <![CDATA[XX]]>
                                    </xs:documentation>
                                </xs:annotation>
                            </xs:element>
                            <xs:element name="verb" type="xs:string" minOccurs="1" maxOccurs="1" >
                                <xs:annotation>
                                    <xs:documentation>
                                        <![CDATA[Activity type verb that describes the activity such as post.]]>
                                    </xs:documentation>
                                </xs:annotation>
                            </xs:element>
                            <xs:element name="content" type="xs:string" minOccurs="1" maxOccurs="1" >
                                <xs:annotation>
                                    <xs:documentation>
                                        <![CDATA[XX]]>
                                    </xs:documentation>
                                </xs:annotation>
                            </xs:element>
                            <xs:element name="extra" type="xs:string" minOccurs="0" >
                                <xs:annotation>
                                    <xs:documentation>
                                        <![CDATA[XX]]>
                                    </xs:documentation>
                                </xs:annotation>
                            </xs:element>
                            <xs:element name="objectURI" type="xs:string" minOccurs="0" maxOccurs="1" >
                                <xs:annotation>
                                    <xs:documentation>
                                        <![CDATA[XX]]>
                                    </xs:documentation>
                                </xs:annotation>
                            </xs:element>
                            <xs:element name="activityStreamID" type="xs:string" minOccurs="1" maxOccurs="1" >
                                <xs:annotation>
                                    <xs:documentation>
                                        <![CDATA[XX]]>
                                    </xs:documentation>
                                </xs:annotation>
                            </xs:element>
                            <xs:element name="title" type="xs:string" minOccurs="0" maxOccurs="1" >
                                <xs:annotation>
                                    <xs:documentation>
                                        <![CDATA[XX]]>
                                    </xs:documentation>
                                </xs:annotation>
                            </xs:element>
                            <xs:element name="visibility" type="activityVisibilityType" minOccurs="0" default="friend" >
                                <xs:annotation>
                                    <xs:documentation>
                                        <![CDATA[XX]]>
                                    </xs:documentation>
                                </xs:annotation>
                            </xs:element>
                            <xs:element name="location" type="geoLocationType" minOccurs="0" >
                                <xs:annotation>
                                    <xs:documentation>
                                        <![CDATA[XX]]>
                                    </xs:documentation>
                                </xs:annotation>
                            </xs:element>
                            <xs:element name="type" type="xs:string" minOccurs="0" >
                                <xs:annotation>
                                    <xs:documentation>
                                        <![CDATA[XX]]>
                                    </xs:documentation>
                                </xs:annotation>
                            </xs:element>
                            <xs:element name="objectRefUrn" type="xs:anyURI" minOccurs="0" >
                                <xs:annotation>
                                    <xs:documentation>
                                        <![CDATA[XX]]>
                                    </xs:documentation>
                                </xs:annotation>
                            </xs:element>
                            <xs:element name="activityType" type="activityType" minOccurs="0" >
                                <xs:annotation>
                                    <xs:documentation>
                                        <![CDATA[XX]]>
                                    </xs:documentation>
                                </xs:annotation>
                            </xs:element>
                        </xs:sequence>
                    </xs:complexType>
                </xs:element>
            </xs:sequence>
            <xs:attribute name="service" type="xs:string" use="required"/>
            <xs:attribute name="method" type="xs:string" use="required"/>
        </xs:complexType>
    </xs:element>
</xs:schema>    
</request>`

type Request struct {
    XMLName xml.Name `xml:"request"`
    Element []struct {
        Name string `xml:"name,attr"`
        Doc struct {
            Body  string `xml:",chardata"`
        } `xml:"annotation>documentation"`
    } `xml:"schema>element>complexType>sequence>element>complexType>sequence>element"`
}

func main() {
    response := Request{}
    if err := xml.Unmarshal([]byte(xmlstring), &response); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%v\n", response)
}

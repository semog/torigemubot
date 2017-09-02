package main

import "encoding/xml"

type misc struct {
	Grade int `xml:"grade"`
	JLPT  int `xml:"jlpt"`
}
type character struct {
	XMLName xml.Name `xml:"character"`
	Literal string   `xml:"literal"`
	Misc    misc     `xml:"misc"`
}
type kanjidic struct {
	XMLName   xml.Name    `xml:"kanjidic2"`
	Character []character `xml:"character"`
}

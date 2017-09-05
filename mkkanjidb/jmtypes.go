package main

import "encoding/xml"

type kele struct {
	XMLName xml.Name `xml:"k_ele"`
	Keb     string   `xml:"keb"`
	Kinf    []string `xml:"ke_inf"`
	Kpri    []string `xml:"ke_pri"`
}
type rele struct {
	XMLName xml.Name `xml:"r_ele"`
	Reb     string   `xml:"reb"`
	NoKanji *string  `xml:"re_nokanji"`
	RestrTo []string `xml:"re_restr"`
	Rinf    []string `xml:"re_inf"`
	Rpri    []string `xml:"re_pri"`
}
type sense struct {
	XMLName        xml.Name  `xml:"sense"`
	RestrToKanji   []string  `xml:"stagk"`
	RestrToReading []string  `xml:"stagr"`
	Pos            []string  `xml:"pos"`
	Xref           []string  `xml:"xref"`
	Antonym        []string  `xml:"ant"`
	Field          []string  `xml:"field"`
	Misc           []string  `xml:"misc"`
	Sinfo          []string  `xml:"s_inf"`
	Lsource        []lsource `xml:"lsource"`
	Dialect        []string  `xml:"dial"`
	Glossy         []gloss   `xml:"gloss"`
}
type gloss struct {
	XMLName xml.Name `xml:"gloss"`
	Lang    string   `xml:"xml:lang,attr"`
	Gender  string   `xml:"g_gend,attr"`
	Pri     []string `xml:"pri"`
}
type lsource struct {
	XMLName   xml.Name `xml:"lsource"`
	Lang      string   `xml:"xml:lang,attr"`
	LangType  string   `xml:"ls_type,attr"`
	LangWasei string   `xml:"ls_wasei,attr"`
}
type entry struct {
	XMLName xml.Name `xml:"entry"`
	Seq     int      `xml:"ent_seq"`
	Kele    []kele   `xml:"k_ele"`
	Rele    []rele   `xml:"r_ele"`
	Sense   []sense  `xml:"sense"`
}
type jmdict struct {
	XMLName xml.Name `xml:"JMdict"`
	Entry   []entry  `xml:"entry"`
}

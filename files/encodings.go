package files

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

var Encodings = map[string]encoding.Encoding{
	// IBM / DOS Code Pages
	"cp037":  charmap.CodePage037,
	"cp437":  charmap.CodePage437,
	"cp850":  charmap.CodePage850,
	"cp852":  charmap.CodePage852,
	"cp855":  charmap.CodePage855,
	"cp858":  charmap.CodePage858,
	"cp860":  charmap.CodePage860,
	"cp862":  charmap.CodePage862,
	"cp863":  charmap.CodePage863,
	"cp865":  charmap.CodePage865,
	"cp866":  charmap.CodePage866,
	"cp1047": charmap.CodePage1047,
	"cp1140": charmap.CodePage1140,

	// ISO-8859
	"iso-8859-1":  charmap.ISO8859_1,
	"iso-8859-2":  charmap.ISO8859_2,
	"iso-8859-3":  charmap.ISO8859_3,
	"iso-8859-4":  charmap.ISO8859_4,
	"iso-8859-5":  charmap.ISO8859_5,
	"iso-8859-6":  charmap.ISO8859_6,
	"iso-8859-6e": charmap.ISO8859_6E,
	"iso-8859-6i": charmap.ISO8859_6I,
	"iso-8859-7":  charmap.ISO8859_7,
	"iso-8859-8":  charmap.ISO8859_8,
	"iso-8859-8e": charmap.ISO8859_8E,
	"iso-8859-8i": charmap.ISO8859_8I,
	"iso-8859-9":  charmap.ISO8859_9,
	"iso-8859-10": charmap.ISO8859_10,
	"iso-8859-13": charmap.ISO8859_13,
	"iso-8859-14": charmap.ISO8859_14,
	"iso-8859-15": charmap.ISO8859_15,
	"iso-8859-16": charmap.ISO8859_16,

	// KOI8
	"koi8-r": charmap.KOI8R,
	"koi8-u": charmap.KOI8U,

	// Macintosh
	"macintosh":          charmap.Macintosh,
	"macintosh-cyrillic": charmap.MacintoshCyrillic,

	// Windows-125x
	"windows-874":  charmap.Windows874,
	"windows-1250": charmap.Windows1250,
	"windows-1251": charmap.Windows1251,
	"windows-1252": charmap.Windows1252,
	"windows-1253": charmap.Windows1253,
	"windows-1254": charmap.Windows1254,
	"windows-1255": charmap.Windows1255,
	"windows-1256": charmap.Windows1256,
	"windows-1257": charmap.Windows1257,
	"windows-1258": charmap.Windows1258,

	// Misc
	"x-user-defined": charmap.XUserDefined,

	// Japanese
	"shift-jis":   japanese.ShiftJIS,
	"euc-jp":      japanese.EUCJP,
	"iso-2022-jp": japanese.ISO2022JP,

	// Korean
	"euc-kr": korean.EUCKR,

	// Simplified Chinese
	"gbk":       simplifiedchinese.GBK,
	"hz-gb2312": simplifiedchinese.HZGB2312,

	// Traditional Chinese
	"big5": traditionalchinese.Big5,
}

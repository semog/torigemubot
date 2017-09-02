package main

var jmentitymap = map[string]string{
	"MA":        "MA",        // MA // martial arts term
	"X":         "X",         // X // rude or X-rated term (not displayed in educational software)
	"abbr":      "abbr",      // abbreviation
	"adj-i":     "adj-i",     // adjective (keiyoushi)
	"adj-ix":    "adj-ix",    // adjective (keiyoushi) - yoi/ii class
	"adj-na":    "adj-na",    // adjectival nouns or quasi-adjectives (keiyodoshi)
	"adj-no":    "adj-no",    // nouns which may take the genitive case particle 'no'
	"adj-pn":    "adj-pn",    // pre-noun adjectival (rentaishi)
	"adj-t":     "adj-t",     // 'taru' adjective
	"adj-f":     "adj-f",     // noun or verb acting prenominally
	"adv":       "adv",       // adverb (fukushi)
	"adv-to":    "adv-to",    // adverb taking the 'to' particle
	"arch":      "arch",      // archaism
	"ateji":     "ateji",     // ateji (phonetic) reading
	"aux":       "aux",       // auxiliary
	"aux-v":     "aux-v",     // auxiliary verb
	"aux-adj":   "aux-adj",   // auxiliary adjective
	"Buddh":     "Buddh",     // Buddhist term
	"chem":      "chem",      // chemistry term
	"chn":       "chn",       // children's language
	"col":       "col",       // colloquialism
	"comp":      "comp",      // computer terminology
	"conj":      "conj",      // conjunction
	"cop-da":    "cop-da",    // copula
	"ctr":       "ctr",       // counter
	"derog":     "derog",     // derogatory
	"eK":        "eK",        // exclusively kanji
	"ek":        "ek",        // exclusively kana
	"exp":       "exp",       // expressions (phrases, clauses, etc.)
	"fam":       "fam",       // familiar language
	"fem":       "fem",       // female term or language
	"food":      "food",      // food term
	"geom":      "geom",      // geometry term
	"gikun":     "gikun",     // gikun (meaning as reading) or jukujikun (special kanji reading)
	"hon":       "hon",       // honorific or respectful (sonkeigo) language
	"hum":       "hum",       // humble (kenjougo) language
	"iK":        "iK",        // word containing irregular kanji usage
	"id":        "id",        // idiomatic expression
	"ik":        "ik",        // word containing irregular kana usage
	"int":       "int",       // interjection (kandoushi)
	"io":        "io",        // irregular okurigana usage
	"iv":        "iv",        // irregular verb
	"ling":      "ling",      // linguistics terminology
	"m-sl":      "m-sl",      // manga slang
	"male":      "male",      // male term or language
	"male-sl":   "male-sl",   // male slang
	"math":      "math",      // mathematics
	"mil":       "mil",       // military
	"n":         "n",         // noun (common) (futsuumeishi)
	"n-adv":     "n-adv",     // adverbial noun (fukushitekimeishi)
	"n-suf":     "n-suf",     // noun, used as a suffix
	"n-pref":    "n-pref",    // noun, used as a prefix
	"n-t":       "n-t",       // noun (temporal) (jisoumeishi)
	"num":       "num",       // numeric
	"oK":        "oK",        // word containing out-dated kanji
	"obs":       "obs",       // obsolete term
	"obsc":      "obsc",      // obscure term
	"ok":        "ok",        // out-dated or obsolete kana usage
	"oik":       "oik",       // old or irregular kana form
	"on-mim":    "on-mim",    // onomatopoeic or mimetic word
	"pn":        "pn",        // pronoun
	"poet":      "poet",      // poetical term
	"pol":       "pol",       // polite (teineigo) language
	"pref":      "pref",      // prefix
	"proverb":   "proverb",   // proverb
	"prt":       "prt",       // particle
	"physics":   "physics",   // physics terminology
	"rare":      "rare",      // rare
	"sens":      "sens",      // sensitive
	"sl":        "sl",        // slang
	"suf":       "suf",       // suffix
	"uK":        "uK",        // word usually written using kanji alone
	"uk":        "uk",        // word usually written using kana alone
	"unc":       "unc",       // unclassified
	"yoji":      "yoji",      // yojijukugo
	"v1":        "v1",        // Ichidan verb
	"v1-s":      "v1-s",      // Ichidan verb - kureru special class
	"v2a-s":     "v2a-s",     // Nidan verb with 'u' ending (archaic)
	"v4h":       "v4h",       // Yodan verb with 'hu/fu' ending (archaic)
	"v4r":       "v4r",       // Yodan verb with 'ru' ending (archaic)
	"v5aru":     "v5aru",     // Godan verb - -aru special class
	"v5b":       "v5b",       // Godan verb with 'bu' ending
	"v5g":       "v5g",       // Godan verb with 'gu' ending
	"v5k":       "v5k",       // Godan verb with 'ku' ending
	"v5k-s":     "v5k-s",     // Godan verb - Iku/Yuku special class
	"v5m":       "v5m",       // Godan verb with 'mu' ending
	"v5n":       "v5n",       // Godan verb with 'nu' ending
	"v5r":       "v5r",       // Godan verb with 'ru' ending
	"v5r-i":     "v5r-i",     // Godan verb with 'ru' ending (irregular verb)
	"v5s":       "v5s",       // Godan verb with 'su' ending
	"v5t":       "v5t",       // Godan verb with 'tsu' ending
	"v5u":       "v5u",       // Godan verb with 'u' ending
	"v5u-s":     "v5u-s",     // Godan verb with 'u' ending (special class)
	"v5uru":     "v5uru",     // Godan verb - Uru old class verb (old form of Eru)
	"vz":        "vz",        // Ichidan verb - zuru verb (alternative form of -jiru verbs)
	"vi":        "vi",        // intransitive verb
	"vk":        "vk",        // Kuru verb - special class
	"vn":        "vn",        // irregular nu verb
	"vr":        "vr",        // irregular ru verb, plain form ends with -ri
	"vs":        "vs",        // noun or participle which takes the aux. verb suru
	"vs-c":      "vs-c",      // su verb - precursor to the modern suru
	"vs-s":      "vs-s",      // suru verb - special class
	"vs-i":      "vs-i",      // suru verb - irregular
	"kyb":       "kyb",       // Kyoto-ben
	"osb":       "osb",       // Osaka-ben
	"ksb":       "ksb",       // Kansai-ben
	"ktb":       "ktb",       // Kantou-ben
	"tsb":       "tsb",       // Tosa-ben
	"thb":       "thb",       // Touhoku-ben
	"tsug":      "tsug",      // Tsugaru-ben
	"kyu":       "kyu",       // Kyuushuu-ben
	"rkb":       "rkb",       // Ryuukyuu-ben
	"nab":       "nab",       // Nagano-ben
	"hob":       "hob",       // Hokkaido-ben
	"vt":        "vt",        // transitive verb
	"vulg":      "vulg",      // vulgar expression or word
	"adj-kari":  "adj-kari",  // 'kari' adjective (archaic)
	"adj-ku":    "adj-ku",    // 'ku' adjective (archaic)
	"adj-shiku": "adj-shiku", // 'shiku' adjective (archaic)
	"adj-nari":  "adj-nari",  // archaic/formal form of na-adjective
	"n-pr":      "n-pr",      // proper noun
	"v-unspec":  "v-unspec",  // verb unspecified
	"v4k":       "v4k",       // Yodan verb with 'ku' ending (archaic)
	"v4g":       "v4g",       // Yodan verb with 'gu' ending (archaic)
	"v4s":       "v4s",       // Yodan verb with 'su' ending (archaic)
	"v4t":       "v4t",       // Yodan verb with 'tsu' ending (archaic)
	"v4n":       "v4n",       // Yodan verb with 'nu' ending (archaic)
	"v4b":       "v4b",       // Yodan verb with 'bu' ending (archaic)
	"v4m":       "v4m",       // Yodan verb with 'mu' ending (archaic)
	"v2k-k":     "v2k-k",     // Nidan verb (upper class) with 'ku' ending (archaic)
	"v2g-k":     "v2g-k",     // Nidan verb (upper class) with 'gu' ending (archaic)
	"v2t-k":     "v2t-k",     // Nidan verb (upper class) with 'tsu' ending (archaic)
	"v2d-k":     "v2d-k",     // Nidan verb (upper class) with 'dzu' ending (archaic)
	"v2h-k":     "v2h-k",     // Nidan verb (upper class) with 'hu/fu' ending (archaic)
	"v2b-k":     "v2b-k",     // Nidan verb (upper class) with 'bu' ending (archaic)
	"v2m-k":     "v2m-k",     // Nidan verb (upper class) with 'mu' ending (archaic)
	"v2y-k":     "v2y-k",     // Nidan verb (upper class) with 'yu' ending (archaic)
	"v2r-k":     "v2r-k",     // Nidan verb (upper class) with 'ru' ending (archaic)
	"v2k-s":     "v2k-s",     // Nidan verb (lower class) with 'ku' ending (archaic)
	"v2g-s":     "v2g-s",     // Nidan verb (lower class) with 'gu' ending (archaic)
	"v2s-s":     "v2s-s",     // Nidan verb (lower class) with 'su' ending (archaic)
	"v2z-s":     "v2z-s",     // Nidan verb (lower class) with 'zu' ending (archaic)
	"v2t-s":     "v2t-s",     // Nidan verb (lower class) with 'tsu' ending (archaic)
	"v2d-s":     "v2d-s",     // Nidan verb (lower class) with 'dzu' ending (archaic)
	"v2n-s":     "v2n-s",     // Nidan verb (lower class) with 'nu' ending (archaic)
	"v2h-s":     "v2h-s",     // Nidan verb (lower class) with 'hu/fu' ending (archaic)
	"v2b-s":     "v2b-s",     // Nidan verb (lower class) with 'bu' ending (archaic)
	"v2m-s":     "v2m-s",     // Nidan verb (lower class) with 'mu' ending (archaic)
	"v2y-s":     "v2y-s",     // Nidan verb (lower class) with 'yu' ending (archaic)
	"v2r-s":     "v2r-s",     // Nidan verb (lower class) with 'ru' ending (archaic)
	"v2w-s":     "v2w-s",     // Nidan verb (lower class) with 'u' ending and 'we' conjugation (archaic)
	"archit":    "archit",    // architecture term
	"astron":    "astron",    // astronomy, etc. term
	"baseb":     "baseb",     // baseball term
	"biol":      "biol",      // biology term
	"bot":       "bot",       // botany term
	"bus":       "bus",       // business term
	"econ":      "econ",      // economics term
	"engr":      "engr",      // engineering term
	"finc":      "finc",      // finance term
	"geol":      "geol",      // geology, etc. term
	"law":       "law",       // law, etc. term
	"mahj":      "mahj",      // mahjong term
	"med":       "med",       // medicine, etc. term
	"music":     "music",     // music term
	"Shinto":    "Shinto",    // Shinto term
	"shogi":     "shogi",     // shogi term
	"sports":    "sports",    // sports term
	"sumo":      "sumo",      // sumo term
	"zool":      "zool",      // zoology term
	"joc":       "joc",       // jocular, humorous term
	"anat":      "anat",      // anatomical term
}

package domain

type IdacResponse struct {
	CalcDate string             `json:"calcDate"`
	Records  []TimeAttackRecord `json:"records"` // Sometimes it's "ranking" or "list"
}

type TimeAttackRecord struct {
	Rank       string `json:"rank"`
	Name       string `json:"name"`
	ShopName   string `json:"shopname"`
	Record     string `json:"record"`
	CarName    string `json:"carname"`
	MyTitleID  string `json:"mytitleId"`
	UpdateDate string `json:"updateDate"`
}

// --- ALIAS CONFIGURATION ---
var CourseAliases = map[string]string{
	// --- Akina Lake (0/2) ---
	"秋名湖／左周り":  "course-0",
	"akina lake ccw": "course-0",
	"lake ccw":       "course-0",
	"秋名湖／右周り":  "course-2",
	"akina lake cw":  "course-2",
	"lake cw":        "course-2",

	// --- Hakone (52/54) ---
	"箱根／下り": "course-52",
	"hakone dh": "course-52",
	"箱根／上り": "course-54",
	"hakone uh": "course-54",
	"hakone hc": "course-54",

	// --- Usui (36/38) ---
	"碓氷／左周り": "course-36",
	"usui ccw":    "course-36",
	"碓氷／右周り": "course-38",
	"usui cw":     "course-38",

	// --- Myogi (4/6) ---
	"妙義／下り": "course-4",
	"myogi dh":  "course-4",
	"妙義／上り": "course-6",
	"myogi uh":  "course-6",
	"myogi hc":  "course-6",

	// --- Akagi (8/10) ---
	"赤城／下り": "course-8",
	"akagi dh":  "course-8",
	"赤城／上り": "course-10",
	"akagi uh":  "course-10",
	"akagi hc":  "course-10",

	// --- Akina (12/14) ---
	"秋名／下り": "course-12",
	"akina dh":  "course-12",
	"秋名／上り": "course-14",
	"akina uh":  "course-14",
	"akina hc":  "course-14",

	// --- Irohazaka (16/18) ---
	"いろは坂／下り": "course-16",
	"irohazaka dh":  "course-16",
	"iro dh":        "course-16",
	"いろは坂／逆走": "course-18",
	"irohazaka uh":  "course-18", // Reverse often acts as Uphill/HC
	"irohazaka hc":  "course-18",
	"iro uh":        "course-18",
	"iro hc":        "course-18",

	// --- Yabitsu (76/78) ---
	"ヤビツ／下り": "course-76",
	"yabitsu dh":  "course-76",
	"yabi dh":     "course-76",
	"ヤビツ／上り": "course-78",
	"yabitsu uh":  "course-78",
	"yabitsu hc":  "course-78",
	"yabi uh":     "course-78",

	// --- Momiji Line (56/58) ---
	"もみじライン／下り": "course-56",
	"momiji dh":         "course-56",
	"もみじライン／上り": "course-58",
	"momiji uh":         "course-58",
	"momiji hc":         "course-58",

	// --- Tsukuba (20/22) ---
	"筑波／往路":  "course-20",
	"tsukuba ob": "course-20",
	"筑波／復路":  "course-22",
	"tsukuba ib": "course-22",

	// --- Happogahara (24/26) ---
	"八方ヶ原／往路":  "course-24",
	"happogahara ob": "course-24",
	"happo ob":       "course-24",
	"八方ヶ原／復路":  "course-26",
	"happogahara ib": "course-26",
	"happo ib":       "course-26",

	// --- Sadamine (40/42) ---
	"定峰／下り":   "course-40",
	"sadamine dh": "course-40",
	"sada dh":     "course-40",
	"定峰／上り":   "course-42",
	"sadamine uh": "course-42",
	"sada uh":     "course-42",

	// --- Tsuchisaka (44/46) ---
	"土坂／往路":     "course-44",
	"tsuchisaka ob": "course-44",
	"tsuchi ob":     "course-44",
	"土坂／復路":     "course-46",
	"tsuchisaka ib": "course-46",
	"tsuchi ib":     "course-46",

	// --- Nagao (28/30) ---
	"長尾／下り": "course-28",
	"nagao dh":  "course-28",
	"長尾／上り": "course-30",
	"nagao uh":  "course-30",
	"nagao hc":  "course-30",

	// --- Nanamagari (60/62) ---
	"七曲り／下り":   "course-60",
	"nanamagari dh": "course-60",
	"nana dh":       "course-60",
	"七曲り／上り":   "course-62",
	"nanamagari uh": "course-62",
	"nana uh":       "course-62",

	// --- Tsubaki Line (32/34) ---
	"椿ライン／下り": "course-32",
	"tsubaki dh":    "course-32",
	"椿ライン／上り": "course-34",
	"tsubaki uh":    "course-34",
	"tsubaki hc":    "course-34",

	// --- Akina Snow (48/50) ---
	"秋名（雪）／下り": "course-48",
	"akina snow dh": "course-48",
	"秋名（雪）／上り": "course-50",
	"akina snow uh": "course-50",

	// --- Gunsai (64/66) ---
	"群サイ／往路": "course-64",
	"gunsai ob":   "course-64",
	"群サイ／復路": "course-66",
	"gunsai ib":   "course-66",

	// --- Odawara (68/70) ---
	"小田原／順走": "course-68",
	"odawara":     "course-68",
	"小田原／逆走": "course-70",
	"odawara rev": "course-70",

	// --- Tsukuba Snow (72/74) ---
	"筑波（雪）／往路":   "course-72",
	"tsukuba snow ob": "course-72",
	"筑波（雪）／復路":   "course-74",
	"tsukuba snow ib": "course-74",

	// --- Tsuchisaka Snow (80/82) ---
	"土坂（雪）／往路":      "course-80",
	"tsuchisaka snow ob": "course-80",
	"土坂（雪）／復路":      "course-82",
	"tsuchisaka snow ib": "course-82",

	// --- Manazuru (84/86) ---
	"真鶴／順走":    "course-84",
	"manazuru":     "course-84",
	"mana":         "course-84",
	"真鶴／逆走":    "course-86",
	"manazuru rev": "course-86",
	"mana rev":     "course-86",

	// --- Usui Snow (88/90) ---
	"碓氷（雪）／左周り": "course-88",
	"usui snow ccw":   "course-88",
	"碓氷（雪）／右周り": "course-90",
	"usui snow cw":    "course-90",

	// --- Akina Rain (92/94) ---
	"秋名（雨）／下り": "course-92",
	"akina rain dh": "course-92",
	"秋名（雨）／上り": "course-94",
	"akina rain uh": "course-94",
}

// Display Name Mapping
var CourseDisplayNameByCode = map[string]string{
	"course-0":  "Akina Lake (CCW)",
	"course-2":  "Akina Lake (CW)",
	"course-52": "Hakone (DH)",
	"course-54": "Hakone (UH)",
	"course-36": "Usui (CCW)",
	"course-38": "Usui (CW)",
	"course-4":  "Myogi (DH)",
	"course-6":  "Myogi (UH)",
	"course-8":  "Akagi (DH)",
	"course-10": "Akagi (UH)",
	"course-12": "Akina (DH)",
	"course-14": "Akina (UH)",
	"course-16": "Irohazaka (DH)",
	"course-18": "Irohazaka (HC)",
	"course-76": "Yabitsu (DH)",
	"course-78": "Yabitsu (UH)",
	"course-56": "Momiji Line (DH)",
	"course-58": "Momiji Line (UH)",
	"course-20": "Tsukuba (OB)",
	"course-22": "Tsukuba (IB)",
	"course-24": "Happogahara (OB)",
	"course-26": "Happogahara (IB)",
	"course-40": "Sadamine (DH)",
	"course-42": "Sadamine (UH)",
	"course-44": "Tsuchisaka (OB)",
	"course-46": "Tsuchisaka (IB)",
	"course-28": "Nagao (DH)",
	"course-30": "Nagao (UH)",
	"course-60": "Nanamagari (DH)",
	"course-62": "Nanamagari (UH)",
	"course-32": "Tsubaki Line (DH)",
	"course-34": "Tsubaki Line (UH)",
	"course-48": "Akina Snow (DH)",
	"course-50": "Akina Snow (UH)",
	"course-64": "Gunsai (OB)",
	"course-66": "Gunsai (IB)",
	"course-68": "Odawara (OB)",
	"course-70": "Odawara (IB)",
	"course-72": "Tsukuba Snow (OB)",
	"course-74": "Tsukuba Snow (IB)",
	"course-80": "Tsuchisaka Snow (OB)",
	"course-82": "Tsuchisaka Snow (IB)",
	"course-84": "Manazuru (OB)",
	"course-86": "Manazuru (IB)",
	"course-88": "Usui Snow (CCW)",
	"course-90": "Usui Snow (CW)",
	"course-92": "Akina Rain (DH)",
	"course-94": "Akina Rain (UH)",
}

var AreaAliases = map[string]string{
	// Global
	"all": "area-all",

	// Domestic (Japan)
	"北海道": "area-0", "hokkaido": "area-0",
	"青森県": "area-1", "aomori": "area-1",
	"岩手県": "area-2", "iwate": "area-2",
	"宮城県": "area-3", "miyagi": "area-3",
	"福島県": "area-4", "fukushima": "area-4",
	"山形県": "area-5", "yamagata": "area-5",
	"秋田県": "area-6", "akita": "area-6",
	"茨城県": "area-7", "ibaraki": "area-7",
	"栃木県": "area-8", "tochigi": "area-8",
	"群馬県": "area-9", "gunma": "area-9",
	"千葉県": "area-10", "chiba": "area-10",
	"埼玉県": "area-11", "saitama": "area-11",
	"東京都": "area-12", "tokyo": "area-12",
	"神奈川県": "area-13", "kanagawa": "area-13",
	"山梨県": "area-14", "yamanashi": "area-14",
	"新潟県": "area-15", "niigata": "area-15",
	"長野県": "area-16", "nagano": "area-16",
	"富山県": "area-17", "toyama": "area-17",
	"石川県": "area-18", "ishikawa": "area-18",
	"愛知県": "area-19", "aichi": "area-19",
	"静岡県": "area-20", "shizuoka": "area-20",
	"岐阜県": "area-21", "gifu": "area-21",
	"三重県": "area-22", "mie": "area-22",
	"福井県": "area-23", "fukui": "area-23",
	"大阪府": "area-24", "osaka": "area-24",
	"京都府": "area-25", "kyoto": "area-25",
	"奈良県": "area-26", "nara": "area-26",
	"滋賀県": "area-27", "shiga": "area-27",
	"和歌山県": "area-28", "wakayama": "area-28",
	"兵庫県": "area-29", "hyogo": "area-29",
	"広島県": "area-30", "hiroshima": "area-30",
	"鳥取県": "area-31", "tottori": "area-31",
	"島根県": "area-32", "shimane": "area-32",
	"岡山県": "area-33", "okayama": "area-33",
	"山口県": "area-34", "yamaguchi": "area-34",
	"徳島県": "area-35", "tokushima": "area-35",
	"香川県": "area-36", "kagawa": "area-36",
	"愛媛県": "area-37", "ehime": "area-37",
	"高知県": "area-38", "kochi": "area-38",
	"福岡県": "area-39", "fukuoka": "area-39",
	"佐賀県": "area-40", "saga": "area-40",
	"長崎県": "area-41", "nagasaki": "area-41",
	"熊本県": "area-42", "kumamoto": "area-42",
	"大分県": "area-43", "oita": "area-43",
	"宮崎県": "area-44", "miyazaki": "area-44",
	"鹿児島県": "area-45", "kagoshima": "area-45",
	"沖縄県": "area-46", "okinawa": "area-46",

	// International
	"マカオ": "area-47", "macau": "area-47", "mo": "area-47",
	"香港": "area-48", "hong kong": "area-48", "hk": "area-48",
	"韓国": "area-49", "korea": "area-49", "kr": "area-49",
	"マレーシア": "area-50", "malaysia": "area-50", "my": "area-50",
	"シンガポール": "area-51", "singapore": "area-51", "sg": "area-51",
	"台湾": "area-52", "taiwan": "area-52", "tw": "area-52",
	"インドネシア": "area-53", "indonesia": "area-53", "id": "area-53",
	"フィリピン": "area-54", "philippines": "area-54", "ph": "area-54",
	"タイ": "area-55", "thailand": "area-55", "th": "area-55",
	"アメリカ": "area-56", "usa": "area-56", "us": "area-56",
	"ベトナム": "area-57", "vietnam": "area-57", "vn": "area-57",
	"ミャンマー": "area-58", "myanmar": "area-58", "mm": "area-58",
	"オーストラリア": "area-59", "australia": "area-59", "au": "area-59",
	"ニュージーランド": "area-60", "new zealand": "area-60", "nz": "area-60",
	"カンボジア": "area-61", "cambodia": "area-61", "kh": "area-61",
}

var AreaDisplayNameByCode = map[string]string{
	"area-all": "All Areas",
	"area-0":   "Hokkaido", "area-1": "Aomori", "area-2": "Iwate", "area-3": "Miyagi", "area-4": "Fukushima",
	"area-5": "Yamagata", "area-6": "Akita", "area-7": "Ibaraki", "area-8": "Tochigi", "area-9": "Gunma",
	"area-10": "Chiba", "area-11": "Saitama", "area-12": "Tokyo", "area-13": "Kanagawa", "area-14": "Yamanashi",
	"area-15": "Niigata", "area-16": "Nagano", "area-17": "Toyama", "area-18": "Ishikawa", "area-19": "Aichi",
	"area-20": "Shizuoka", "area-21": "Gifu", "area-22": "Mie", "area-23": "Fukui", "area-24": "Osaka",
	"area-25": "Kyoto", "area-26": "Nara", "area-27": "Shiga", "area-28": "Wakayama", "area-29": "Hyogo",
	"area-30": "Hiroshima", "area-31": "Tottori", "area-32": "Shimane", "area-33": "Okayama", "area-34": "Yamaguchi",
	"area-35": "Tokushima", "area-36": "Kagawa", "area-37": "Ehime", "area-38": "Kochi", "area-39": "Fukuoka",
	"area-40": "Saga", "area-41": "Nagasaki", "area-42": "Kumamoto", "area-43": "Oita", "area-44": "Miyazaki",
	"area-45": "Kagoshima", "area-46": "Okinawa",
	"area-47": "Macau", "area-48": "Hong Kong", "area-49": "Korea", "area-50": "Malaysia", "area-51": "Singapore",
	"area-52": "Taiwan", "area-53": "Indonesia", "area-54": "Philippines", "area-55": "Thailand", "area-56": "USA",
	"area-57": "Vietnam", "area-58": "Myanmar", "area-59": "Australia", "area-60": "New Zealand", "area-61": "Cambodia",
}

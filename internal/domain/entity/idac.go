package entity

import (
	"fmt"
	"strconv"
	"strings"
)

// ta
type IdacTimeAttackRecordResponse struct {
	CalcDate string             `json:"calcDate"`
	Records  []TimeAttackRecord `json:"records"`
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

// team
type IdacTeamRankingResponse struct {
	Records []TeamRecord `json:"records"`
}
type TeamRecord struct {
	Rank           string `json:"rank"`
	TeamName       string `json:"team_name"`
	ShopName       string `json:"shopname"`
	Country        int    `json:"country"`
	AceUserName    string `json:"ace_user_name"`
	LeaderUserName string `json:"leader_user_name"`
	Point          string `json:"point"`
	LeagueEmoji    string `json:"league_emoji"`
}

// OB ranking
type IdacOBRankingResponse struct {
	CalcDate string            `json:"calcDate"`
	Records  []OBRankingRecord `json:"records"`
}

type OBRankingRecord struct {
	Rank       string `json:"rank"`
	Name       string `json:"name"`
	ShopName   string `json:"shopname"`
	UpdateDate string `json:"updateDate"`
	CarName    string `json:"carname"`
	Point      string `json:"point"`
	//MytitleId          string `json:"mytitleId"`
	PrideId            string `json:"prideId"`
	PridePoint         int    `json:"pridePoint"`
	OnlineBattleRankId string `json:"onlineBattleRankId"`
	StarCnt            int    `json:"starCnt"`
}

// Player ranking
type IdacPlayerRankingResponse struct {
	CalcDate string                `json:"calcDate"`
	Records  []PlayerRankingRecord `json:"records"`
}

type PlayerRankingRecord struct {
	ID         int    `json:"id"`
	Rank       int    `json:"rank"`
	Name       string `json:"name"`
	ShopName   string `json:"shopname"`
	GradeID    string `json:"gradeId"`
	GradeExp   int    `json:"gradeExp"`
	MytitleID  string `json:"mytitleId"`
	UpdateDate string `json:"updateDate"`
	NumberIcon string `json:"numberIcon"`
}

type PlayerTournamentInfo struct {
	Name string `json:"name"`
	// account grade fields
	Grade    string `json:"grade"`
	GradeNum string `json:"gradeNum"`
	GradeExp int    `json:"gradeExp"`
	// online battle fields
	OBRank    string `json:"obRank"`
	OBRankNum string `json:"obRankNum"`
	OBRankExp int    `json:"obRankExp"`
}

func (r OBRankingRecord) GetDisplayStarCount() string {
	return strconv.Itoa(r.StarCnt) + " (★)"
}

// --- ALIAS CONFIGURATION ---
var CourseAliases = map[string]string{
	// --- Akina Lake (0/2) ---
	"秋名湖／左周り":        "course-0",
	"akina lake ccw": "course-0",
	"lake ccw":       "course-0",
	"秋名湖／右周り":        "course-2",
	"akina lake cw":  "course-2",
	"lake cw":        "course-2",

	// --- Hakone (52/54) ---
	"箱根／下り":     "course-52",
	"hakone dh": "course-52",
	"箱根／上り":     "course-54",
	"hakone uh": "course-54",
	"hakone hc": "course-54",

	// --- Usui (36/38) ---
	"碓氷／左周り":   "course-36",
	"usui ccw": "course-36",
	"碓氷／右周り":   "course-38",
	"usui cw":  "course-38",

	// --- Myogi (4/6) ---
	"妙義／下り":    "course-4",
	"myogi dh": "course-4",
	"妙義／上り":    "course-6",
	"myogi uh": "course-6",
	"myogi hc": "course-6",

	// --- Akagi (8/10) ---
	"赤城／下り":    "course-8",
	"akagi dh": "course-8",
	"赤城／上り":    "course-10",
	"akagi uh": "course-10",
	"akagi hc": "course-10",

	// --- Akina (12/14) ---
	"秋名／下り":    "course-12",
	"akina dh": "course-12",
	"秋名／上り":    "course-14",
	"akina uh": "course-14",
	"akina hc": "course-14",

	// --- Irohazaka (16/18) ---
	"いろは坂／下り":      "course-16",
	"irohazaka dh": "course-16",
	"iro dh":       "course-16",
	"いろは坂／逆走":      "course-18",
	"irohazaka uh": "course-18", // Reverse often acts as Uphill/HC
	"irohazaka hc": "course-18",
	"iro uh":       "course-18",
	"iro hc":       "course-18",

	// --- Yabitsu (76/78) ---
	"ヤビツ／下り":     "course-76",
	"yabitsu dh": "course-76",
	"yabi dh":    "course-76",
	"ヤビツ／上り":     "course-78",
	"yabitsu uh": "course-78",
	"yabitsu hc": "course-78",
	"yabi uh":    "course-78",

	// --- Momiji Line (56/58) ---
	"もみじライン／下り": "course-56",
	"momiji dh": "course-56",
	"もみじライン／上り": "course-58",
	"momiji uh": "course-58",
	"momiji hc": "course-58",

	// --- Tsukuba (20/22) ---
	"筑波／往路":      "course-20",
	"tsukuba ob": "course-20",
	"筑波／復路":      "course-22",
	"tsukuba ib": "course-22",

	// --- Happogahara (24/26) ---
	"八方ヶ原／往路":        "course-24",
	"happogahara ob": "course-24",
	"happo ob":       "course-24",
	"八方ヶ原／復路":        "course-26",
	"happogahara ib": "course-26",
	"happo ib":       "course-26",

	// --- Sadamine (40/42) ---
	"定峰／下り":       "course-40",
	"sadamine dh": "course-40",
	"sada dh":     "course-40",
	"定峰／上り":       "course-42",
	"sadamine uh": "course-42",
	"sada uh":     "course-42",

	// --- Tsuchisaka (44/46) ---
	"土坂／往路":         "course-44",
	"tsuchisaka ob": "course-44",
	"tsuchi ob":     "course-44",
	"土坂／復路":         "course-46",
	"tsuchisaka ib": "course-46",
	"tsuchi ib":     "course-46",

	// --- Nagao (28/30) ---
	"長尾／下り":    "course-28",
	"nagao dh": "course-28",
	"長尾／上り":    "course-30",
	"nagao uh": "course-30",
	"nagao hc": "course-30",

	// --- Nanamagari (60/62) ---
	"七曲り／下り":        "course-60",
	"nanamagari dh": "course-60",
	"nana dh":       "course-60",
	"七曲り／上り":        "course-62",
	"nanamagari uh": "course-62",
	"nana uh":       "course-62",

	// --- Tsubaki Line (32/34) ---
	"椿ライン／下り":    "course-32",
	"tsubaki dh": "course-32",
	"椿ライン／上り":    "course-34",
	"tsubaki uh": "course-34",
	"tsubaki hc": "course-34",

	// --- Akina Snow (48/50) ---
	"秋名（雪）／下り":      "course-48",
	"akina snow dh": "course-48",
	"秋名（雪）／上り":      "course-50",
	"akina snow uh": "course-50",

	// --- Gunsai (64/66) ---
	"群サイ／往路":    "course-64",
	"gunsai ob": "course-64",
	"群サイ／復路":    "course-66",
	"gunsai ib": "course-66",

	// --- Odawara (68/70) ---
	"小田原／順走":      "course-68",
	"odawara":     "course-68",
	"小田原／逆走":      "course-70",
	"odawara rev": "course-70",

	// --- Tsukuba Snow (72/74) ---
	"筑波（雪）／往路":        "course-72",
	"tsukuba snow ob": "course-72",
	"筑波（雪）／復路":        "course-74",
	"tsukuba snow ib": "course-74",

	// --- Tsuchisaka Snow (80/82) ---
	"土坂（雪）／往路":           "course-80",
	"tsuchisaka snow ob": "course-80",
	"土坂（雪）／復路":           "course-82",
	"tsuchisaka snow ib": "course-82",

	// --- Manazuru (84/86) ---
	"真鶴／順走":        "course-84",
	"manazuru":     "course-84",
	"mana":         "course-84",
	"真鶴／逆走":        "course-86",
	"manazuru rev": "course-86",
	"mana rev":     "course-86",

	// --- Usui Snow (88/90) ---
	"碓氷（雪）／左周り":     "course-88",
	"usui snow ccw": "course-88",
	"碓氷（雪）／右周り":     "course-90",
	"usui snow cw":  "course-90",

	// --- Akina Rain (92/94) ---
	"秋名（雨）／下り":      "course-92",
	"akina rain dh": "course-92",
	"秋名（雨）／上り":      "course-94",
	"akina rain uh": "course-94",
}

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
	"course-96": "Irohazaka Rain (DH)",
	"course-98": "Irohazaka Rain (UC)",
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
	"area-all": "All",
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

// Add this to internal/domain/idac.go

var CarAliases = map[string]string{
	"all": "car-all",
	// TOYOTA
	"ae86": "car-0", "trueno": "car-0", "86": "car-0",
	"ae86 levin": "car-1", "levin": "car-1",
	"ae85": "car-2", "levin sr": "car-2",
	"zn6": "car-7", "86 gt": "car-7",
	"sw20": "car-3", "mr2": "car-3",
	"zzw30": "car-5", "mrs": "car-5",
	"sxe10": "car-4", "altezza": "car-4",
	"st205": "car-10", "celica": "car-10",
	"jza80": "car-6", "supra": "car-6",
	"gxpa16": "car-11", "gr yaris": "car-11", "yaris": "car-11",
	"db42": "car-12", "gr supra": "car-12",
	"jzx100": "car-13", "chaser": "car-13",
	"zn8": "car-14", "gr86": "car-14",
	"mxwh61": "car-15", "prius": "car-15",

	// NISSAN
	"bnr32": "car-256", "r32": "car-256", "gtr32": "car-256",
	"bnr34": "car-257", "r34": "car-257", "gtr34": "car-257",
	"s13": "car-258", "silvia s13": "car-258",
	"s14": "car-259", "silvia s14": "car-259",
	"s15": "car-260", "silvia s15": "car-260",
	"rps13": "car-261", "180sx": "car-261",
	"z33": "car-262", "350z": "car-262",
	"r35 nismo": "car-263", "gtr35 nismo": "car-263",
	"r35 tspec": "car-267", "gtr35 my25": "car-267",
	"er34": "car-264", "skyline 25gt": "car-264",
	"s30": "car-265", "fairlady s30": "car-265", "devil z": "car-265",
	"rz34": "car-266", "new z": "car-266",

	// HONDA
	"eg6": "car-512", "civic eg6": "car-512",
	"ek9": "car-513", "civic ek9": "car-513",
	"dc2": "car-514", "integra": "car-514",
	"ap1": "car-515", "s2000": "car-515",
	"na1": "car-516", "nsx": "car-516",
	"fl5": "car-517", "civic fl5": "car-517",

	// MAZDA
	"fc3s": "car-768", "fc": "car-768", "rx7 fc": "car-768", // fc
	"rx7 efini": "car-769", "fd efini": "car-769", // fd efini
	"fd rs": "car-773", "rx7 rs": "car-773", // fd rs

	"se3p": "car-770", "rx8": "car-770",
	"na6ce": "car-771", "roadster na": "car-771", "mx5": "car-771",
	"nb8c": "car-772", "roadster nb": "car-772",

	// SUBARU
	"gc8": "car-1024", "impreza gc8": "car-1024",
	"gdbf": "car-1025", "impreza gdbf": "car-1025",
	"gdba": "car-1026", "impreza gdba": "car-1026",
	"zc6": "car-1027", "brz": "car-1027",
	"vab": "car-1028", "wrx sti": "car-1028",
	"zd8": "car-1029", "brz s": "car-1029",

	// MITSUBISHI
	"ce9a": "car-1280", "evo 3": "car-1280", "lan evo 3": "car-1280",
	"cn9a": "car-1281", "evo 4": "car-1281",
	"ct9a 9": "car-1282", "evo 9": "car-1282",
	"ct9a 7": "car-1283", "evo 7": "car-1283",
	"cz4a": "car-1284", "evo 10": "car-1284",
	"cp9a 5": "car-1285", "evo 5": "car-1285",
	"cp9a 6": "car-1286", "evo 6": "car-1286",

	// SUZUKI
	"ea11r": "car-1536", "cappuccino": "car-1536", "cappu": "car-1536",
	"zc33s": "car-1537", "swift": "car-1537",

	// OTHERS
	"sileighty": "car-1792", "sil80": "car-1792",
	"964": "car-2304", "porsche 964": "car-2304", "blackbird": "car-2304",
	"982": "car-2305", "cayman": "car-2305",
	"991": "car-2306", "gt3": "car-2306",
}

// CarDisplayNameByCode maps the internal IDAC car codes to their full display names
var CarDisplayNameByCode = map[string]string{
	"car-all": "All Cars",
	"":        "All Cars",
	// TOYOTA
	"car-0":  "SPRINTER TRUENO GT-APEX (AE86)",
	"car-9":  "SPRINTER TRUENO 2door GT-APEX (AE86)",
	"car-1":  "COROLLA LEVIN GT-APEX (AE86)",
	"car-2":  "COROLLA LEVIN SR (AE85)",
	"car-7":  "86 GT (ZN6)",
	"car-3":  "MR2 G-Limited (SW20)",
	"car-5":  "MR-S S EDITION (ZZW30)",
	"car-4":  "ALTEZZA RS200 Z EDITION (SXE10)",
	"car-10": "CELICA GT-FOUR (ST205)",
	"car-6":  "SUPRA RZ (JZA80)",
	"car-11": "GR YARIS 1st Edition RZ “High performance” (GXPA16)",
	"car-12": "GR SUPRA RZ (DB42)",
	"car-13": "CHASER 2.5 TourerV (JZX100)",
	"car-14": "GR86 RZ (ZN8)",
	"car-15": "PRIUS Z (MXWH61)",

	// NISSAN
	"car-256": "SKYLINE GT-R V･specⅡ (BNR32)",
	"car-257": "SKYLINE GT-R V･specⅡ Nür (BNR34)",
	"car-258": "SILVIA K's (S13)",
	"car-259": "Silvia Q's (S14)",
	"car-260": "Silvia spec-R (S15)",
	"car-261": "180SX TYPE Ⅱ (RPS13)",
	"car-262": "FAIRLADY Z Version S (Z33)",
	"car-263": "NISSAN GT-R NISMO (R35)",
	"car-264": "SKYLINE 25GT TURBO (ER34)",
	"car-265": "Fairlady Z (S30)",
	"car-266": "FAIRLADY Z Version ST (RZ34)",
	"car-267": "NISSAN GT-R Premium edition T-spec (R35)",

	// HONDA
	"car-512": "Civic SiR･Ⅱ (EG6)",
	"car-513": "CIVIC TYPE R (EK9)",
	"car-514": "INTEGRA TYPE R (DC2)",
	"car-515": "S2000 (AP1)",
	"car-516": "NSX (NA1)",
	"car-517": "CIVIC TYPE R (FL5)",

	// MAZDA
	"car-768": "SAVANNA RX-7 ∞Ⅲ (FC3S)",
	"car-769": "ε~fini RX-7 Type R (FD3S)",
	"car-770": "RX-8 Type S (SE3P)",
	"car-771": "EUNOS ROADSTER (NA6CE)",
	"car-772": "ROADSTER RS (NB8C)",
	"car-773": "RX-7 Type RS (FD3S)",

	// SUBARU
	"car-1024": "IMPREZA WRX type R STi Version Ⅴ (GC8)",
	"car-1025": "IMPREZA WRX STI (GDBF)",
	"car-1026": "IMPREZA WRX STi (GDBA)",
	"car-1027": "SUBARU BRZ S (ZC6)",
	"car-1028": "STI S207 NBR CHALLENGE PACKAGE (VAB)",
	"car-1029": "BRZ S (ZD8)",

	// MITSUBISHI
	"car-1280": "LANCER GSR Evolution Ⅲ (CE9A)",
	"car-1281": "LANCER RS EVOLUTION Ⅳ (CN9A)",
	"car-1282": "LANCER Evolution Ⅸ GSR (CT9A)",
	"car-1283": "LANCER EVOLUTION Ⅶ GSR (CT9A)",
	"car-1284": "LANCER EVOLUTION Ⅹ GSR (CZ4A)",
	"car-1285": "LANCER RS EVOLUTION Ⅴ (CP9A)",
	"car-1286": "LANCER GSR EVOLUTION Ⅵ T.M.EDITION (CP9A)",

	// SUZUKI
	"car-1536": "Cappuccino (EA11R)",
	"car-1537": "SWIFT Sport (ZC33S)",
	"car-1538": "CARRY KC (DC51T)",

	// OTHERS
	"car-1792": "SILEIGHTY",
	"car-2304": "911Turbo3.6 (964)",
	"car-2305": "718Cayman GTS (982)",
	"car-2306": "911 GT3 (991)",
	"car-2560": "4C (96018)",
}

type TrackVariant struct {
	Name string
	ID   string
}

// TrackRegistry maps a Track Name to its available Variants
var TrackRegistry = map[string][]TrackVariant{
	"Akina Lake「秋名湖」": {
		{"『左周り』Counter-Clockwise (CCW)", "course-0"},
		{"『右周り』Clockwise (CW)", "course-2"},
	},
	"Hakone「箱根」": {
		{"『下り』Downhill", "course-52"},
		{"『上り』Uphill", "course-54"},
	},
	"Usui「碓氷」": {
		{"『左周り』Counter-Clockwise (CCW)", "course-36"},
		{"『右周り』Clockwise (CW)", "course-38"},
	},
	"Myogi「妙義」": {
		{"『下り』Downhill", "course-4"},
		{"『上り』Uphill", "course-6"},
	},
	"Akagi「赤城」": {
		{"『下り』Downhill", "course-8"},
		{"『上り』Uphill", "course-10"},
	},
	"Akina「秋名」": {
		{"『下り』Downhill", "course-12"},
		{"『上り』Uphill", "course-14"},
	},
	"Irohazaka「いろは坂」": {
		{"『下り』Downhill", "course-16"},
		{"『逆走』Uphill", "course-18"},
	},
	"Tsukuba「筑波」": {
		{"『往路』Outbound", "course-20"},
		{"『復路』Inbound", "course-22"},
	},
	"Happogahara「八方ヶ原」": {
		{"『往路』Outbound", "course-24"},
		{"『復路』Inbound", "course-26"},
	},
	"Nagao「長岡」": {
		{"『下り』Downhill", "course-28"},
		{"『上り』Uphill", "course-30"},
	},
	"Tsubaki Line「椿ライン」": {
		{"『下り』Downhill", "course-32"},
		{"『上り』Uphill", "course-34"},
	},
	"Sadamine「定峰」": {
		{"『下り』Downhill", "course-40"},
		{"『上り』Uphill", "course-42"},
	},
	"Tsuchisaka「土坂」": {
		{"『往路』Outbound", "course-44"},
		{"『復路』Inbound", "course-46"},
	},
	"Akina Snow「秋名 (雪)」": {
		{"『下り』Downhill", "course-48"},
		{"『上り』Uphill", "course-50"},
	},
	"Momiji Line「もみじライン」": {
		{"『下り』Downhill", "course-56"},
		{"『上り』Uphill", "course-58"},
	},
	"Nanamagari「七曲り」": {
		{"『下り』Downhill", "course-60"},
		{"『上り』Uphill", "course-62"},
	},
	"Gunsai「群サイ」": {
		{"『往路』Outbound", "course-64"},
		{"『復路』Inbound", "course-66"},
	},
	"Odawara「小田原」": {
		{"『往路』Outbound", "course-68"},
		{"『復路』Inbound", "course-70"},
	},
	"Tsukuba Snow「筑波 (雪)」": {
		{"『往路』Outbound", "course-72"},
		{"『復路』Inbound", "course-74"},
	},
	"Yabitsu「ヤビツ」": {
		{"『下り』Downhill", "course-76"},
		{"『上り』Uphill", "course-78"},
	},
	"Tsuchisaka (Snow)「土坂 (雪)」": {
		{"『往路』Outbound", "course-80"},
		{"『復路』Inbound", "course-82"},
	},
	"Manazuru「真鶴」": {
		{"『順走』Forward", "course-84"},
		{"『逆走』Reverse", "course-86"},
	},
	"Usui (Snow)「碓氷 (雪)」": {
		{"『左周り』Counter-Clockwise (CCW)", "course-88"},
		{"『右周り』Clockwise (CW)", "course-90"},
	},
	"Akina (Rain)「秋名 (雨)」": {
		{"『下り』Downhill", "course-92"},
		{"『上り』Uphill", "course-94"},
	},
	"Irohazaka (Rain)「いろは坂（雨）」": {
		{"『下り』Downhill", "course-96"},
		{"『逆走』Uphill", "course-98"},
	},
}

var MergedTrackRegistry = map[string]string{
	"Akina Lake / Counter-Clockwise (CCW) |『秋名湖 / 左周り』":     "course-0",
	"Akina Lake / Clockwise (CW) |『秋名湖 / 右周り』":              "course-2",
	"Hakone / Downhill |『箱根 / 下り』":                          "course-52",
	"Hakone / Uphill |『箱根 / 上り』":                            "course-54",
	"Usui / Counter-Clockwise (CCW) |『碓氷 / 左周り』":            "course-36",
	"Usui / Clockwise (CW) |『碓氷 / 右周り』":                     "course-38",
	"Myogi / Downhill |『妙義 / 下り』":                           "course-4",
	"Myogi / Uphill |『妙義 / 上り』":                             "course-6",
	"Akagi / Downhill |『赤城 / 下り』":                           "course-8",
	"Akagi / Uphill |『赤城 / 上り』":                             "course-10",
	"Akina / Downhill |『秋名 / 下り』":                           "course-12",
	"Akina / Uphill |『秋名 / 上り』":                             "course-14",
	"Irohazaka / Downhill |『いろは坂 / 下り』":                     "course-16",
	"Irohazaka / Uphill |『いろは坂 / 逆走』":                       "course-18",
	"Tsukuba / Outbound |『筑波 / 往路』":                         "course-20",
	"Tsukuba / Inbound |『筑波 / 復路』":                          "course-22",
	"Happogahara / Outbound |『八方ヶ原 / 往路』":                   "course-24",
	"Happogahara / Inbound |『八方ヶ原 / 復路』":                    "course-26",
	"Nagao / Downhill |『長岡 / 下り』":                           "course-28",
	"Nagao / Uphill |『長岡 / 上り』":                             "course-30",
	"Tsubaki Line / Downhill |『椿ライン / 下り』":                  "course-32",
	"Tsubaki Line / Uphill |『椿ライン / 上り』":                    "course-34",
	"Sadamine / Downhill |『定峰 / 下り』":                        "course-40",
	"Sadamine / Uphill |『定峰 / 上り』":                          "course-42",
	"Tsuchisaka / Outbound |『土坂 / 往路』":                      "course-44",
	"Tsuchisaka / Inbound |『土坂 / 復路』":                       "course-46",
	"Akina Snow / Downhill |『秋名 (雪) / 下り』":                  "course-48",
	"Akina Snow / Uphill |『秋名 (雪) / 上り』":                    "course-50",
	"Momiji Line / Downhill |『もみじライン / 下り』":                 "course-56",
	"Momiji Line / Uphill |『もみじライン / 上り』":                   "course-58",
	"Nanamagari / Downhill |『七曲り / 下り』":                     "course-60",
	"Nanamagari / Uphill |『七曲り / 上り』":                       "course-62",
	"Gunsai / Outbound |『群サイ / 往路』":                         "course-64",
	"Gunsai / Inbound |『群サイ / 復路』":                          "course-66",
	"Odawara / Outbound |『小田原 / 往路』":                        "course-68",
	"Odawara / Inbound |『小田原 / 復路』":                         "course-70",
	"Tsukuba Snow / Outbound |『筑波 (雪) / 往路』":                "course-72",
	"Tsukuba Snow / Inbound |『筑波 (雪) / 復路』":                 "course-74",
	"Yabitsu / Downhill |『ヤビツ / 下り』":                        "course-76",
	"Yabitsu / Uphill |『ヤビツ / 上り』":                          "course-78",
	"Tsuchisaka (Snow) / Outbound |『土坂 (雪) / 往路』":           "course-80",
	"Tsuchisaka (Snow) / Inbound |『土坂 (雪) / 復路』":            "course-82",
	"Manazuru / Forward |『真鶴 / 順走』":                         "course-84",
	"Manazuru / Reverse |『真鶴 / 逆走』":                         "course-86",
	"Usui (Snow) / Counter-Clockwise (CCW) |『碓氷 (雪) / 左周り』": "course-88",
	"Usui (Snow) / Clockwise (CW) |『碓氷 (雪) / 右周り』":          "course-90",
	"Akina (Rain) / Downhill |『秋名 (雨) / 下り』":                "course-92",
	"Akina (Rain) / Uphill |『秋名 (雨) / 上り』":                  "course-94",
	"Irohazaka (Rain) / Downhill |『いろは坂（雨） / 下り』":           "course-96",
	"Irohazaka (Rain) / Uphill |『いろは坂（雨） / 逆走』":             "course-98",
}

// GetTrackNames Helper to get keys for the Dropdown
//func GetTrackNames() []Merge {
//	keys := make([]string, 0, len(MergedTrackRegistry))
//	for k := range MergedTrackRegistry {
//		keys = append(keys, k)
//	}
//	sort.Strings(keys) // Ensure consistent order
//	return keys
//}

// ResolveCarID calculates the final car ID based on the model alias and spec string.
// Logic: FinalID = BaseID + (SpecIndex * 65536)
func ResolveCarID(alias string, specInput string) string {
	// 1. Normalize input
	lowerAlias := strings.ToLower(alias)

	// 2. Resolve Alias to Base ID (e.g. "r34" -> "car-257")
	baseID, ok := CarAliases[lowerAlias]
	if !ok {
		// Fallback: assume input is already a raw code like "car-257"
		baseID = lowerAlias
	}

	// 3. Resolve Spec String to Index
	specIndex := 0
	if specInput != "" {
		lowerSpec := strings.ToLower(specInput)
		if idx, ok := SpecAliases[lowerSpec]; ok {
			specIndex = idx
		} else {
			// Fallback: Check if user typed a number directly (e.g. "1")
			if val, err := strconv.Atoi(specInput); err == nil {
				specIndex = val
			}
		}
	}

	// 4. If no spec requested (index 0), or if input is "car-all", return base
	if specIndex <= 0 || baseID == "car-all" {
		return baseID
	}

	// 5. Parse the numeric ID (remove "car-" prefix)
	rawIDStr := strings.TrimPrefix(baseID, "car-")
	rawID, err := strconv.Atoi(rawIDStr)
	if err != nil {
		return baseID
	}

	// 6. Apply Bit Shift Logic for Specs
	// Spec 1 = +65536, Spec 2 = +131072, etc.
	finalID := rawID + (specIndex * 65536)

	return fmt.Sprintf("car-%d", finalID)
}

// SpecAliases maps user-friendly spec names to their internal index (bitshift multiplier)
var SpecAliases = map[string]int{
	"dh": 0, "c": 0,
	"ar": 1, "a": 1, // Spec 1
	"hc": 2, "b": 2, // Spec 2

}

// SpecEmojis maps spec codes to emojis (Unicode or Custom Discord IDs)
var SpecEmojis = map[string]string{
	"dh":    "<:dh:1448368217574740038>",
	"ar":    "<:ar:1448368501550092339>",
	"hc":    "<:hc:1448368366732447908>",
	"speed": "<:speed:1473575226418790551>",
	"tech":  "<:tech:1473575264171851817>",
}

var TeamLeagueEmojis = map[string]string{
	"6": "<:league_06_large:1449571849104130232>",
	"5": "<:league_05_large:1449571870285238294>",
	"4": "<:league_04_large:1449571893215756461>",
	"3": "<:league_03_large:1449571909493850233>",
}

// Helper to map Country IDs to Flags

func GetCountryFlag(id int) string {
	if id >= 0 && id <= 46 {
		return "🇯🇵" // Japan Prefectures
	}
	switch id {
	case 47:
		return "🇲🇴"
	case 48:
		return "🇭🇰"
	case 49:
		return "🇰🇷"
	case 50:
		return "🇲🇾"
	case 51:
		return "🇸🇬"
	case 52:
		return "🇹🇼"
	case 53:
		return "🇮🇩"
	case 54:
		return "🇵🇭"
	case 55:
		return "🇹🇭"
	case 56:
		return "🇺🇸"
	case 57:
		return "🇻🇳"
	case 58:
		return "🇲🇲"
	case 59:
		return "🇦🇺"
	case 60:
		return "🇳🇿"
	case 61:
		return "🇰🇭"
	}
	return "🏳️"
}

// Helper: Parse "3'14"727" into milliseconds (int)
func ParseIdacTime(timeStr string) (int, error) {
	var m, s, ms int
	// Standard format: M'SS"mmm
	_, err := fmt.Sscanf(timeStr, "%d'%d\"%d", &m, &s, &ms)
	if err != nil {
		return 0, err
	}
	return (m * 60000) + (s * 1000) + ms, nil
}

// Helper: Format milliseconds back to "+M'SS"mmm" string
func FormatIdacTimeDelta(diff int) string {
	sign := "+"
	if diff < 0 {
		sign = "-"
		diff = -diff
	}
	m := diff / 60000
	rem := diff % 60000
	s := rem / 1000
	ms := rem % 1000
	return fmt.Sprintf("%s%d'%02d\"%03d", sign, m, s, ms)
}

// NormalizeTextWidth converts Full-width (Zenkaku) characters to Half-width (ASCII).
// E.g., "ＳＨＩＲＯ～" -> "SHIRO~"
// E.g., "ＵｗＵ" -> "UwU"
func NormalizeTextWidth(s string) string {
	return strings.Map(func(r rune) rune {
		// 1. Handle Full-width Space
		if r == '\u3000' {
			return ' '
		}
		// 2. Handle Full-width ASCII variants (！ to ～)
		// Range: U+FF01 to U+FF5E
		if r >= 0xFF01 && r <= 0xFF5E {
			return r - 0xFEE0
		}
		// Return the character unchanged if it's not full-width
		return r
	}, s)
}

type IdacConstResponse struct {
	Cars   []CarMetadata      `json:"car"`
	Styles []CarStyleMetadata `json:"style"`
}

// CarSpecInfo holds the canonical model code and base spec for lookups
type CarSpecInfo struct {
	Maker         string
	CarName       string
	ModelCode     string // Canonical model code (e.g. "CZ4A")
	BaseSpec      string // "speed" or "tech"
	SpecStyleName string
	SegaSpecID    string
	Aliases       []string
}

// AreaSyncInfo represents an area that is part of a synchronization group
type AreaSyncInfo struct {
	AreaCode string
	AreaName string
	Timezone string
}

type CarMetadata struct {
	ID                  string  `json:"id"` // pk
	SegaCarID           int64   `json:"car_id"`
	Name                string  `json:"car_name"`
	ModelCode           string  `json:"model_code"`
	Maker               string  `json:"maker_name"`
	BaseStyleName       string  `json:"base_style_name"`
	NormalizedBaseStyle string  `json:"normalized_base_style"`
	CarStyleIDs         []int64 `json:"style_car_id"`
	// Aggregated fields
	SpecIDs   []string `json:"spec_ids"`
	SpecNames []string `json:"spec_names"`
	// for searching
	SearchBlob string `json:"-"`
}

func (c *CarMetadata) GetNormalizedBaseStyle() string {
	switch c.BaseStyleName {
	case "テクニカル":
		return "tech"
	case "高速":
		return "speed"
	default:
		return c.BaseStyleName
	}
}
func (c *CarMetadata) GetNormalizedCarName() string {
	nameWithoutSpec := strings.Split(c.Name, "(")[0]
	return strings.TrimSpace(nameWithoutSpec)
}

func (c *CarMetadata) GetNormalizedMakerName() string {
	return strings.ToLower(c.Maker)
}

// GetCarModelCode extracts the text inside the last set of parentheses from a display name.
// e.g. "SPRINTER TRUENO GT-APEX (AE86)" -> "AE86"
func (c *CarMetadata) GetCarModelCode() string {
	start := strings.LastIndex(c.Name, "(")
	end := strings.LastIndex(c.Name, ")")
	if start == -1 || end == -1 || end <= start {
		return "" // No valid brackets found
	}
	return c.Name[start+1 : end]
}

type CarStyleMetadata struct {
	ID             string `json:"id"` // pk
	StyleCarID     int64  `json:"style_car_id"`
	CarID          string `json:"car_id"` // fk of cars
	Name           string `json:"style_name"`
	RouteStyleName string `json:"route_style_name"`
}

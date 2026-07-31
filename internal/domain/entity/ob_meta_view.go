package entity

// OBMetaCarStat is one car's usage count within a style group.
type OBMetaCarStat struct {
	Name  string
	Count int
}

// OBMetaStyleGroup groups cars by driving style (e.g. DH, HC), sorted by count
// descending.
type OBMetaStyleGroup struct {
	Style string
	Cars  []OBMetaCarStat
}

// OBMetaSpecGroup groups styles under a base spec, with styles sorted for stable
// output.
type OBMetaSpecGroup struct {
	BaseSpec string
	Styles   []OBMetaStyleGroup
}

// OBMetaView is the aggregated "most used cars in Online Battle" result: base
// spec -> style -> cars, already counted and sorted, plus the sampling context.
type OBMetaView struct {
	Round        int
	TotalSampled int
	CalcDate     string
	Specs        []OBMetaSpecGroup
}
